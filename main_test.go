package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateWallet(t *testing.T) {
	w := generateWallet()
	require.Equal(t, w.Address[:6], "gonka1", "Incorrect bech32 prefix")
	require.Equal(t, len(w.Address), 44, "Incorrect address length")
	require.Equal(t, len(w.Pubkey), 33, "Incorrect pubkey length")
	require.Equal(t, len(w.Privkey), 32, "Incorrect privkey length")
}

func TestPrefix(t *testing.T) {
	m := matcher{Prefix: "aaaa"}
	require.True(t, m.Match("gonka1aaaaqztg6eu45nlljp0wp947juded46aln83kr"))
	require.False(t, m.Match("gonka1aaa9qztg6eu45nlljp0wp947juded46aln83kr"))
}

func TestSuffix(t *testing.T) {
	m := matcher{Suffix: "8888"}
	require.True(t, m.Match("gonka14sy657pp6tgclhgqnl3dkwzwu3ewt4cf3f8888"))
	require.False(t, m.Match("gonka14sy657pp6tgclhgqnl3dkwzwu3ewt4cf3ff888"))
}

func TestContains(t *testing.T) {
	m := matcher{Contains: "k2k2k"}
	require.True(t, m.Match("gonka1s6rlmknaj3swdd7hua6s852sk2k2k409a3z9f9"))
	require.False(t, m.Match("gonka14sy657pp6tgclhgqnl3dkwzwu3ewt4cf3ff888"))
}

func TestLetters(t *testing.T) {
	m := matcher{Letters: 38}
	require.True(t, m.Match("gonka1gcjsgsglhacarlumkjzywedykkvkuvrzqlnlxd"))
	require.False(t, m.Match("gonka1gcjsgsglhacarlumkjzywedykkvkuvrzqlnlx8"))
}

func TestDigits(t *testing.T) {
	m := matcher{Digits: 26}
	require.True(t, m.Match("gonka1j666m3qz66t786s48t540536465p56zrve5893"))
	require.False(t, m.Match("gonka1j666m3qz66t786s48t540536465p56zrve589z"))
}

func TestFindMatchingWalletConcurrent(t *testing.T) {
	goroutineCount := 5
	lastChars := "zz"
	m := matcher{Suffix: lastChars}
	w := findMatchingWalletConcurrent(m, goroutineCount)
	require.Equal(t, w.Address[len(w.Address)-len(lastChars):], lastChars, "Incorrect address suffix")
}
