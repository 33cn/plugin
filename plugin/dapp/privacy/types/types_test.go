// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/stretchr/testify/assert"
)

func TestReplyPrivacyAccount(t *testing.T) {
	reply := &ReplyPrivacyAccount{}
	reply.Displaymode = 1
	reply.Utxos = &UTXOs{}
	reply.Ftxos = &UTXOs{}
	for n := 0; n < 10; n++ {
		utxo := &UTXO{}
		utxo.Amount = 1000000
		utxo.UtxoBasic = &UTXOBasic{
			OnetimePubkey: common.Hex2Bytes("123fds"),
			UtxoGlobalIndex: &UTXOGlobalIndex{
				Outindex: 1,
				Txhash:   common.Hex2Bytes("0x2c4aa7aea82de4a971bceb6cfef3d09dbeac7c7df3a4b49b5a311d23d772f027"),
			},
		}
		reply.Utxos.Utxos = append(reply.Utxos.Utxos, utxo)
		reply.Ftxos.Utxos = append(reply.Ftxos.Utxos, utxo)
	}
	_, err := types.PBToJSON(reply)
	assert.NoError(t, err)
}
