// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package neutrino

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test 1: file does not exist
	exists, err := fileExists(filepath.Join(tmpDir, "nonexistent.db"))
	require.NoError(t, err)
	assert.False(t, exists, "file should not exist")

	// Test 2: file exists
	testFile := filepath.Join(tmpDir, "test.db")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	exists, err = fileExists(testFile)
	require.NoError(t, err)
	assert.True(t, exists, "file should exist")

	// Test 3: directory exists (should return true)
	exists, err = fileExists(tmpDir)
	require.NoError(t, err)
	assert.True(t, exists, "directory should exist")
}

func TestGetPrivKey(t *testing.T) {
	// Test with valid private key
	privKeyHex := "0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"

	cryptoDriver, privKey, err := getPrivKey(secp256k1.Name, privKeyHex)
	require.NoError(t, err)
	require.NotNil(t, cryptoDriver)
	require.NotNil(t, privKey)

	// Test error cases (previously panics)
	_, _, err = getPrivKey(secp256k1.Name, "")
	assert.Error(t, err, "should return error with empty privKey")

	_, _, err = getPrivKey(secp256k1.Name, "invalid_hex")
	assert.Error(t, err, "should return error with invalid hex")

	_, _, err = getPrivKey("invalid_crypto", privKeyHex)
	assert.Error(t, err, "should return error with invalid crypto driver")
}

func TestEstimateBtcFee(t *testing.T) {
	// Create a simple tx with 1 input and 2 outputs
	tx := wire.NewMsgTx(2)
	tx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{Index: 0},
	})
	tx.AddTxOut(&wire.TxOut{Value: 1000})
	tx.AddTxOut(&wire.TxOut{Value: 2000})

	feeRate := btcutil.Amount(10) // 10 sat/byte
	fee := estimateBtcFee(tx, feeRate)

	// fee = (tx.SerializeSize() + len(tx.TxIn)*108) * feeRate
	expectedSize := tx.SerializeSize() + len(tx.TxIn)*108
	expectedFee := btcutil.Amount(expectedSize) * feeRate

	assert.Equal(t, int64(expectedFee), int64(fee))
	assert.Greater(t, int64(fee), int64(0))
}

func TestWaitUntilDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := &neutrinoClient{ctx: ctx}

	// Test 1: done immediately
	callCount := 0
	done := func() bool {
		callCount++
		return true
	}
	n.waitUntilDone("test1", done, 0)
	assert.Equal(t, 1, callCount)

	// Test 2: done after retries
	callCount = 0
	doneAfter3 := func() bool {
		callCount++
		return callCount >= 3
	}
	n.waitUntilDone("test2", doneAfter3, time.Millisecond*10)
	assert.Equal(t, 3, callCount)

	// Test 3: context cancelled
	ctx2, cancel2 := context.WithCancel(context.Background())
	n2 := &neutrinoClient{ctx: ctx2}
	cancel2() // cancel immediately

	neverDone := func() bool { return false }
	// Should return immediately due to context cancelled
	n2.waitUntilDone("test3", neverDone, time.Millisecond*10)
}

func TestGetCommitKey(t *testing.T) {
	n := &neutrinoClient{}
	// Initially nil
	assert.Nil(t, n.getCommitKey())

	// Set commit key
	testKey := "0xcc38546e9e659d15e6b4893f0ab32a06d103931a8230b0bde71459d2b27d6944"
	_, n.commitKey, _ = getPrivKey(secp256k1.Name, testKey)

	key := n.getCommitKey()
	assert.NotNil(t, key)
	assert.Equal(t, n.commitKey, key)
}
