package executor

import (
	"errors"

	"github.com/33cn/chain33/common"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
)

/*
 * 执行器相关定义
 * 重载基类相关接口
 */

var (
	//日志
	zlog = log.New("module", "zksync.executor")
)

var driverName = zt.Zksync
var historyProof zt.HistoryAccountProofInfo

// Init register dapp
func Init(name string, cfg *types.Chain33Config, sub []byte) {
	drivers.Register(cfg, GetName(), NewZksync, cfg.GetDappFork(driverName, "Enable"))
	InitExecType()
}

// InitExecType Init Exec Type
func InitExecType() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&zksync{}))
}

type zksync struct {
	drivers.DriverBase
}

//NewExchange ...
func NewZksync() drivers.Driver {
	t := &zksync{}
	t.SetChild(t)
	t.SetExecutorType(types.LoadExecutorType(driverName))
	return t
}

// GetName get driver name
func GetName() string {
	return NewZksync().GetName()
}

//GetDriverName ...
func (z *zksync) GetDriverName() string {
	return driverName
}

// CheckTx 实现自定义检验交易接口，供框架调用
func (z *zksync) CheckTx(tx *types.Transaction, index int) error {
	action := new(zt.ZksyncAction)
	if err := types.Decode(tx.Payload, action); err != nil {
		return err
	}
	var signature *zt.ZkSignature
	var msg *zt.ZkMsg
	switch action.GetTy() {
	case zt.TyDepositAction:
		signature = action.GetDeposit().GetSignature()
		msg = wallet.GetDepositMsg(action.GetDeposit())
	case zt.TyWithdrawAction:
		signature = action.GetZkWithdraw().GetSignature()
		msg = wallet.GetWithdrawMsg(action.GetZkWithdraw())
	case zt.TyContractToTreeAction:
		signature = action.GetContractToTree().GetSignature()
		msg = wallet.GetContractToTreeMsg(action.GetContractToTree())
	case zt.TyTreeToContractAction:
		signature = action.GetTreeToContract().GetSignature()
		msg = wallet.GetTreeToContractMsg(action.GetTreeToContract())
	case zt.TyTransferAction:
		signature = action.GetZkTransfer().GetSignature()
		msg = wallet.GetTransferMsg(action.GetZkTransfer())
	case zt.TyTransferToNewAction:
		signature = action.GetTransferToNew().GetSignature()
		msg = wallet.GetTransferToNewMsg(action.GetTransferToNew())
	case zt.TyProxyExitAction:
		signature = action.GetProxyExit().GetSignature()
		msg = wallet.GetProxyExitMsg(action.GetProxyExit())
	case zt.TySetPubKeyAction:
		signature = action.GetSetPubKey().GetSignature()
		msg = wallet.GetSetPubKeyMsg(action.GetSetPubKey())
	case zt.TyFullExitAction:
		signature = action.GetFullExit().GetSignature()
		msg = wallet.GetFullExitMsg(action.GetFullExit())
	case zt.TyMintNFTAction:
		signature = action.GetMintNFT().GetSignature()
		msg = wallet.GetMintNFTMsg(action.GetMintNFT())
	case zt.TyWithdrawNFTAction:
		signature = action.GetWithdrawNFT().GetSignature()
		msg = wallet.GetWithdrawNFTMsg(action.GetWithdrawNFT())
	case zt.TyTransferNFTAction:
		signature = action.GetTransferNFT().GetSignature()
		msg = wallet.GetTransferNFTMsg(action.GetTransferNFT())
	default:
		return nil
	}
	if signature == nil {
		zlog.Error("checkTx.signature is nil", "ty", action.GetTy())
		return types.ErrInvalidParam
	}
	pubKey := eddsa.PublicKey{}
	pubKey.A.X.SetString(signature.PubKey.X)
	pubKey.A.Y.SetString(signature.PubKey.Y)
	signInfo, err := common.FromHex(signature.GetSignInfo())
	if err != nil {
		return err
	}
	success, err := pubKey.Verify(signInfo, wallet.GetMsgHash(msg), mimc.NewMiMC(zt.ZkMimcHashSeed))
	if err != nil {
		return err
	}
	if !success {
		return errors.New("verify sign failed")
	}
	return nil
}

//ExecutorOrder Exec 的时候 同时执行 ExecLocal
func (z *zksync) ExecutorOrder() int64 {
	return drivers.ExecLocalSameTime
}

// GetPayloadValue get payload value
func (z *zksync) GetPayloadValue() types.Message {
	return &zt.ZksyncAction{}
}
