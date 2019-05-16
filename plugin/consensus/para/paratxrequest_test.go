// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log"
	"github.com/stretchr/testify/mock"

	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	typesmocks "github.com/33cn/chain33/types/mocks"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	//types.Init("user.p.para.", nil)
	log.SetLogLevel("error")
}

func getPrivKey(t *testing.T) crypto.PrivKey {
	pk, err := common.FromHex("6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b")
	assert.Nil(t, err)

	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.Nil(t, err)

	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)
	return priKey
}

func TestCalcCommitMsgTxs(t *testing.T) {
	priKey := getPrivKey(t)
	client := commitMsgClient{
		privateKey: priKey,
	}
	nt1 := &pt.ParacrossNodeStatus{
		Height: 1,
		Title:  "user.p.para",
	}
	nt2 := &pt.ParacrossNodeStatus{
		Height: 2,
		Title:  "user.p.para",
	}
	notify := []*pt.ParacrossNodeStatus{nt1}
	tx, count, err := client.calcCommitMsgTxs(notify)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)
	assert.NotNil(t, tx)

	notify = append(notify, nt2)
	tx, count, err = client.calcCommitMsgTxs(notify)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), count)
	assert.NotNil(t, tx)

	tx, err = client.singleCalcTx(nt2)
	assert.Nil(t, err)
	assert.NotNil(t, tx)

}

func TestGetConsensusStatus(t *testing.T) {
	para := new(client)
	grpcClient := &typesmocks.Chain33Client{}
	//grpcClient.On("GetFork", mock.Anything, &types.ReqKey{Key: []byte("ForkBlockHash")}).Return(&types.Int64{Data: 1}, errors.New("err")).Once()
	para.grpcClient = grpcClient
	commitCli := new(commitMsgClient)
	commitCli.paraClient = para

	block := &types.Block{
		Height:     1,
		MainHeight: 10,
	}

	status := &pt.ParacrossStatus{
		Height: 1,
	}
	reply := &types.Reply{
		IsOk: true,
		Msg:  types.Encode(status),
	}
	grpcClient.On("QueryChain", mock.Anything, mock.Anything).Return(reply, nil).Once()
	ret, err := commitCli.getConsensusStatus(block)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), ret.Height)
}

func TestSendCommitMsg(t *testing.T) {
	para := new(client)
	grpcClient := &typesmocks.Chain33Client{}
	//grpcClient.On("GetFork", mock.Anything, &types.ReqKey{Key: []byte("ForkBlockHash")}).Return(&types.Int64{Data: 1}, errors.New("err")).Once()
	para.grpcClient = grpcClient
	commitCli := new(commitMsgClient)
	commitCli.paraClient = para
	commitCli.quit = make(chan struct{})

	commitCli.paraClient.wg.Add(1)
	sendMsgCh := make(chan *types.Transaction, 1)
	go commitCli.sendCommitMsg(sendMsgCh)

	//reply := &types.Reply{
	//	IsOk: true,
	//	Msg:  types.Encode(status),
	//}
	grpcClient.On("SendTransaction", mock.Anything, mock.Anything).Return(nil, types.ErrNotFound).Twice()
	tx := &types.Transaction{}

	sendMsgCh <- tx
	time.Sleep(3 * time.Second)

	//para.BaseClient.Close()
	close(commitCli.quit)

}
