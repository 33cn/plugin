package executor

import (
	"strconv"
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	x2eTy "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
	"github.com/pkg/errors"
)

type action struct {
	api          client.QueueProtocolAPI
	coinsAccount *account.DB
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	index        int32
	execaddr     string
	oracle       *Oracle
}

func newAction(a *x2ethereum, tx *types.Transaction, index int32) *action {
	hash := tx.Hash()
	fromaddr := tx.From()

	return &action{a.GetAPI(), a.GetCoinsAccount(), a.GetStateDB(), hash, fromaddr,
		a.GetBlockTime(), a.GetHeight(), index, address.ExecAddress(string(tx.Execer)), NewOracle(a.GetStateDB(), x2eTy.DefaultConsensusNeeded)}
}

// ethereum ---> chain33
// lock
func (a *action) procEth2Chain33_lock(ethBridgeClaim *x2eTy.Eth2Chain33) (*types.Receipt, error) {
	ethBridgeClaim.IssuerDotSymbol = strings.ToLower(ethBridgeClaim.IssuerDotSymbol)

	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	status, err := a.oracle.ProcessClaim(*ethBridgeClaim)
	if err != nil {
		return nil, err
	}

	ID := strconv.Itoa(int(ethBridgeClaim.EthereumChainID)) + strconv.Itoa(int(ethBridgeClaim.Nonce)) + ethBridgeClaim.EthereumSender + ethBridgeClaim.TokenContractAddress + "lock"

	//记录ethProphecy
	bz, err := a.db.Get(x2eTy.CalProphecyPrefix(ID))
	if err != nil {
		return nil, x2eTy.ErrProphecyGet
	}

	var dbProphecy x2eTy.ReceiptEthProphecy
	err = types.Decode(bz, &dbProphecy)
	if err != nil {
		return nil, types.ErrUnmarshal
	}

	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   x2eTy.CalProphecyPrefix(ID),
		Value: bz,
	})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: x2eTy.TyProphecyLog, Log: types.Encode(&dbProphecy)})

	if status.Text == x2eTy.EthBridgeStatus_SuccessStatusText {
		// mavl-x2ethereum-eth+tokenAddress
		// 这里为了区分相同tokensymbol不同tokenAddress做了级联处理
		accDB, err := account.NewAccountDB(a.api.GetConfig(), x2eTy.X2ethereumX, strings.ToLower(ethBridgeClaim.IssuerDotSymbol+ethBridgeClaim.TokenContractAddress), a.db)
		if err != nil {
			return nil, errors.Wrapf(err, "relay procMsgEth2Chain33,exec=%s,sym=%s", x2eTy.X2ethereumX, ethBridgeClaim.IssuerDotSymbol)
		}

		r, err := a.oracle.ProcessSuccessfulClaimForLock(status.FinalClaim, a.execaddr, accDB)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)

		//记录成功lock的日志
		msgEthBridgeClaimBytes := types.Encode(ethBridgeClaim)

		receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalEth2Chain33Prefix(), Value: msgEthBridgeClaimBytes})

		execlog := &types.ReceiptLog{Ty: x2eTy.TyEth2Chain33Log, Log: types.Encode(&x2eTy.ReceiptEth2Chain33{
			EthereumChainID:       ethBridgeClaim.EthereumChainID,
			BridgeContractAddress: ethBridgeClaim.BridgeContractAddress,
			Nonce:                 ethBridgeClaim.Nonce,
			IssuerDotSymbol:       ethBridgeClaim.IssuerDotSymbol,
			EthereumSender:        ethBridgeClaim.EthereumSender,
			Chain33Receiver:       ethBridgeClaim.Chain33Receiver,
			ValidatorAddress:      ethBridgeClaim.ValidatorAddress,
			Amount:                ethBridgeClaim.Amount,
			ClaimType:             ethBridgeClaim.ClaimType,
			XTxHash:               a.txhash,
			XHeight:               uint64(a.height),
			ProphecyID:            ID,
			Decimals:              ethBridgeClaim.Decimals,
			TokenAddress:          ethBridgeClaim.TokenContractAddress,
		})}
		receipt.Logs = append(receipt.Logs, execlog)

	}

	receipt.Ty = types.ExecOk
	return receipt, nil
}

// chain33 -> ethereum
// 返还在chain33上生成的erc20代币
func (a *action) procChain33ToEth_burn(msgBurn *x2eTy.Chain33ToEth) (*types.Receipt, error) {
	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	accDB, err := account.NewAccountDB(a.api.GetConfig(), x2eTy.X2ethereumX, strings.ToLower(msgBurn.IssuerDotSymbol+msgBurn.TokenContract), a.db)
	if err != nil {
		return nil, errors.Wrapf(err, "relay procMsgBurn,exec=%s,sym=%s", x2eTy.X2ethereumX, msgBurn.IssuerDotSymbol)
	}
	r, err := a.oracle.ProcessBurn(a.fromaddr, msgBurn.Amount, accDB)
	if err != nil {
		return nil, err
	}

	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	execlog := &types.ReceiptLog{Ty: x2eTy.TyWithdrawChain33Log, Log: types.Encode(&x2eTy.ReceiptChain33ToEth{
		Chain33Sender:    a.fromaddr,
		EthereumReceiver: msgBurn.EthereumReceiver,
		Amount:           msgBurn.Amount,
		IssuerDotSymbol:  msgBurn.IssuerDotSymbol,
		Decimals:         msgBurn.Decimals,
		TokenContract:    msgBurn.TokenContract,
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	msgBurnBytes := types.Encode(msgBurn)

	receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalWithdrawChain33Prefix(), Value: msgBurnBytes})

	receipt.Ty = types.ExecOk
	return receipt, nil
}

func (a *action) procChain33ToEth_lock(msgLock *x2eTy.Chain33ToEth) (*types.Receipt, error) {
	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	var accDB *account.DB
	exec, symbol, _ := x2eTy.DivideDot(msgLock.IssuerDotSymbol)
	if exec == "coins" {
		accDB = account.NewCoinsAccount(a.api.GetConfig())
		accDB.SetDB(a.db)
	} else {
		accDB, err = account.NewAccountDB(a.api.GetConfig(), exec, strings.ToLower(symbol), a.db)
		if err != nil {
			return nil, errors.Wrap(err, "newAccountDB")
		}
	}
	r, err := a.oracle.ProcessLock(a.fromaddr, address.ExecAddress(symbol), a.execaddr, msgLock.Amount, accDB)
	if err != nil {
		return nil, err
	}

	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	execlog := &types.ReceiptLog{Ty: x2eTy.TyChain33ToEthLog, Log: types.Encode(&x2eTy.ReceiptChain33ToEth{
		Chain33Sender:    a.fromaddr,
		EthereumReceiver: msgLock.EthereumReceiver,
		Amount:           msgLock.Amount,
		IssuerDotSymbol:  msgLock.IssuerDotSymbol,
		Decimals:         msgLock.Decimals,
		TokenContract:    msgLock.TokenContract,
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	msgLockBytes := types.Encode(msgLock)

	receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalChain33ToEthPrefix(), Value: msgLockBytes})

	receipt.Ty = types.ExecOk
	return receipt, nil
}

// ethereum -> chain33
// burn
func (a *action) procEth2Chain33_burn(withdrawEth *x2eTy.Eth2Chain33) (*types.Receipt, error) {
	elog.Info("procWithdrawEth", "receive a procWithdrawEth tx", "start")

	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	status, err := a.oracle.ProcessClaim(*withdrawEth)
	if err != nil {
		return nil, err
	}

	ID := strconv.Itoa(int(withdrawEth.EthereumChainID)) + strconv.Itoa(int(withdrawEth.Nonce)) + withdrawEth.EthereumSender + withdrawEth.TokenContractAddress + "burn"

	//记录ethProphecy
	bz, err := a.db.Get(x2eTy.CalProphecyPrefix(ID))
	if err != nil {
		return nil, x2eTy.ErrProphecyGet
	}

	var dbProphecy x2eTy.ReceiptEthProphecy
	err = types.Decode(bz, &dbProphecy)
	if err != nil {
		return nil, types.ErrUnmarshal
	}

	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   x2eTy.CalProphecyPrefix(ID),
		Value: bz,
	})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: x2eTy.TyProphecyLog, Log: types.Encode(&dbProphecy)})

	if status.Text == x2eTy.EthBridgeStatus_SuccessStatusText {

		var accDB *account.DB
		exec, symbol, _ := x2eTy.DivideDot(withdrawEth.IssuerDotSymbol)
		if exec == "coins" {
			accDB = account.NewCoinsAccount(a.api.GetConfig())
			accDB.SetDB(a.db)
		} else {
			accDB, err = account.NewAccountDB(a.api.GetConfig(), exec, strings.ToLower(symbol), a.db)
			if err != nil {
				return nil, errors.Wrap(err, "newAccountDB")
			}
		}

		r, err := a.oracle.ProcessSuccessfulClaimForBurn(status.FinalClaim, a.execaddr, symbol, accDB)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)

		msgWithdrawEthBytes := types.Encode(withdrawEth)

		receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalWithdrawEthPrefix(), Value: msgWithdrawEthBytes})

		execlog := &types.ReceiptLog{Ty: x2eTy.TyWithdrawEthLog, Log: types.Encode(&x2eTy.ReceiptEth2Chain33{
			EthereumChainID:       withdrawEth.EthereumChainID,
			BridgeContractAddress: withdrawEth.BridgeContractAddress,
			Nonce:                 withdrawEth.Nonce,
			IssuerDotSymbol:       withdrawEth.IssuerDotSymbol,
			EthereumSender:        withdrawEth.EthereumSender,
			Chain33Receiver:       withdrawEth.Chain33Receiver,
			ValidatorAddress:      withdrawEth.ValidatorAddress,
			Amount:                withdrawEth.Amount,
			ClaimType:             withdrawEth.ClaimType,
			XTxHash:               a.txhash,
			XHeight:               uint64(a.height),
			ProphecyID:            ID,
			Decimals:              withdrawEth.Decimals,
			TokenAddress:          withdrawEth.TokenContractAddress,
		})}
		receipt.Logs = append(receipt.Logs, execlog)

	}

	receipt.Ty = types.ExecOk
	return receipt, nil
}

func (a *action) procMsgTransfer(msgTransfer *types.AssetsTransfer) (*types.Receipt, error) {
	token := msgTransfer.GetCointoken()

	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	accDB, err := account.NewAccountDB(a.api.GetConfig(), x2eTy.X2ethereumX, token, a.db)
	if err != nil {
		return nil, err
	}
	r, err := accDB.ExecTransfer(a.fromaddr, msgTransfer.To, address.ExecAddress(x2eTy.X2ethereumX), msgTransfer.Amount)
	if err != nil {
		return nil, err
	}

	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	receipt.Ty = types.ExecOk
	return receipt, nil
}

func (a *action) procMsgTransferToExec(msgTransferToExec *types.AssetsTransferToExec) (*types.Receipt, error) {
	token := msgTransferToExec.GetCointoken()

	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	accDB, err := account.NewAccountDB(a.api.GetConfig(), x2eTy.X2ethereumX, token, a.db)
	if err != nil {
		return nil, err
	}

	r, err := accDB.TransferToExec(a.fromaddr, address.ExecAddress(msgTransferToExec.ExecName), msgTransferToExec.Amount)
	if err != nil {
		return nil, err
	}

	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil

}

func (a *action) procMsgWithDrawFromExec(msgWithdrawFromExec *types.AssetsWithdraw) (*types.Receipt, error) {
	token := msgWithdrawFromExec.GetCointoken()

	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	accDB, err := account.NewAccountDB(a.api.GetConfig(), x2eTy.X2ethereumX, token, a.db)
	if err != nil {
		return nil, err
	}

	r, err := accDB.TransferWithdraw(a.fromaddr, address.ExecAddress(msgWithdrawFromExec.ExecName), msgWithdrawFromExec.Amount)
	if err != nil {
		return nil, err
	}

	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

//需要一笔交易来注册validator
//这里注册的validator的power之和可能不为1，需要在内部进行加权
//返回的回执中，KV包含所有validator的power值，Log中包含本次注册的validator的power值
func (a *action) procAddValidator(msgAddValidator *x2eTy.MsgValidator) (*types.Receipt, error) {
	elog.Info("procAddValidator", "start", msgAddValidator)

	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	if !x2eTy.CheckPower(msgAddValidator.Power) {
		return nil, x2eTy.ErrInvalidPower
	}

	r, err := a.oracle.ProcessAddValidator(msgAddValidator.Address, msgAddValidator.Power)
	if err != nil {
		return nil, err
	}

	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	execlog := &types.ReceiptLog{Ty: x2eTy.TyAddValidatorLog, Log: types.Encode(&x2eTy.ReceiptValidator{
		Address: msgAddValidator.Address,
		Power:   msgAddValidator.Power,
		XTxHash: a.txhash,
		XHeight: uint64(a.height),
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	receipt.Ty = types.ExecOk
	return receipt, nil
}

func (a *action) procRemoveValidator(msgRemoveValidator *x2eTy.MsgValidator) (*types.Receipt, error) {
	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	r, err := a.oracle.ProcessRemoveValidator(msgRemoveValidator.Address)
	if err != nil {
		return nil, err
	}

	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	execlog := &types.ReceiptLog{Ty: x2eTy.TyRemoveValidatorLog, Log: types.Encode(&x2eTy.ReceiptValidator{
		Address: msgRemoveValidator.Address,
		Power:   msgRemoveValidator.Power,
		XTxHash: a.txhash,
		XHeight: uint64(a.height),
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	receipt.Ty = types.ExecOk
	return receipt, nil
}

func (a *action) procModifyValidator(msgModifyValidator *x2eTy.MsgValidator) (*types.Receipt, error) {
	receipt, err := a.checkConsensusThreshold()
	if err != nil {
		return nil, err
	}

	if !x2eTy.CheckPower(msgModifyValidator.Power) {
		return nil, x2eTy.ErrInvalidPower
	}

	receipt, err = a.oracle.ProcessModifyValidator(msgModifyValidator.Address, msgModifyValidator.Power)
	if err != nil {
		return nil, err
	}

	execlog := &types.ReceiptLog{Ty: x2eTy.TyModifyPowerLog, Log: types.Encode(&x2eTy.ReceiptValidator{
		Address: msgModifyValidator.Address,
		Power:   msgModifyValidator.Power,
		XTxHash: a.txhash,
		XHeight: uint64(a.height),
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	receipt.Ty = types.ExecOk
	return receipt, nil
}

func (a *action) procMsgSetConsensusThreshold(msgSetConsensusThreshold *x2eTy.MsgConsensusThreshold) (*types.Receipt, error) {
	receipt := new(types.Receipt)
	if !x2eTy.CheckPower(msgSetConsensusThreshold.ConsensusThreshold) {
		return nil, x2eTy.ErrInvalidPower
	}

	preConsensusNeeded, nowConsensusNeeded, err := a.oracle.ProcessSetConsensusNeeded(msgSetConsensusThreshold.ConsensusThreshold)
	if err != nil {
		return nil, err
	}

	setConsensusThreshold := &x2eTy.ReceiptSetConsensusThreshold{
		PreConsensusThreshold: preConsensusNeeded,
		NowConsensusThreshold: nowConsensusNeeded,
		XTxHash:               a.txhash,
		XHeight:               uint64(a.height),
	}
	execlog := &types.ReceiptLog{Ty: x2eTy.TySetConsensusThresholdLog, Log: types.Encode(setConsensusThreshold)}
	receipt.Logs = append(receipt.Logs, execlog)

	msgSetConsensusThresholdBytes := types.Encode(&x2eTy.ReceiptQueryConsensusThreshold{
		ConsensusThreshold: nowConsensusNeeded,
	})

	receipt.KV = append(receipt.KV, &types.KeyValue{Key: x2eTy.CalConsensusThresholdPrefix(), Value: msgSetConsensusThresholdBytes})

	receipt.Ty = types.ExecOk
	return receipt, nil
}

func (a *action) checkConsensusThreshold() (*types.Receipt, error) {
	receipt := new(types.Receipt)
	consensusNeededBytes, err := a.db.Get(x2eTy.CalConsensusThresholdPrefix())
	if err != nil {
		if err == types.ErrNotFound {
			setConsensusThreshold := &x2eTy.ReceiptQueryConsensusThreshold{ConsensusThreshold: x2eTy.DefaultConsensusNeeded}
			msgSetConsensusThresholdBytes := types.Encode(setConsensusThreshold)

			receipt.KV = append(receipt.KV, &types.KeyValue{
				Key:   x2eTy.CalConsensusThresholdPrefix(),
				Value: msgSetConsensusThresholdBytes,
			})
			consensusThreshold := &x2eTy.ReceiptSetConsensusThreshold{
				PreConsensusThreshold: int64(0),
				NowConsensusThreshold: x2eTy.DefaultConsensusNeeded,
				XTxHash:               a.txhash,
				XHeight:               uint64(a.height),
			}
			receipt.Logs = append(receipt.Logs, &types.ReceiptLog{Ty: x2eTy.TySetConsensusThresholdLog, Log: types.Encode(consensusThreshold)})
		} else {
			return nil, err
		}
	} else {
		var mc x2eTy.ReceiptQueryConsensusThreshold
		_ = types.Decode(consensusNeededBytes, &mc)
		_, _, err = a.oracle.ProcessSetConsensusNeeded(mc.ConsensusThreshold)
		if err != nil {
			return nil, err
		}
	}
	return receipt, nil
}
