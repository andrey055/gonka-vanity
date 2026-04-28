Vanity wallet generator for Gonka AI addresses.

[![Build Status](https://github.com/andrey055/gonka-vanity/workflows/Tests/badge.svg?branch=master)](https://github.com/andrey055/gonka-vanity/actions?query=workflow%3ATests+branch%3Amaster+event%3Apush)

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

```text
:::: Matching wallet 1/1 found ::::
Address:     gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr
Public key:  02b6e4b7e8d2b1f0d6d9c0d3c58a9a7bd0123456789abcdef0123456789abcd
Private key: 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
```

## Attribution
This project is based on [`hukkin/cosmosvanity`](https://github.com/hukkin/cosmosvanity), originally released under the MIT License.
