package legacymimc

import (
	"encoding/hex"
	"testing"
)

// TestLegacyMiMCHashCompat 验证 legacymimc 与 gnark-crypto v0.5.3 的 MiMC 哈希字节级一致。
// 期望值由 gnark-crypto v0.5.3 的 fr/mimc 独立生成（seed 为 mix MimcHashSeed）。
func TestLegacyMiMCHashCompat(t *testing.T) {
	seed := "19172955941344617222923168298456110557655645809646772800021167670156933290312"
	cases := []struct {
		input    []byte
		expected string // gnark-crypto v0.5.3 输出
	}{
		{[]byte("hello legacy mimc"), "2645479389b840df42f30c64309d576df3c26ea258f0c80c1f36f7d325203e4a"},
		{[]byte{0x01, 0x02, 0x03, 0x04}, "1fb87d5df883ba21ada64ceb1d0cbfaa3484dd53a0f8ef04cbe3719f13f5589a"},
		{[]byte(seed), "08e8bee2bf41f3db60636c6a1d1b15a5809cf34953db0e09dfe665d100cbbc7a"},
	}
	for _, c := range cases {
		h := NewMiMC(seed)
		h.Write(c.input)
		got := hex.EncodeToString(h.Sum(nil))
		if got != c.expected {
			t.Fatalf("mimc(%q) = %s, want %s (v0.5.3)", c.input, got, c.expected)
		}
	}
}
