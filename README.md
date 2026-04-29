# gonka-vanity

CLI tool for generating Gonka AI vanity wallet addresses locally.

## Important

**RU**
- Это **НЕ** онлайн-генератор кошельков.
- Это локальная утилита: ключи генерируются только на вашем устройстве.
- Не публикуйте `privkey`, mnemonic phrases и файлы результатов.

**EN**
- This is **NOT** an online wallet generator.
- This is a local tool: keys are generated only on your machine.
- Never publish `privkey`, mnemonic phrases, or result files.

## Features

- Generate Gonka AI bech32 vanity addresses with the `gonka` prefix.
- Match address payloads by prefix, suffix, substring, repeated characters, letters, or digits.
- Use multiple CPU cores.
- Show live progress with attempts, speed, elapsed time, expected attempts, and ETA.
- Print full wallet records to stdout or save them to a TXT file with `--out`.
- Optional BIP39 mnemonic mode with configurable word count and derivation path.

## Installing

### Option 1: run from source

Install Go from https://go.dev/dl/, then check it:

```powershell
go version
```

Clone the repository:

```powershell
git clone https://github.com/andrey055/gonka-vanity.git
cd gonka-vanity
```

Download dependencies and run:

```powershell
go mod tidy
go run . --repeat 5 -n 1 --out results.txt
```

### Option 2: install with Go

```powershell
go install github.com/andrey055/gonka-vanity@latest
gonka-vanity --repeat 5 -n 1 --out results.txt
```

### Option 3: download a release

Download a prebuilt binary from the [Releases](https://github.com/andrey055/gonka-vanity/releases) page, unpack it, and run the `gonka-vanity` executable from your terminal.

## Usage Examples

Find an address whose payload starts with `aaaa`:

```bash
./gonka-vanity --prefix aaaa
```

Find an address whose payload ends with `8888`:

```bash
./gonka-vanity --suffix 8888
```

Find an address containing substring `k2k2k` anywhere:

```bash
./gonka-vanity --contains k2k2k
```

Find an address containing a repeated run of 5 identical chars:

```bash
./gonka-vanity --repeat 5
```

Find an address whose payload starts with 7 identical chars:

```bash
./gonka-vanity --repeat-prefix 7
```

Find an address whose payload ends with 7 identical chars:

```bash
./gonka-vanity --repeat-suffix 7
```

Find an address that matches either a repeated prefix or a repeated suffix:

```bash
./gonka-vanity --repeat-prefix 7 --repeat-suffix 7 -n 10
```

Find an address with at least 26 digits:

```bash
./gonka-vanity --digits 26
```

Find an address with at least 30 letters:

```bash
./gonka-vanity --letters 30
```

Generate 5 results:

```bash
./gonka-vanity -n 5 --repeat 5
```

Restrict CPU threads:

```bash
./gonka-vanity --cpus 1 --repeat 5
```

Save full wallet records to a local TXT file:

```bash
./gonka-vanity --repeat 5 -n 3 --out results.txt
```

Generate mnemonic-based wallets:

```bash
./gonka-vanity --mnemonic --repeat 5 -n 1 --out results.txt
```

Customize mnemonic length and derivation path:

```bash
./gonka-vanity --mnemonic --words 12 --path "m/44'/1200'/0'/0/0" --repeat 5 -n 1 --out results.txt
```

Show all CLI flags:

```bash
./gonka-vanity --help
```

## Output Behavior

Without `--out`, stdout prints the full wallet record, including `privkey`:

```text
address=gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr pubkey=02b6e4b7e8d2b1f0d6d9c0d3c58a9a7bd0123456789abcdef0123456789abcd privkey=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef matched=prefix,repeat attempts=123456 elapsed=00:00:02
```

With `--out`, stdout prints only the address, while the selected file stores the full wallet record:

```text
gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr
```

Output file line:

```text
address=gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr pubkey=02b6e4b7e8d2b1f0d6d9c0d3c58a9a7bd0123456789abcdef0123456789abcd privkey=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef matched=prefix,repeat attempts=123456 elapsed=00:00:02
```

Mnemonic mode also includes `mnemonic` and `path`:

```text
address=gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr pubkey=02b6e4b7e8d2b1f0d6d9c0d3c58a9a7bd0123456789abcdef0123456789abcd privkey=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef mnemonic="word1 word2 ... word24" path="m/44'/1200'/0'/0/0" matched=repeat attempts=123456 elapsed=00:00:02
```

All private keys and mnemonic phrases in this README are fake examples.

## Security Notes

- Treat every generated `privkey`, mnemonic phrase, and output file as a secret.
- Prefer `--out results.txt` for long runs so terminal line wrapping does not break copied private keys.
- Keep result files offline or in a trusted encrypted location.
- Do not commit `results.txt`, `results2.txt`, screenshots, terminal logs, or copied output containing private keys.

## Development

Run tests:

```bash
go test
```

Run linters through pre-commit:

```bash
pre-commit run -a
```

