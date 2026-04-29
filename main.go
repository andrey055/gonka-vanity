package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"sync/atomic"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

type matcher struct {
	Prefix   string
	Suffix   string
	Contains string
	Letters  int
	Digits   int
	Repeat   int
	RepeatPrefix int
	RepeatSuffix int
}

func (m matcher) anyRuleProvided() bool {
	return m.Prefix != "" || m.Suffix != "" || m.Contains != "" ||
		m.Repeat > 0 || m.RepeatPrefix > 0 || m.RepeatSuffix > 0 ||
		m.Digits > 0 || m.Letters > 0
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
	if m.RepeatPrefix > 0 && hasRepeatedPrefix(payload, m.RepeatPrefix) {
		matched = append(matched, "repeat-prefix")
	}
	if m.RepeatSuffix > 0 && hasRepeatedSuffix(payload, m.RepeatSuffix) {
		matched = append(matched, "repeat-suffix")
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
	if m.RepeatPrefix < 0 {
		errs = append(errs, "ERROR: --repeat-prefix can't be negative")
	}
	if m.RepeatPrefix == 1 {
		errs = append(errs, "ERROR: --repeat-prefix must be 2 or more")
	}
	if m.RepeatSuffix < 0 {
		errs = append(errs, "ERROR: --repeat-suffix can't be negative")
	}
	if m.RepeatSuffix == 1 {
		errs = append(errs, "ERROR: --repeat-suffix must be 2 or more")
	}
	if !m.anyRuleProvided() {
		errs = append(errs, "ERROR: Please provide at least one matcher: --prefix/--suffix/--contains/--repeat/--repeat-prefix/--repeat-suffix/--digits/--letters")
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
	// Mnemonic is optional. It is only set when the wallet was generated from a BIP39 seed phrase.
	Mnemonic string
	// DerivationPath is optional. It is only set when Mnemonic is set.
	DerivationPath string
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

	return wallet{Address: bech32Addr, Pubkey: pubkey, Privkey: privkey}
}

func generateMnemonic(words int) (string, error) {
	// NOTE: BIP39 word count maps to entropy size: 12 -> 128 bits, 24 -> 256 bits.
	entropyBits := 0
	switch words {
	case 12:
		entropyBits = 128
	case 24:
		entropyBits = 256
	default:
		return "", fmt.Errorf("unsupported mnemonic length: %d (use 12 or 24)", words)
	}

	entropy, err := bip39.NewEntropy(entropyBits)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

func deriveWalletFromMnemonic(mnemonic string, derivationPath string) (wallet, error) {
	// NOTE:
	// - We derive the private key using BIP32 (hdkeychain) instead of Cosmos SDK hd helpers.
	// - This avoids `go mod tidy` resolving dependency test packages that can conflict with
	//   Tendermint/Cosmos-SDK versioning in older SDK releases.
	// - Coin type should match Gonka's `inferenced` (1200), so default path is m/44'/1200'/0'/0/0.
	if !bip39.IsMnemonicValid(mnemonic) {
		return wallet{}, fmt.Errorf("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, "")
	derivedPrivKeyBytes, err := deriveBIP32PrivateKeyForPath(seed, derivationPath)
	if err != nil {
		return wallet{}, err
	}

	// Tendermint secp256k1 expects 32 bytes.
	if len(derivedPrivKeyBytes) != 32 {
		return wallet{}, fmt.Errorf("unexpected derived private key length: %d", len(derivedPrivKeyBytes))
	}

	priv := secp256k1.PrivKey(derivedPrivKeyBytes)
	pub := priv.PubKey().(secp256k1.PubKey)

	bech32Addr, err := bech32.ConvertAndEncode("gonka", pub.Address())
	if err != nil {
		return wallet{}, err
	}

	return wallet{
		Address:        bech32Addr,
		Pubkey:         pub,
		Privkey:        priv,
		Mnemonic:       mnemonic,
		DerivationPath: derivationPath,
	}, nil
}

func generateWalletMnemonic(words int, derivationPath string) (wallet, error) {
	mn, err := generateMnemonic(words)
	if err != nil {
		return wallet{}, err
	}
	return deriveWalletFromMnemonic(mn, derivationPath)
}

func deriveBIP32PrivateKeyForPath(seed []byte, derivationPath string) ([]byte, error) {
	// NOTE:
	// - `hdkeychain` is Bitcoin-oriented but implements standard BIP32 derivation which is
	//   also used for Cosmos-style wallets.
	// - We use chaincfg.MainNetParams only as a required parameter; it doesn't affect the
	//   underlying key derivation.
	//
	// Expected path format: m/44'/1200'/0'/0/0
	path := strings.TrimSpace(derivationPath)
	if path == "" {
		return nil, fmt.Errorf("empty derivation path")
	}
	if !strings.HasPrefix(path, "m/") {
		return nil, fmt.Errorf("unsupported derivation path (expected to start with \"m/\"): %q", path)
	}
	parts := strings.Split(path[2:], "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid derivation path: %q", path)
	}

	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	key := master
	for _, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("invalid derivation path segment in %q", path)
		}

		hardened := strings.HasSuffix(p, "'")
		if hardened {
			p = strings.TrimSuffix(p, "'")
		}

		var idx uint32
		_, err := fmt.Sscanf(p, "%d", &idx)
		if err != nil {
			return nil, fmt.Errorf("invalid derivation index %q in %q", p, path)
		}
		if hardened {
			idx = idx + hdkeychain.HardenedKeyStart
		}

		// NOTE: `hdkeychain` uses Derive() for child keys.
		key, err = key.Derive(idx)
		if err != nil {
			return nil, err
		}
	}

	ecPriv, err := key.ECPrivKey()
	if err != nil {
		return nil, err
	}

	// 32-byte serialized secret scalar.
	return ecPriv.Serialize(), nil
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

func hasRepeatedPrefix(s string, runLen int) bool {
	// NOTE: Checks if the string starts with runLen identical characters.
	if runLen <= 1 {
		return true
	}
	if len(s) == 0 {
		return false
	}

	// bech32 payload is ASCII; using bytes keeps this fast and simple.
	b := []byte(s)
	if len(b) < runLen {
		return false
	}
	first := b[0]
	for i := 1; i < runLen; i++ {
		if b[i] != first {
			return false
		}
	}
	return true
}

func hasRepeatedSuffix(s string, runLen int) bool {
	// NOTE: Checks if the string ends with runLen identical characters.
	if runLen <= 1 {
		return true
	}
	if len(s) == 0 {
		return false
	}

	b := []byte(s)
	if len(b) < runLen {
		return false
	}
	last := b[len(b)-1]
	for i := 2; i <= runLen; i++ {
		if b[len(b)-i] != last {
			return false
		}
	}
	return true
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

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func pow32(n int) float64 {
	// NOTE: 32 is the bech32 charset size (excluding the separator '1').
	// Using math.Pow is fine here; this is used only for ETA estimation, not the hot path.
	return math.Pow(32.0, float64(n))
}

func binomialTailProbability(n int, p float64, atLeast int) float64 {
	// NOTE:
	// - Computes P(X >= atLeast) for X ~ Binomial(n, p).
	// - n is small (38 for bech32 payload), so a simple stable iterative PMF is fast enough.
	if atLeast <= 0 {
		return 1.0
	}
	if atLeast > n {
		return 0.0
	}
	if p <= 0 {
		return 0.0
	}
	if p >= 1 {
		return 1.0
	}

	// Compute pmf(k) iteratively from pmf(0) to pmf(n).
	pmf := math.Pow(1.0-p, float64(n))
	tail := 0.0
	for k := 0; k <= n; k++ {
		if k >= atLeast {
			tail += pmf
		}
		if k == n {
			break
		}
		// pmf(k+1) = pmf(k) * ((n-k)/(k+1)) * (p/(1-p))
		pmf *= (float64(n-k) / float64(k+1)) * (p / (1.0 - p))
	}
	return clamp01(tail)
}

func estimateSuccessProbability(m matcher) float64 {
	// NOTE:
	// - This is an approximation intended for user-facing ETA only.
	// - We treat enabled matchers as OR and estimate each matcher's probability independently,
	//   then combine via: p = 1 - Π(1 - p_i).
	//
	// Assumptions:
	// - bech32 payload length for accounts is ~38 chars.
	// - charset size is 32.
	const payloadLen = 38

	var probs []float64

	if m.Prefix != "" {
		probs = append(probs, 1.0/pow32(len(m.Prefix)))
	}
	if m.Suffix != "" {
		probs = append(probs, 1.0/pow32(len(m.Suffix)))
	}
	if m.Contains != "" {
		k := len(m.Contains)
		if k > 0 && k <= payloadLen {
			positions := float64(payloadLen - k + 1)
			// Union bound approximation for "substring occurs anywhere".
			probs = append(probs, clamp01(positions/pow32(k)))
		}
	}
	if m.Repeat > 0 && m.Repeat <= payloadLen {
		positions := float64(payloadLen - m.Repeat + 1)
		// For a fixed start position, probability that next (N-1) chars equal the first is 1/32^(N-1).
		probs = append(probs, clamp01(positions/pow32(m.Repeat-1)))
	}
	if m.RepeatPrefix > 0 && m.RepeatPrefix <= payloadLen {
		// First N chars all equal: 1/32^(N-1)
		probs = append(probs, 1.0/pow32(m.RepeatPrefix-1))
	}
	if m.RepeatSuffix > 0 && m.RepeatSuffix <= payloadLen {
		// Last N chars all equal: 1/32^(N-1)
		probs = append(probs, 1.0/pow32(m.RepeatSuffix-1))
	}
	if m.Digits > 0 {
		// bech32digits = 9 out of 32 chars.
		probs = append(probs, binomialTailProbability(payloadLen, 9.0/32.0, m.Digits))
	}
	if m.Letters > 0 {
		// bech32letters = 23 out of 32 chars.
		probs = append(probs, binomialTailProbability(payloadLen, 23.0/32.0, m.Letters))
	}

	if len(probs) == 0 {
		return 0.0
	}

	none := 1.0
	for _, pi := range probs {
		none *= (1.0 - clamp01(pi))
	}
	return clamp01(1.0 - none)
}

func writeMatchTXT(w io.Writer, res matchResult) error {
	// NOTE: One line per result (grep-friendly).
	//
	// SECURITY WARNING:
	// - This tool exposes private keys. Users MUST protect output files and terminal logs.
	// NOTE: Keep fields stable and single-line. Mnemonic-related fields are only present
	// when mnemonic mode is enabled.
	if res.Wallet.Mnemonic != "" {
		_, err := fmt.Fprintf(
			w,
			"address=%s pubkey=%s privkey=%s mnemonic=%q path=%q matched=%s attempts=%d elapsed=%s\n",
			res.Wallet.Address,
			hex.EncodeToString(res.Wallet.Pubkey),
			hex.EncodeToString(res.Wallet.Privkey),
			res.Wallet.Mnemonic,
			res.Wallet.DerivationPath,
			strings.Join(res.Matched, ","),
			res.Attempts,
			formatDuration(res.Elapsed),
		)
		return err
	}

	_, err := fmt.Fprintf(
		w,
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

type walletGenerator func() (wallet, error)

func findMatchingWallets(ch chan matchResult, quit chan struct{}, m matcher, attempts *uint64, startedAt time.Time, gen walletGenerator) {
	for {
		select {
		case <-quit:
			return
		default:
			w, err := gen()
			if err != nil {
				// NOTE: Generator errors should be extremely rare. We skip and continue so the
				// search keeps running without crashing.
				continue
			}
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

func findMatchingWalletConcurrent(m matcher, goroutines int, quit chan struct{}, attempts *uint64, startedAt time.Time, gen walletGenerator) matchResult {
	ch := make(chan matchResult)

	for i := 0; i < goroutines; i++ {
		go findMatchingWallets(ch, quit, m, attempts, startedAt, gen)
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
	var repeatPrefix = flag.Int("repeat-prefix", 0, "Match addresses whose payload starts with N identical characters (e.g. 5 matches aaaaa / 77777 / qqqqq)")
	var repeatSuffix = flag.Int("repeat-suffix", 0, "Match addresses whose payload ends with N identical characters (e.g. 5 matches aaaaa / 77777 / qqqqq)")
	var outPath = flag.String("out", "", "Write full wallet records including private keys to a local TXT file; stdout prints addresses only when set")
	var outFormat = flag.String("format", "txt", "Output format for --out; only txt is currently supported")
	var useMnemonic = flag.Bool("mnemonic", false, "Generate wallets from BIP39 mnemonic (slower but wallet-friendly)")
	var mnemonicWords = flag.Int("words", 24, "Mnemonic word count when using --mnemonic (12 or 24)")
	var derivationPath = flag.String("path", "m/44'/1200'/0'/0/0", "HD derivation path when using --mnemonic")

	// Keep the built-in flag list, but make the private-key output behavior explicit.
	// This matters because users must know whether secrets are printed to the terminal
	// or persisted to a file before starting a long search.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Output behavior:")
		fmt.Fprintln(os.Stderr, "  Without --out: stdout prints full wallet records including private keys.")
		fmt.Fprintln(os.Stderr, "  With --out: full wallet records are written to the file, while stdout prints addresses only.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
	}
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
		RepeatPrefix: *repeatPrefix,
		RepeatSuffix: *repeatSuffix,
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

	// NOTE:
	// - Print a one-time approximation for expected attempts and ETA.
	// - This is best-effort and can be off significantly depending on rule interactions.
	pSuccess := estimateSuccessProbability(m)
	var expectedAttempts float64
	if pSuccess > 0 {
		expectedAttempts = 1.0 / pSuccess
	}
	if expectedAttempts > 0 {
		fmt.Fprintf(os.Stderr, "Estimated attempts per success: ~%.0f (p≈%.6g)\n", expectedAttempts, pSuccess)
	}

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
	var progressPrinted uint32
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

				etaStr := "--:--:--"
				if speed > 0 && expectedAttempts > 0 {
					remaining := expectedAttempts - float64(cur)
					if remaining < 0 {
						remaining = 0
					}
					etaStr = formatDuration(time.Duration(remaining/speed) * time.Second)
				}

				if expectedAttempts > 0 {
					fmt.Fprintf(
						os.Stderr,
						"\rAttempts: %d | Speed: %.0f/sec | Elapsed: %s | Expected: ~%.0f | ETA: %s",
						cur, speed, formatDuration(time.Since(startedAt)), expectedAttempts, etaStr,
					)
				} else {
					fmt.Fprintf(os.Stderr, "\rAttempts: %d | Speed: %.0f/sec | Elapsed: %s", cur, speed, formatDuration(time.Since(startedAt)))
				}
				atomic.StoreUint32(&progressPrinted, 1)
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
		gen := func() (wallet, error) {
			return generateWallet(), nil
		}
		if *useMnemonic {
			gen = func() (wallet, error) {
				return generateWalletMnemonic(*mnemonicWords, *derivationPath)
			}
		}

		res := findMatchingWalletConcurrent(m, *cpuCount, quit, &attempts, startedAt, gen)

		// Progress is drawn with carriage returns, so finish that line before stdout output.
		if atomic.LoadUint32(&progressPrinted) == 1 {
			fmt.Fprintln(os.Stderr)
			atomic.StoreUint32(&progressPrinted, 0)
		}

		if outFile != nil {
			// Console output stays address-only when full private key records are saved to a file.
			fmt.Println(res.Wallet.Address)

			// File output: full record with matched rule list.
			if err := writeMatchTXT(outFile, res); err != nil {
				fmt.Println("ERROR: Failed to write output file:", err)
				os.Exit(1)
			}
			continue
		}

		// Without --out, stdout must contain the full record so the private key is not lost.
		if err := writeMatchTXT(os.Stdout, res); err != nil {
			fmt.Println("ERROR: Failed to write wallet to stdout:", err)
			os.Exit(1)
		}
	}
}
