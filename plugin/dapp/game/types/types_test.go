/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package types

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

func TestCreateRawGamePreCreateTx(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	param := &GamePreCreateTx{
		Amount:   100,
		HashType: "SHA256",
	}
	tx, err := CreateRawGamePreCreateTx(cfg, param)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, []byte(GameX), tx.Execer)
	assert.NotEqual(t, 0, tx.Fee)
}

func TestCreateRawGamePreMatchTx(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	param := &GamePreMatchTx{
		GameID: "xxxxxxxxxxxxxxxxxxxxxxxxxx",
		Guess:  1,
	}
	tx, err := CreateRawGamePreMatchTx(cfg, param)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, []byte(GameX), tx.Execer)
	assert.NotEqual(t, 0, tx.Fee)
}

func TestCreateRawGamePreCloseTx(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	param := &GamePreCloseTx{
		GameID: "xxxxxxxxxxxxxxxxxxxxxxxxxx",
		Result: 1,
	}
	tx, err := CreateRawGamePreCloseTx(cfg, param)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, []byte(GameX), tx.Execer)
	assert.NotEqual(t, 0, tx.Fee)
}

func TestCreateRawGamePreCancelTx(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	param := &GamePreCancelTx{
		GameID: "xxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	tx, err := CreateRawGamePreCancelTx(cfg, param)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, []byte(GameX), tx.Execer)
	assert.NotEqual(t, 0, tx.Fee)
}
