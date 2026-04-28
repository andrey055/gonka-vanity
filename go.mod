module github.com/andrey055/gonka-vanity

go 1.17

require (
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/cosmos/cosmos-sdk v0.45.1
	github.com/cosmos/go-bip39 v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.35.1
)

require (
	github.com/cosmos/btcutil v1.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/petermattis/goid v0.0.0-20180202154549-b0b1615b78e5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sasha-s/go-deadlock v0.2.1-0.20190427202633-1595213edefa // indirect
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// see https://github.com/cosmos/cosmos-sdk/issues/8469
replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

// NOTE:
// - Go may try to resolve Tendermint as "@latest" while running `go mod tidy`.
// - cosmos-sdk v0.45.x expects Tendermint packages that are present in v0.35.1.
// - This replace forces a compatible Tendermint version and prevents `go mod tidy` from
//   selecting a newer v0.35.x that removed some packages.
replace github.com/tendermint/tendermint => github.com/tendermint/tendermint v0.35.1
