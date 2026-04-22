// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package neutrino

import (
	"sync"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/lightninglabs/neutrino/headerfs"
	"github.com/stretchr/testify/assert"
)

// TestBestBlockGetterSetter tests the concurrent-safe best block access
func TestBestBlockGetterSetter(t *testing.T) {
	n := &neutrinoClient{}

	// Initially nil
	assert.Nil(t, n.getBestBlock())

	// Set a valid block
	blk := &headerfs.BlockStamp{
		Height: 100,
		Hash:   [32]byte{1, 2, 3, 4},
	}
	n.setBestBlock(blk)

	// Verify we can read it back
	result := n.getBestBlock()
	assert.NotNil(t, result)
	assert.Equal(t, int32(100), result.Height)
	assert.Equal(t, chainhash.Hash([32]byte{1, 2, 3, 4}), result.Hash)

	// Test setting nil doesn't overwrite existing value
	n.setBestBlock(nil)
	result = n.getBestBlock()
	assert.NotNil(t, result) // Should still be the previous value
	assert.Equal(t, int32(100), result.Height)
}

// TestBestBlockConcurrentAccess tests thread safety
func TestBestBlockConcurrentAccess(t *testing.T) {
	n := &neutrinoClient{}

	var wg sync.WaitGroup
	numGoroutines := 100
	numIterations := 100

	// Concurrent writes
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				blk := &headerfs.BlockStamp{
					Height: int32(idx*numIterations + j),
				}
				n.setBestBlock(blk)
			}
		}(i)
	}

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numIterations; j++ {
				_ = n.getBestBlock()
			}
		}()
	}

	wg.Wait()

	// Final value should be valid (not panic or corrupt)
	result := n.getBestBlock()
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.Height, int32(0))
}
