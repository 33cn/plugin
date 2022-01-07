package executor

import (
	"errors"

	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
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
func (e *zksync) GetDriverName() string {
	return driverName
}

// CheckTx 实现自定义检验交易接口，供框架调用
func (e *zksync) CheckTx(tx *types.Transaction, index int) error {
	action := new(zt.ZksyncAction)
	if err := types.Decode(tx.Payload, action); err != nil {
		return err
	}
	pubKey := eddsa.PublicKey{}
	if _, err := pubKey.SetBytes(action.GetPubKey()); err != nil {
		return err
	}
	var msg []byte
	switch action.GetTy() {
	case zt.TyDepositAction:
		msg = types.Encode(action.GetDeposit())
	case zt.TyWithdrawAction:
		msg = types.Encode(action.GetWithdraw())
	case zt.TyContractToLeafAction:
		msg = types.Encode(action.GetContractToLeaf())
	case zt.TyLeafToContractAction:
		msg = types.Encode(action.GetLeafToContract())
	case zt.TyTransferAction:
		msg = types.Encode(action.GetTransfer())
	case zt.TyTransferToNewAction:
		msg = types.Encode(action.GetTransferToNew())
	case zt.TyForceExitAction:
		msg = types.Encode(action.GetForceQuit())
	case zt.TySetPubKeyAction:
		msg = types.Encode(action.GetSetPubKey())
	default:
		return types.ErrNotSupport
	}
	success, err := pubKey.Verify(action.GetSignInfo(), msg, mimc.NewMiMC(mixTy.MimcHashSeed))
	if err != nil {
		return err
	}
	if !success {
		return errors.New("verify sign failed")
	}
	return nil
}

//ExecutorOrder Exec 的时候 同时执行 ExecLocal
func (e *zksync) ExecutorOrder() int64 {
	return drivers.ExecLocalSameTime
}

// GetPayloadValue get payload value
func (e *zksync) GetPayloadValue() types.Message {
	return &zt.ZksyncAction{}
}
