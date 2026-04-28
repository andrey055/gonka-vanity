package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type matcher struct {
	Prefix   string
	Suffix   string
	Contains string
	Letters  int
	Digits   int
	Repeat   int
}

func (m matcher) anyRuleProvided() bool {
	return m.Prefix != "" || m.Suffix != "" || m.Contains != "" || m.Repeat > 0 || m.Digits > 0 || m.Letters > 0
}

func (m matcher) matchPayload(candidate string) string {
	// NOTE: We match only the bech32 payload (part after "gonka1") so that rules are stable
	// regardless of bech32 human-readable prefix.
	return strings.TrimPrefix(candidate, "gonka1")
}

func (m matcher) MatchDetailed(candidate string) (bool, []string) {
	payload := m.matchPayload(candidate)

	var matched []string

	// NOTE: OR logic. Any enabled matcher can satisfy the search.
	if m.Prefix != "" && strings.HasPrefix(payload, m.Prefix) {
		matched = append(matched, "prefix")
	}
	if m.Suffix != "" && strings.HasSuffix(payload, m.Suffix) {
		matched = append(matched, "suffix")
	}
	if m.Contains != "" && strings.Contains(payload, m.Contains) {
		matched = append(matched, "contains")
	}
	if m.Repeat > 0 && hasRepeatedRun(payload, m.Repeat) {
		matched = append(matched, "repeat")
	}
	if m.Digits > 0 && countUnionChars(payload, bech32digits) >= m.Digits {
		matched = append(matched, "digits")
	}
	if m.Letters > 0 && countUnionChars(payload, bech32letters) >= m.Letters {
		matched = append(matched, "letters")
	}

	return len(matched) > 0, matched
}

func (m matcher) Match(candidate string) bool {
	ok, _ := m.MatchDetailed(candidate)
	return ok
}

func (m matcher) ValidationErrors() []string {
	var errs []string
	if len(m.Contains) > 38 || len(m.Prefix) > 38 || len(m.Suffix) > 38 {
		errs = append(errs, "ERROR: A provided matcher is too long. Must be max 38 characters.")
	}
	if m.Digits < 0 || m.Letters < 0 {
		errs = append(errs, "ERROR: Can't require negative amount of characters")
	}
	if m.Digits+m.Letters > 38 {
		errs = append(errs, "ERROR: Can't require more than 38 characters")
	}
	if m.Repeat < 0 {
		errs = append(errs, "ERROR: --repeat can't be negative")
	}
	if m.Repeat == 1 {
		// A run length of 1 would match every address and is almost certainly a user mistake.
		errs = append(errs, "ERROR: --repeat must be 2 or more")
	}
	if !m.anyRuleProvided() {
		errs = append(errs, "ERROR: Please provide at least one matcher: --prefix/--suffix/--contains/--repeat/--digits/--letters")
	}
	if (!bech32Only(m.Contains)) || (!bech32Only(m.Prefix)) || (!bech32Only(m.Suffix)) {
		errs = append(errs, "ERROR: A provided matcher contains bech32 incompatible characters")
	}
	return errs
}

type wallet struct {
	Address string
	Pubkey  []byte
	Privkey []byte
}

func (w wallet) String() string {
	return "Address:\t" + w.Address + "\n" +
		"Public key:\t" + hex.EncodeToString(w.Pubkey) + "\n" +
		"Private key:\t" + hex.EncodeToString(w.Privkey)
}

func generateWallet() wallet {
	var privkey secp256k1.PrivKey = secp256k1.GenPrivKey()
	var pubkey secp256k1.PubKey = privkey.PubKey().(secp256k1.PubKey)
	bech32Addr, err := bech32.ConvertAndEncode("gonka", pubkey.Address())
	if err != nil {
		panic(err)
	}

	return wallet{bech32Addr, pubkey, privkey}
}

type matchResult struct {
	Wallet   wallet
	Matched  []string
	Attempts uint64
	Elapsed  time.Duration
}

func hasRepeatedRun(s string, runLen int) bool {
	// NOTE: One pass, no allocations. "runLen <= 1" is handled by validation and is kept
	// here only for defensive completeness.
	if runLen <= 1 {
		return true
	}

	var last rune
	cur := 0
	for _, r := range s {
		if r == last {
			cur++
		} else {
			last = r
			cur = 1
		}
		if cur >= runLen {
			return true
		}
	}
	return false
}

func formatDuration(d time.Duration) string {
	// NOTE: Stable HH:MM:SS keeps progress/file output easy to read and parse.
	totalSeconds := int(d.Seconds())
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	h := totalSeconds / 3600
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func writeMatchTXT(f *os.File, res matchResult) error {
	// NOTE: One line per result (grep-friendly).
	//
	// SECURITY WARNING:
	// - This tool persists private keys. Users MUST protect output files.
	_, err := fmt.Fprintf(
		f,
		"address=%s pubkey=%s privkey=%s matched=%s attempts=%d elapsed=%s\n",
		res.Wallet.Address,
		hex.EncodeToString(res.Wallet.Pubkey),
		hex.EncodeToString(res.Wallet.Privkey),
		strings.Join(res.Matched, ","),
		res.Attempts,
		formatDuration(res.Elapsed),
	)
	return err
}

func findMatchingWallets(ch chan matchResult, quit chan struct{}, m matcher, attempts *uint64, startedAt time.Time) {
	for {
		select {
		case <-quit:
			return
		default:
			w := generateWallet()
			curAttempts := atomic.AddUint64(attempts, 1)
			ok, matched := m.MatchDetailed(w.Address)
			if ok {
				res := matchResult{
					Wallet:   w,
					Matched:  matched,
					Attempts: curAttempts,
					Elapsed:  time.Since(startedAt),
				}
				// Do a non-blocking write instead of simple `ch <- w` to prevent
				// blocking when it's time to quit and ch is full.
				select {
				case ch <- res:
				default:
				}
			}
		}
	}
}

func findMatchingWalletConcurrent(m matcher, goroutines int, quit chan struct{}, attempts *uint64, startedAt time.Time) matchResult {
	ch := make(chan matchResult)

	for i := 0; i < goroutines; i++ {
		go findMatchingWallets(ch, quit, m, attempts, startedAt)
	}
	return <-ch
}

const bech32digits = "023456789"
const bech32letters = "acdefghjklmnpqrstuvwxyzACDEFGHJKLMNPQRSTUVWXYZ"

// This is alphanumeric chars minus chars "1", "b", "i", "o" (case insensitive)
const bech32chars = bech32digits + bech32letters

func bech32Only(s string) bool {
	return countUnionChars(s, bech32chars) == len(s)
}

func countUnionChars(s string, letterSet string) int {
	count := 0
	for _, char := range s {
		if strings.Contains(letterSet, string(char)) {
			count++
		}
	}
	return count
}

func main() {
	var walletsToFind = flag.IntP("count", "n", 1, "Amount of matching wallets to find")
	var cpuCount = flag.Int("cpus", runtime.NumCPU(), "Amount of CPU cores to use")

	var mustContain = flag.StringP("contains", "c", "", "Match addresses containing this substring")
	var mustStartWith = flag.StringP("prefix", "p", "", "Match addresses whose payload starts with this substring")
	var mustEndWith = flag.StringP("suffix", "s", "", "Match addresses whose payload ends with this substring")
	var letters = flag.IntP("letters", "l", 0, "Match addresses whose payload contains at least this many letters (a-z)")
	var digits = flag.IntP("digits", "d", 0, "Match addresses whose payload contains at least this many digits (0-9)")
	var repeat = flag.IntP("repeat", "r", 0, "Minimum length of a repeated-character run that the address must contain (e.g. 5 matches aaaaa / 77777 / qqqqq)")
	var outPath = flag.String("out", "", "Write found wallets to a local file (TXT)")
	var outFormat = flag.String("format", "txt", "Output format when using --out (txt)")
	flag.Parse()

	if *walletsToFind < 1 {
		fmt.Println("ERROR: The number of wallets to generate must be 1 or more")
		os.Exit(1)
	}
	if *cpuCount < 1 {
		fmt.Println("ERROR: Must use at least 1 CPU core")
		os.Exit(1)
	}

	m := matcher{
		Prefix:   strings.ToLower(*mustStartWith),
		Suffix:   strings.ToLower(*mustEndWith),
		Contains: strings.ToLower(*mustContain),
		Letters:  *letters,
		Digits:   *digits,
		Repeat:   *repeat,
	}
	matcherValidationErrs := m.ValidationErrors()
	if len(matcherValidationErrs) > 0 {
		for i := 0; i < len(matcherValidationErrs); i++ {
			fmt.Println(matcherValidationErrs[i])
		}
		os.Exit(1)
	}

	startedAt := time.Now()
	var attempts uint64

	var outFile *os.File
	if *outPath != "" {
		if strings.ToLower(*outFormat) != "txt" {
			fmt.Println("ERROR: Only --format=txt is supported")
			os.Exit(1)
		}
		f, err := os.OpenFile(*outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
		if err != nil {
			fmt.Println("ERROR: Failed to open output file:", err)
			os.Exit(1)
		}
		outFile = f
		defer func() { _ = outFile.Close() }()
	}

	quit := make(chan struct{})
	defer close(quit)

	// Progress loop: stderr only (keeps stdout clean/pipable).
	doneProgress := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		lastAttempts := uint64(0)
		lastAt := time.Now()
		for {
			select {
			case <-doneProgress:
				return
			case <-ticker.C:
				now := time.Now()
				cur := atomic.LoadUint64(&attempts)
				delta := cur - lastAttempts
				dt := now.Sub(lastAt).Seconds()
				speed := 0.0
				if dt > 0 {
					speed = float64(delta) / dt
				}

				fmt.Fprintf(os.Stderr, "\rAttempts: %d | Speed: %.0f/sec | Elapsed: %s", cur, speed, formatDuration(time.Since(startedAt)))
				lastAttempts = cur
				lastAt = now
			}
		}
	}()
	defer func() {
		close(doneProgress)
		fmt.Fprintln(os.Stderr)
	}()

	for i := 0; i < *walletsToFind; i++ {
		res := findMatchingWalletConcurrent(m, *cpuCount, quit, &attempts, startedAt)

		// Console output: address only.
		fmt.Println(res.Wallet.Address)

		// File output: full record with matched rule list.
		if outFile != nil {
			if err := writeMatchTXT(outFile, res); err != nil {
				fmt.Println("ERROR: Failed to write output file:", err)
				os.Exit(1)
			}
		}
	}
}
