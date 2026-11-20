// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	ctypes "github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

// TestForkEVMCompatForkRegisteredOffByDefault pins the code-side default: the fork is
// registered at MaxHeight, so a chain that does not configure it never activates it.
func TestForkEVMCompatForkRegisteredOffByDefault(t *testing.T) {
	cfg := newTestConfig(t)
	require.Equal(t, int64(ctypes.MaxHeight), cfg.GetDappFork(evmtypes.ExecutorName, evmtypes.ForkEVMCompat))
	require.False(t, cfg.IsDappFork(1_000_000, evmtypes.ExecutorName, evmtypes.ForkEVMCompat))
}

// TestForkEVMCompatDeclaredInConfigs asserts that every tracked chain config carries
// ForkEVMCompat inside its [fork.sub.evm] section. initForkConfig panics on a missing
// key, so a config released without it cannot start a node.
func TestForkEVMCompatDeclaredInConfigs(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	for _, rel := range []string{
		"chain33.para.toml",
		"chain33.fork.toml",
		filepath.Join("plugin", "dapp", "evm", "cmd", "ci2", "chain33.proxyminer.toml"),
	} {
		t.Run(rel, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(repoRoot, rel))
			require.NoError(t, err)
			text := string(data)
			require.Contains(t, text, "[fork.sub.evm]", "%s has no evm fork section", rel)

			raw, ok := sectionValue(text, "[fork.sub.evm]", evmtypes.ForkEVMCompat)
			require.Truef(t, ok, "%s: [fork.sub.evm] is missing %s (initForkConfig panics)",
				rel, evmtypes.ForkEVMCompat)
			val, parseErr := strconv.ParseInt(raw, 10, 64)
			require.NoErrorf(t, parseErr, "%s: %s=%q is not an integer", rel, evmtypes.ForkEVMCompat, raw)
			// Repository configs ship the fork disabled (-1 is loaded as MaxHeight). A chain
			// that enables it sets a concrete height in its own deployment config.
			require.Equalf(t, int64(-1), val, "%s: %s=%d must stay disabled in the repository config",
				rel, evmtypes.ForkEVMCompat, val)
		})
	}
}

// sectionValue returns the raw value of key inside [section], ignoring the same key in
// other sections and commented-out entries.
func sectionValue(text, section, key string) (string, bool) {
	inSection := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inSection = line == section
			continue
		}
		if !inSection {
			continue
		}
		k, v, found := strings.Cut(line, "=")
		if !found || strings.TrimSpace(k) != key {
			continue
		}
		return strings.TrimSpace(v), true
	}
	return "", false
}
