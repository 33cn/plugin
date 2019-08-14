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

	apimocks "github.com/33cn/chain33/client/mocks"
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
	tx, count, err := client.calcCommitMsgTxs(notify, 0)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), count)
	assert.NotNil(t, tx)

	notify = append(notify, nt2)
	tx, count, err = client.calcCommitMsgTxs(notify, 0)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), count)
	assert.NotNil(t, tx)

	tx, err = client.singleCalcTx(nt2, 0)
	assert.Nil(t, err)
	assert.NotNil(t, tx)

}

func TestGetConsensusStatus(t *testing.T) {
	mainFork := mainParaSelfConsensusForkHeight
	mainParaSelfConsensusForkHeight = 1
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
	getMockLastBlock(para, block)

	api := &apimocks.QueueProtocolAPI{}
	para.SetAPI(api)

	status := &pt.ParacrossStatus{
		Height: 1,
	}

	//msgx := &types.Message{types.Encode(status)}
	//msg := types.Encode(status)
	//reply := &types.Reply{
	//	IsOk: true,
	//	Msg:  types.Encode(status),
	//}
	api.On("QueryChain", mock.Anything, mock.Anything, mock.Anything).Return(status, nil).Once()
	detail := &types.BlockDetail{Block: block}
	details := &types.BlockDetails{Items: []*types.BlockDetail{detail}}

	api.On("GetBlocks", mock.Anything).Return(details, nil).Once()
	ret, err := commitCli.getSelfConsensusStatus()

	assert.Nil(t, err)
	assert.Equal(t, int64(1), ret.Height)
	mainParaSelfConsensusForkHeight = mainFork
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
	commitCli.sendMsgCh = make(chan *types.Transaction, 1)
	go commitCli.sendCommitMsg()

	//reply := &types.Reply{
	//	IsOk: true,
	//	Msg:  types.Encode(status),
	//}
	grpcClient.On("SendTransaction", mock.Anything, mock.Anything).Return(nil, types.ErrNotFound).Twice()
	tx := &types.Transaction{}

	commitCli.sendMsgCh <- tx
	time.Sleep(3 * time.Second)

	//para.BaseClient.Close()
	close(commitCli.quit)

}

func TestVerifyMainBlocks(t *testing.T) {
	hash0 := []byte("0")
	hash1 := []byte("1")
	hash2 := []byte("2")
	hash3 := []byte("3")
	//hash4 := []byte("4")
	//hash5 := []byte("5")
	hash6 := []byte("6")

	header1 := &types.Header{
		ParentHash: hash0,
		Hash:       hash1,
	}
	block1 := &types.ParaTxDetail{
		Type:   addAct,
		Header: header1,
	}

	header2 := &types.Header{
		ParentHash: hash1,
		Hash:       hash2,
	}
	block2 := &types.ParaTxDetail{
		Type:   addAct,
		Header: header2,
	}

	header3 := &types.Header{
		ParentHash: hash2,
		Hash:       hash3,
	}
	block3 := &types.ParaTxDetail{
		Type:   addAct,
		Header: header3,
	}

	//del3
	header4 := &types.Header{
		ParentHash: hash2,
		Hash:       hash3,
	}
	block4 := &types.ParaTxDetail{
		Type:   delAct,
		Header: header4,
	}
	//del2
	header5 := &types.Header{
		ParentHash: hash1,
		Hash:       hash2,
	}
	block5 := &types.ParaTxDetail{
		Type:   delAct,
		Header: header5,
	}

	header6 := &types.Header{
		ParentHash: hash1,
		Hash:       hash6,
	}
	block6 := &types.ParaTxDetail{
		Type:   addAct,
		Header: header6,
	}

	mainBlocks := &types.ParaTxDetails{
		Items: []*types.ParaTxDetail{block1, block2, block3, block4, block5, block6},
	}

	err := verifyMainBlocks(hash0, mainBlocks)
	assert.Equal(t, nil, err)
}
