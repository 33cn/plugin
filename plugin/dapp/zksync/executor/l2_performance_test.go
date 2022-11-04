package executor

import (
	"fmt"
	"testing"

	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"

	chain33Common "github.com/33cn/chain33/common"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/assert"
)

func BenchmarkTransfer(b *testing.B) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(b, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(b, err)

	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, localReceipt, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(b, err)
	assert.Equal(b, receipt.Ty, int32(types.ExecOk))
	assert.Greater(b, len(localReceipt.KV), 0)
	accountID := uint64(4)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(b, err)
	assert.Equal(b, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(b, acc4token1Balance.TokenId, tokenId)

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(b, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(b, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(b, err)

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyTransferAction], zksyncTypes.TyTransferAction)
	assert.Nil(b, err)
	assert.Equal(b, receipt.Ty, int32(types.ExecOk))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyTransferToNewAction], zksyncTypes.TyTransferToNewAction)
	assert.Nil(b, err)
	assert.Equal(b, receipt.Ty, int32(types.ExecOk))

	//测试向新账户进行转币操作
	toEthAddr := "12a0e25e62c1dbd32e505446062b26aecb65f028"
	toL2Chain33Addr := "2afff20cc3c20f9def369626463fb027ebeba0bd976025f68316bb8eab55d48c"
	//toAddrprivkey := "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
	receipt, localReceipt, err = transfer2New(zksyncHandle, acc1privkey, tokenId, accountID, "200", toEthAddr, toL2Chain33Addr)
	assert.Nil(b, err)
	assert.Equal(b, receipt.Ty, int32(types.ExecOk))
	assert.Greater(b, len(localReceipt.KV), 0)
	//继续发送交易
	fromAccountId := accountID
	toAccountId := accountID + 1
	receipt, localReceipt, err = transfer(zksyncHandle, acc1privkey, fromAccountId, toAccountId, tokenId, "200")
	assert.Nil(b, err)
	assert.Equal(b, receipt.Ty, int32(types.ExecOk))
	assert.Greater(b, len(localReceipt.KV), 0)

	//确认发送者的balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(b, err)
	tranferFee := 100000 * 2
	balance := fmt.Sprintf("%d", int64(1000000000000)-int64(200*2)-int64(tranferFee))
	fmt.Println("Balance is", balance)
	assert.Equal(b, balance, acc4token1Balance.Balance)
	assert.Equal(b, acc4token1Balance.TokenId, tokenId)

	//确认接收者的balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), toAccountId, tokenId)
	assert.Nil(b, err)
	toBalance := fmt.Sprintf("%d", 200*2)
	fmt.Println("Balance is", toBalance)
	assert.Equal(b, acc4token1Balance.Balance, toBalance)
	assert.Equal(b, acc4token1Balance.TokenId, tokenId)
}
