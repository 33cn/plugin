package wallet

import (
	"bytes"
	"fmt"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"sync"
)

var (
	bizlog = log15.New("module", "wallet.zksync")
	// MaxTxHashsPerTime 单词处理的最大哈希书
	MaxTxHashsPerTime int64 = 100
	// maxTxNumPerBlock 单个区块最大数
	maxTxNumPerBlock int64 = types.MaxTxsPerBlock
)

func init() {
	wcom.RegisterPolicy(zt.Zksync, New())
}

// New 创建一盒钱包业务策略
func New() wcom.WalletBizPolicy {
	return &zksyncPolicy{
		mtx:            &sync.Mutex{},
	}
}

type zksyncPolicy struct {
	mtx            *sync.Mutex
	walletOperate  wcom.WalletOperate
}

func (policy *zksyncPolicy) setWalletOperate(walletBiz wcom.WalletOperate) {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	policy.walletOperate = walletBiz
}

func (policy *zksyncPolicy) getWalletOperate() wcom.WalletOperate {
	policy.mtx.Lock()
	defer policy.mtx.Unlock()
	return policy.walletOperate
}

// Init 初始化处理
func (policy *zksyncPolicy) Init(walletOperate wcom.WalletOperate, sub []byte) {
	policy.setWalletOperate(walletOperate)
}

// OnCreateNewAccount 在账号创建时做一些处理
func (policy *zksyncPolicy) OnCreateNewAccount(acc *types.Account) {
}

// OnImportPrivateKey 在私钥导入时做一些处理
func (policy *zksyncPolicy) OnImportPrivateKey(acc *types.Account) {
}

// OnAddBlockFinish 在区块被添加成功时做一些处理
func (policy *zksyncPolicy) OnAddBlockFinish(block *types.BlockDetail) {

}

// OnDeleteBlockFinish 在区块被删除成功时做一些处理
func (policy *zksyncPolicy) OnDeleteBlockFinish(block *types.BlockDetail) {

}

// OnClose 在钱包关闭时做一些处理
func (policy *zksyncPolicy) OnClose() {

}

// OnSetQueueClient 在钱包消息队列初始化时做一些处理
func (policy *zksyncPolicy) OnSetQueueClient() {
}

// OnWalletLocked 在钱包加锁时做一些处理
func (policy *zksyncPolicy) OnWalletLocked() {
}

// OnWalletUnlocked 在钱包解锁时做一些处理
func (policy *zksyncPolicy) OnWalletUnlocked(WalletUnLock *types.WalletUnLock) {
}

// Call 调用隐私的方法
func (policy *zksyncPolicy) Call(funName string, in types.Message) (ret types.Message, err error) {
	err = types.ErrNotSupport
	return
}

// SignTransaction 对zksync交易进行签名
func (policy *zksyncPolicy) SignTransaction(key crypto.PrivKey, req *types.ReqSignRawTx) (needSysSign bool, signtxhex string, err error) {
	needSysSign = true
	bytesVal, err := common.FromHex(req.GetTxHex())
	if err != nil {
		bizlog.Error("SignTransaction", "common.FromHex error", err)
		return
	}
	tx := new(types.Transaction)
	if err = types.Decode(bytesVal, tx); err != nil {
		bizlog.Error("SignTransaction", "Decode Transaction error", err)
		return
	}

	privateKey, err := eddsa.GenerateKey(bytes.NewReader(key.Bytes()))
	if err != nil {
		bizlog.Error("SignTransaction", "eddsa.GenerateKey error", err)
		return
	}


	sigNature := new(zt.EddsaSigNature)
	if err = types.Decode(tx.Payload, sigNature); err != nil {
		bizlog.Error("SignTransaction", "Decode EddsaSigNature error", err)
		return
	}

	privateKey.PublicKey.Bytes()

	sign, err := privateKey.Sign(types.Encode(sigNature.GetAction()), mimc.NewMiMC(mixTy.MimcHashSeed))
	sigNature.SignInfo = sign
	sigNature.PubKey = privateKey.PublicKey.Bytes()
	tx.Payload = types.Encode(sigNature)
	signtxhex = common.ToHex(types.Encode(tx))
	return
}


// OnAddBlockTx 响应区块交易添加的处理
func (policy *zksyncPolicy) OnAddBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbbatch db.Batch) *types.WalletTxDetail {
	txdetail := &types.WalletTxDetail {
	}

	blockheight := block.Block.Height*maxTxNumPerBlock + int64(index)
	heightstr := fmt.Sprintf("%018d", blockheight)
	key := wcom.CalcTxKey(heightstr)
	txdetail.Tx = tx
	txdetail.Height = block.Block.Height
	txdetail.Index = int64(index)
	txdetail.Receipt = block.Receipts[index]
	txdetail.Blocktime = block.Block.BlockTime

	txdetail.ActionName = tx.ActionName()
	txdetail.Amount, _ = tx.Amount()
	txdetail.Txhash = tx.Hash()

	pubkey := block.Block.Txs[index].Signature.GetPubkey()
	addr := address.PubKeyToAddress(pubkey)
	txdetail.Fromaddr = addr.String()

	txdetailbyte := types.Encode(txdetail)
	dbbatch.Set(key, txdetailbyte)

	return nil
}

// OnDeleteBlockTx 响应删除区块交易的处理
func (policy *zksyncPolicy) OnDeleteBlockTx(block *types.BlockDetail, tx *types.Transaction, index int32, dbbatch db.Batch) *types.WalletTxDetail {
	blockheight := block.Block.Height*maxTxNumPerBlock + int64(index)
	heightstr := fmt.Sprintf("%018d", blockheight)
	dbbatch.Delete(wcom.CalcTxKey(heightstr))
	// 自己处理掉所有事务，不需要外部处理了
	return nil
}
