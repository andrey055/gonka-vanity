Самый простой безопасный запуск (Windows, PowerShell)

1) Установить Go (один раз)
Пользователь ставит Go с https://go.dev/dl/ и проверяет:

```powershell
go version
```

2) Склонировать репозиторий

```powershell
git clone https://github.com/andrey055/gonka-vanity.git
cd gonka-vanity
```

3) Подтянуть зависимости и запустить

```powershell
go mod tidy
go run . --repeat 5 -n 1 --out results.txt
```

Как сделать ещё проще, но без бинарников

Вариант A: go install (очень удобно)
Тогда человек делает:

```powershell
go install github.com/andrey055/gonka-vanity@latest
gonka-vanity --repeat 5 -n 1 --out results.txt
```
Vanity wallet generator for Gonka AI addresses.

[![Build Status](https://github.com/andrey055/gonka-vanity/workflows/Tests/badge.svg?branch=main)](https://github.com/andrey055/gonka-vanity/actions?query=workflow%3ATests+branch%3Amain+event%3Apush)

# gonka-vanity

<!--- Don't edit the version line below manually. Let bump2version do it for you. -->
> Version 1.0.0

> CLI tool for generating Gonka AI vanity wallet addresses

## Важно / Important

**RU**
- Это **НЕ** онлайн-генератор кошельков.
- Это **локальный офлайн** vanity address search tool.
- **Никаких API вызовов, телеметрии, облачной синхронизации**.
- Ключи генерируются **только локально** на вашем устройстве.

**EN**
- This is **NOT** an online wallet generator.
- This is a **local offline** vanity address search tool.
- **No API calls, no telemetry, no cloud sync**.
- Keys are generated **locally only** on your machine.

## Features

- Generate Gonka AI bech32 vanity addresses with the `gonka` prefix
- Use all CPU cores
- Match addresses by **OR** rules (any rule can match):
  - `--prefix` (payload starts with)
  - `--suffix` (payload ends with)
  - `--contains` (payload contains)
  - `--repeat` (payload has N repeated chars in a row)
  - `--letters` / `--digits` (payload has at least N letters/digits)
- Clean terminal progress UI (attempts / speed / elapsed)
- Save found wallets locally to TXT (`--out`)
- Optional mnemonic mode (`--mnemonic`) for wallet-friendly backups

## Current status

**RU**
- По умолчанию утилита генерирует **приватный ключ напрямую** (secp256k1) и выводит только адрес в stdout.
- Опционально можно включить **BIP39 mnemonic** режим (`--mnemonic`), чтобы сохранять сид-фразу вместе с найденным адресом.
- При использовании `--out` в файл сохраняются **приватные ключи** (и в mnemonic-режиме — **мнемоника**) — храните файл безопасно.

**EN**
- By default the tool generates a **raw secp256k1 private key** and prints only the address to stdout.
- You can enable optional **BIP39 mnemonic** mode (`--mnemonic`) to store a seed phrase for each found address.
- When using `--out`, the output file contains **private keys** (and in mnemonic mode — **mnemonic phrases**) — keep it secure.

## Installing

Download the latest binary release from the [_Releases_](https://github.com/andrey055/gonka-vanity/releases) page. Alternatively, build from source yourself.

## Usage examples

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

Find an address containing a repeated run of 5 identical chars (e.g. `aaaaa`, `77777`, `qqqqq`):

```bash
./gonka-vanity --repeat 5
```

Generate 5 results (default is 1):

```bash
./gonka-vanity -n 5 --repeat 5
```

Restrict CPU threads (defaults to all cores):

```bash
./gonka-vanity --cpus 1 --repeat 5
```

Save results to a local TXT file (one line per result; file contains private keys):

```bash
./gonka-vanity --repeat 5 -n 3 --out results.txt
```

Mnemonic mode (24 words by default, path compatible with `inferenced` coin type 1200):

```bash
./gonka-vanity --mnemonic --repeat 5 -n 1 --out results.txt
```

Customize mnemonic length and derivation path:

```bash
./gonka-vanity --mnemonic --words 12 --path "m/44'/1200'/0'/0/0" --repeat 5 -n 1 --out results.txt
```

## Output example (illustrative)

**NOTE:** The private key / mnemonic below are **fake** and shown for documentation purposes only.

**stdout (address-only):**

```text
gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr
```

**--out file (txt, one line per result):**

```text
address=gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr pubkey=02b6e4b7e8d2b1f0d6d9c0d3c58a9a7bd0123456789abcdef0123456789abcd privkey=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef matched=prefix,repeat attempts=123456 elapsed=00:00:02
```

**Mnemonic mode example line (txt):**

```text
address=gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr pubkey=02b6e4b7e8d2b1f0d6d9c0d3c58a9a7bd0123456789abcdef0123456789abcd privkey=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef mnemonic="word1 word2 ... word24" path="m/44'/1200'/0'/0/0" matched=repeat attempts=123456 elapsed=00:00:02
```

## Attribution

This project is based on [`hukkin/cosmosvanity`](https://github.com/hukkin/cosmosvanity), originally released under the MIT License.
Vanity wallet generator for Gonka AI addresses.

[![Build Status](https://github.com/andrey055/gonka-vanity/workflows/Tests/badge.svg?branch=main)](https://github.com/andrey055/gonka-vanity/actions?query=workflow%3ATests+branch%3Amain+event%3Apush)

# gonka-vanity

<!--- Don't edit the version line below manually. Let bump2version do it for you. -->
> Version 1.0.0

> CLI tool for generating Gonka AI vanity wallet addresses

## Важно / Important

**RU**
- Это **НЕ** онлайн-генератор кошельков.
- Это **локальный офлайн** vanity address search tool.
- **Никаких API вызовов, телеметрии, облачной синхронизации**.
- Ключи генерируются **только локально** на вашем устройстве.

**EN**
- This is **NOT** an online wallet generator.
- This is a **local offline** vanity address search tool.
- **No API calls, no telemetry, no cloud sync**.
- Keys are generated **locally only** on your machine.

## Features

- Generate Gonka AI bech32 vanity addresses with the `gonka` prefix
- Use all CPU cores
- Match addresses by **OR** rules (any rule can match):
  - `--prefix` (payload starts with)
  - `--suffix` (payload ends with)
  - `--contains` (payload contains)
  - `--repeat` (payload has N repeated chars in a row)
  - `--letters` / `--digits` (payload has at least N letters/digits)
- Clean terminal progress UI (attempts / speed / elapsed)
- Save found wallets locally to TXT (`--out`)

## Current status

**RU**
- Сейчас утилита генерирует **приватный ключ напрямую** (secp256k1) и выводит адрес в stdout.
- **Seed phrase (mnemonic) пока не генерируется**.
- При использовании `--out` в файл сохраняются **приватные ключи** — храните файл безопасно.

**EN**
- Currently the tool generates a **raw secp256k1 private key** and prints the address to stdout.
- **Seed phrase (mnemonic) is not generated yet**.
- When using `--out`, the output file contains **private keys** — keep it secure.

## Installing

Download the latest binary release from the [_Releases_](https://github.com/andrey055/gonka-vanity/releases) page. Alternatively, build from source yourself.

## Usage examples

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

Find an address containing a repeated run of 5 identical chars (e.g. `aaaaa`, `77777`, `qqqqq`):

```bash
./gonka-vanity --repeat 5
```

Generate 5 results (default is 1):

```bash
./gonka-vanity -n 5 --repeat 5
```

Restrict CPU threads (defaults to all cores):

```bash
./gonka-vanity --cpus 1 --repeat 5
```

Save results to a local TXT file (one line per result; file contains private keys):

```bash
./gonka-vanity --repeat 5 -n 3 --out results.txt
```

## Output example (illustrative)

**NOTE:** The private key below is **fake** and shown for documentation purposes only.

**stdout (address-only):**

```text
gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr
```

**--out file (txt, one line per result):**

```text
address=gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr pubkey=02b6e4b7e8d2b1f0d6d9c0d3c58a9a7bd0123456789abcdef0123456789abcd privkey=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef matched=prefix,repeat attempts=123456 elapsed=00:00:02
```

## Attribution

This project is based on [`hukkin/cosmosvanity`](https://github.com/hukkin/cosmosvanity), originally released under the MIT License.
Vanity wallet generator for Gonka AI addresses.

[![Build Status](https://github.com/andrey055/gonka-vanity/workflows/Tests/badge.svg?branch=main)](https://github.com/andrey055/gonka-vanity/actions?query=workflow%3ATests+branch%3Amain+event%3Apush)

# gonka-vanity

<!--- Don't edit the version line below manually. Let bump2version do it for you. -->
> Version 1.0.0

> CLI tool for generating Gonka AI vanity wallet addresses

## Важно / Important

**RU**
- Это **НЕ** онлайн-генератор кошельков.
- Это **локальный офлайн** vanity address search tool.
- **Никаких API вызовов, телеметрии, облачной синхронизации**.
- Ключи генерируются **только локально** на вашем устройстве.

**EN**
- This is **NOT** an online wallet generator.
- This is a **local offline** vanity address search tool.
- **No API calls, no telemetry, no cloud sync**.
- Keys are generated **locally only** on your machine.

## Features
* Generate Gonka AI bech32 vanity addresses with the `gonka` prefix
* Use all CPU cores
* Specify a substring that the addresses must
    * start with
    * end with
    * contain
* Set required minimum amount of letters (a-z) or digits (0-9) in the addresses
* Binaries built for Linux, macOS and Windows

## Current status (baseline)

**RU**
- Сейчас утилита генерирует **приватный ключ напрямую** (secp256k1) и показывает address/pubkey/privkey.
- **Seed phrase (mnemonic) пока не генерируется**. Это будет добавлено на следующем этапе.

**EN**
- Currently the tool generates a **raw secp256k1 private key** and prints address/pubkey/privkey.
- **Seed phrase (mnemonic) is not generated yet**. This will be added in a later stage.

## Security notes

**RU**
- Никогда не публикуйте и не отправляйте приватный ключ третьим лицам.
- Рекомендуется запускать генерацию на доверенном устройстве (при желании — офлайн).

**EN**
- Never share or publish private keys.
- Run generation on a trusted machine (optionally offline/air-gapped).

## Installing
Download the latest binary release from the [_Releases_](https://github.com/andrey055/gonka-vanity/releases) page. Alternatively, build from source yourself.

## Usage examples
Find an address that starts with "aaaa" (e.g. gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr)
```bash
./gonka-vanity --prefix aaaa
```

Find an address that ends with "8888" (e.g. gonka14sy657pp6tgclhgqnl3dkwzwu3ewt4cf3f8888)
```bash
./gonka-vanity --suffix 8888
```

Find an address containing the substring "k2k2k" (e.g. gonka1s6rlmknaj3swdd7hua6s852sk2k2k409a3z9f9)
```bash
./gonka-vanity --contains k2k2k
```

Find an address consisting of letters only (e.g. gonka1gcjsgsglhacarlumkjzywedykkvkuvrzqlnlxd)
```bash
./gonka-vanity --letters 38
```

Find an address with at least 26 digits (e.g. gonka1j666m3qz66t786s48t540536465p56zrve5893)
```bash
./gonka-vanity --digits 26
```

Generate 5 addresses (the default is 1)
```bash
./gonka-vanity -n 5
```

Restrict to using only 1 CPU thread. This value defaults to the number of CPUs available.
```bash
./gonka-vanity --cpus 1
```

Combine flags introduced above
```bash
./gonka-vanity --contains 8888 --prefix a --suffix c
```

## Output example (illustrative)

**NOTE:** The private key below is **fake** and shown for documentation purposes only.

**stdout (address-only):**

```text
:::: Matching wallet 1/1 found ::::
gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr

```

