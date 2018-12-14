/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateRawGamePreCreateTx(t *testing.T) {
	param := &GamePreCreateTx{
		Amount:   100,
		HashType: "SHA256",
	}
	tx, err := CreateRawGamePreCreateTx(param)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, []byte(GameX), tx.Execer)
	assert.NotEqual(t, 0, tx.Fee)
}

func TestCreateRawGamePreMatchTx(t *testing.T) {
	param := &GamePreMatchTx{
		GameID: "xxxxxxxxxxxxxxxxxxxxxxxxxx",
		Guess:  1,
	}
	tx, err := CreateRawGamePreMatchTx(param)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, []byte(GameX), tx.Execer)
	assert.NotEqual(t, 0, tx.Fee)
}

func TestCreateRawGamePreCloseTx(t *testing.T) {
	param := &GamePreCloseTx{
		GameID: "xxxxxxxxxxxxxxxxxxxxxxxxxx",
		Result: 1,
	}
	tx, err := CreateRawGamePreCloseTx(param)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, []byte(GameX), tx.Execer)
	assert.NotEqual(t, 0, tx.Fee)
}

func TestCreateRawGamePreCancelTx(t *testing.T) {
	param := &GamePreCancelTx{
		GameID: "xxxxxxxxxxxxxxxxxxxxxxxxxx",
	}
	tx, err := CreateRawGamePreCancelTx(param)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, []byte(GameX), tx.Execer)
	assert.NotEqual(t, 0, tx.Fee)
}
