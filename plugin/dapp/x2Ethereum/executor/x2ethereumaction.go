package executor

import (
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	chain33types "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/executor/ethbridge"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/executor/oracle"
	types2 "github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
	"github.com/pkg/errors"
)

// stateDB存储KV:
//		CalProphecyPrefix --> DBProphecy
//		CalEth2Chain33Prefix -- > Eth2Chain33
//		CalWithdrawEthPrefix -- > Eth2Chain33
//		CalWithdrawChain33Prefix -- > Chain33ToEth
//		CalChain33ToEthPrefix -- > Chain33ToEth
//		CalValidatorMapsPrefix -- > MsgValidator maps
//		CalLastTotalPowerPrefix -- > ReceiptQueryTotalPower
//		CalConsensusThresholdPrefix -- > ReceiptSetConsensusThreshold
//		CalTokenSymbolTotalAmountPrefix -- > ReceiptQuerySymbolAssets

// 当前存在一个问题：
// token的发行需要提前授权，所以账户模型该如何设计？
//
// 解决方案：
// 当eth-->chain33时，采用 mavl-x2ethereum-symbol的账户模型，但是这样的资产是无法提现的，是一个完全虚拟的资产
// 而在chain33-->eth时，采用 mavl-coins-bty 的账户模型（后续可以升级为mavl-token-symbol以支持多个token资产）

// token 合约转币到x2ethereum合约
// 个人账户地址 = mavl-token-symbol-execAddr:aliceAddr
// 不同币种账户地址 = mavl-token-symbol-execAddr

// eth -- > chain33:
// 在 mavl-token-symbol-execAddr 上铸币，然后转到 mavl-token-symbol-execAddr:addr 上
// withdraw 的时候先从mavl-coins-symbol-execAddr:addr 中withdraw到 mavl-token-symbol-execAddr，然后 burn

// chain33 -- > eth:
// 在 mavl-token-symbol-execAddr:addr 上withdraw到 mavl-token-symbol-execAddr 上，然后frozen住
// withdraw 的时候从 mavl-token-symbol-execAddr 上 deposit mavl-token-symbol-execAddr:addr

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
	keeper       ethbridge.Keeper
}

func newAction(a *x2ethereum, tx *chain33types.Transaction, index int32) *action {
	hash := tx.Hash()
	fromaddr := tx.From()

	oracleKeeper := oracle.NewOracleKeeper(a.GetStateDB(), types2.DefaultConsensusNeeded)
	if oracleKeeper == nil {
		return nil
	}

	elog.Info("newAction", "newAction", "done")
	return &action{a.GetAPI(), a.GetCoinsAccount(), a.GetStateDB(), hash, fromaddr,
		a.GetBlockTime(), a.GetHeight(), index, address.ExecAddress(string(tx.Execer)), ethbridge.NewKeeper(*oracleKeeper, a.GetStateDB())}
}

// ethereum ---> chain33
// lock
func (a *action) procMsgEth2Chain33(ethBridgeClaim *types2.Eth2Chain33) (*chain33types.Receipt, error) {
	receipt := new(chain33types.Receipt)
	ethBridgeClaim.LocalCoinSymbol = strings.ToLower(ethBridgeClaim.LocalCoinSymbol)

	consensusNeededBytes, err := a.db.Get(types2.CalConsensusThresholdPrefix())
	if err != nil {
		if err == chain33types.ErrNotFound {
			setConsensusThreshold := &types2.ReceiptQueryConsensusThreshold{ConsensusThreshold: types2.DefaultConsensusNeeded}
			msgSetConsensusThresholdBytes, err := proto.Marshal(setConsensusThreshold)
			if err != nil {
				return nil, chain33types.ErrMarshal
			}
			receipt.KV = append(receipt.KV, &chain33types.KeyValue{
				Key:   types2.CalConsensusThresholdPrefix(),
				Value: msgSetConsensusThresholdBytes,
			})
			consensusThreshold := &types2.ReceiptSetConsensusThreshold{
				PreConsensusThreshold: int64(0),
				NowConsensusThreshold: types2.DefaultConsensusNeeded,
				XTxHash:               a.txhash,
				XHeight:               uint64(a.height),
			}
			receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TySetConsensusThresholdLog, Log: chain33types.Encode(consensusThreshold)})
		} else {
			return nil, err
		}
	} else {
		var mc types2.ReceiptQueryConsensusThreshold
		_ = proto.Unmarshal(consensusNeededBytes, &mc)
		_, _, err = a.keeper.ProcessSetConsensusNeeded(mc.ConsensusThreshold)
		if err != nil {
			return nil, err
		}
	}

	status, err := a.keeper.ProcessClaim(*ethBridgeClaim)
	if err != nil {
		return nil, err
	}

	ID := strconv.Itoa(int(ethBridgeClaim.EthereumChainID)) + strconv.Itoa(int(ethBridgeClaim.Nonce)) + ethBridgeClaim.EthereumSender + ethBridgeClaim.TokenContractAddress + "lock"

	//记录ethProphecy
	bz, err := a.db.Get(types2.CalProphecyPrefix(ID))
	if err != nil {
		return nil, types2.ErrProphecyGet
	}

	var dbProphecy types2.ReceiptEthProphecy
	err = proto.Unmarshal(bz, &dbProphecy)
	if err != nil {
		return nil, chain33types.ErrUnmarshal
	}

	receipt.KV = append(receipt.KV, &chain33types.KeyValue{
		Key:   types2.CalProphecyPrefix(ID),
		Value: bz,
	})
	receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TyProphecyLog, Log: chain33types.Encode(&dbProphecy)})

	if status.Text == types2.EthBridgeStatus_SuccessStatusText {
		// mavl-x2ethereum-eth
		// 这里为了区分相同tokensymbol不同tokenAddress做了级联处理
		accDB, err := account.NewAccountDB(a.api.GetConfig(), ethBridgeClaim.LocalCoinExec, strings.ToLower(ethBridgeClaim.LocalCoinSymbol+ethBridgeClaim.TokenContractAddress), a.db)
		if err != nil {
			return nil, errors.Wrapf(err, "relay procMsgEth2Chain33,exec=%s,sym=%s", ethBridgeClaim.LocalCoinExec, ethBridgeClaim.LocalCoinSymbol)
		}

		r, err := a.keeper.ProcessSuccessfulClaimForLock(status.FinalClaim, a.execaddr, ethBridgeClaim.LocalCoinSymbol, ethBridgeClaim.TokenContractAddress, accDB)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)

		//记录成功lock的日志
		msgEthBridgeClaimBytes, err := proto.Marshal(ethBridgeClaim)
		if err != nil {
			return nil, chain33types.ErrMarshal
		}
		receipt.KV = append(receipt.KV, &chain33types.KeyValue{Key: types2.CalEth2Chain33Prefix(), Value: msgEthBridgeClaimBytes})

		execlog := &chain33types.ReceiptLog{Ty: types2.TyEth2Chain33Log, Log: chain33types.Encode(&types2.ReceiptEth2Chain33{
			EthereumChainID:       ethBridgeClaim.EthereumChainID,
			BridgeContractAddress: ethBridgeClaim.BridgeContractAddress,
			Nonce:                 ethBridgeClaim.Nonce,
			LocalCoinSymbol:       ethBridgeClaim.LocalCoinSymbol,
			LocalCoinExec:         ethBridgeClaim.LocalCoinExec,
			TokenContractAddress:  ethBridgeClaim.TokenContractAddress,
			EthereumSender:        ethBridgeClaim.EthereumSender,
			Chain33Receiver:       ethBridgeClaim.Chain33Receiver,
			ValidatorAddress:      ethBridgeClaim.ValidatorAddress,
			Amount:                ethBridgeClaim.Amount,
			ClaimType:             ethBridgeClaim.ClaimType,
			XTxHash:               a.txhash,
			XHeight:               uint64(a.height),
			ProphecyID:            ID,
			Decimals:              ethBridgeClaim.Decimals,
		})}
		receipt.Logs = append(receipt.Logs, execlog)

	}

	receipt.Ty = chain33types.ExecOk
	return receipt, nil
}

// chain33 -> ethereum
// 返还在chain33上生成的erc20代币
func (a *action) procMsgBurn(msgBurn *types2.Chain33ToEth) (*chain33types.Receipt, error) {
	msgBurn.LocalCoinExec = types2.X2ethereumX
	receipt := new(chain33types.Receipt)

	consensusNeededBytes, err := a.db.Get(types2.CalConsensusThresholdPrefix())
	if err != nil {
		if err == chain33types.ErrNotFound {
			setConsensusThreshold := &types2.ReceiptQueryConsensusThreshold{ConsensusThreshold: types2.DefaultConsensusNeeded}
			msgSetConsensusThresholdBytes, err := proto.Marshal(setConsensusThreshold)
			if err != nil {
				return nil, chain33types.ErrMarshal
			}
			receipt.KV = append(receipt.KV, &chain33types.KeyValue{
				Key:   types2.CalConsensusThresholdPrefix(),
				Value: msgSetConsensusThresholdBytes,
			})
			consensusThreshold := &types2.ReceiptSetConsensusThreshold{
				PreConsensusThreshold: int64(0),
				NowConsensusThreshold: types2.DefaultConsensusNeeded,
				XTxHash:               a.txhash,
				XHeight:               uint64(a.height),
			}
			receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TySetConsensusThresholdLog, Log: chain33types.Encode(consensusThreshold)})
		} else {
			return nil, err
		}
	} else {
		var mc types2.ReceiptQueryConsensusThreshold
		_ = proto.Unmarshal(consensusNeededBytes, &mc)
		_, _, err = a.keeper.ProcessSetConsensusNeeded(mc.ConsensusThreshold)
		if err != nil {
			return nil, err
		}
	}

	accDB, err := account.NewAccountDB(a.api.GetConfig(), msgBurn.LocalCoinExec, strings.ToLower(msgBurn.LocalCoinSymbol+msgBurn.TokenContract), a.db)
	if err != nil {
		return nil, errors.Wrapf(err, "relay procMsgBurn,exec=%s,sym=%s", msgBurn.LocalCoinExec, msgBurn.LocalCoinSymbol)
	}
	receipt, err = a.keeper.ProcessBurn(a.fromaddr, a.execaddr, msgBurn.Amount, msgBurn.TokenContract, msgBurn.Decimals, accDB)
	if err != nil {
		return nil, err
	}

	execlog := &chain33types.ReceiptLog{Ty: types2.TyWithdrawChain33Log, Log: chain33types.Encode(&types2.ReceiptChain33ToEth{
		TokenContract:    msgBurn.TokenContract,
		Chain33Sender:    a.fromaddr,
		EthereumReceiver: msgBurn.EthereumReceiver,
		Amount:           msgBurn.Amount,
		EthSymbol:        msgBurn.LocalCoinSymbol,
		Decimals:         msgBurn.Decimals,
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	msgBurnBytes, err := proto.Marshal(msgBurn)
	if err != nil {
		return nil, chain33types.ErrMarshal
	}
	receipt.KV = append(receipt.KV, &chain33types.KeyValue{Key: types2.CalWithdrawChain33Prefix(), Value: msgBurnBytes})

	receipt.Ty = chain33types.ExecOk
	return receipt, nil
}

func (a *action) procMsgLock(msgLock *types2.Chain33ToEth) (*chain33types.Receipt, error) {
	msgLock.LocalCoinExec = types2.X2ethereumX
	receipt := new(chain33types.Receipt)

	consensusNeededBytes, err := a.db.Get(types2.CalConsensusThresholdPrefix())
	if err != nil {
		if err == chain33types.ErrNotFound {
			setConsensusThreshold := &types2.ReceiptQueryConsensusThreshold{ConsensusThreshold: types2.DefaultConsensusNeeded}
			msgSetConsensusThresholdBytes, err := proto.Marshal(setConsensusThreshold)
			if err != nil {
				return nil, chain33types.ErrMarshal
			}
			receipt.KV = append(receipt.KV, &chain33types.KeyValue{
				Key:   types2.CalConsensusThresholdPrefix(),
				Value: msgSetConsensusThresholdBytes,
			})
			consensusThreshold := &types2.ReceiptSetConsensusThreshold{
				PreConsensusThreshold: int64(0),
				NowConsensusThreshold: types2.DefaultConsensusNeeded,
				XTxHash:               a.txhash,
				XHeight:               uint64(a.height),
			}
			receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TySetConsensusThresholdLog, Log: chain33types.Encode(consensusThreshold)})
		} else {
			return nil, err
		}
	} else {
		var mc types2.ReceiptQueryConsensusThreshold
		_ = proto.Unmarshal(consensusNeededBytes, &mc)
		_, _, err = a.keeper.ProcessSetConsensusNeeded(mc.ConsensusThreshold)
		if err != nil {
			return nil, err
		}
	}

	accDB := account.NewCoinsAccount(a.api.GetConfig())
	accDB.SetDB(a.db)
	receipt, err = a.keeper.ProcessLock(a.fromaddr, address.ExecAddress(msgLock.LocalCoinSymbol), a.execaddr, msgLock.Amount, accDB)
	if err != nil {
		return nil, err
	}

	execlog := &chain33types.ReceiptLog{Ty: types2.TyChain33ToEthLog, Log: chain33types.Encode(&types2.ReceiptChain33ToEth{
		TokenContract:    msgLock.TokenContract,
		Chain33Sender:    a.fromaddr,
		EthereumReceiver: msgLock.EthereumReceiver,
		Amount:           msgLock.Amount,
		EthSymbol:        msgLock.LocalCoinSymbol,
		Decimals:         msgLock.Decimals,
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	msgLockBytes, err := proto.Marshal(msgLock)
	if err != nil {
		return nil, chain33types.ErrMarshal
	}
	receipt.KV = append(receipt.KV, &chain33types.KeyValue{Key: types2.CalChain33ToEthPrefix(), Value: msgLockBytes})

	receipt.Ty = chain33types.ExecOk
	return receipt, nil
}

// ethereum -> chain33
// burn
func (a *action) procWithdrawEth(withdrawEth *types2.Eth2Chain33) (*chain33types.Receipt, error) {
	elog.Info("procWithdrawEth", "receive a procWithdrawEth tx", "start")
	receipt := new(chain33types.Receipt)

	consensusNeededBytes, err := a.db.Get(types2.CalConsensusThresholdPrefix())
	if err != nil {
		if err == chain33types.ErrNotFound {
			setConsensusThreshold := &types2.ReceiptQueryConsensusThreshold{ConsensusThreshold: types2.DefaultConsensusNeeded}
			msgSetConsensusThresholdBytes, err := proto.Marshal(setConsensusThreshold)
			if err != nil {
				return nil, chain33types.ErrMarshal
			}
			receipt.KV = append(receipt.KV, &chain33types.KeyValue{
				Key:   types2.CalConsensusThresholdPrefix(),
				Value: msgSetConsensusThresholdBytes,
			})
			consensusThreshold := &types2.ReceiptSetConsensusThreshold{
				PreConsensusThreshold: int64(0),
				NowConsensusThreshold: types2.DefaultConsensusNeeded,
				XTxHash:               a.txhash,
				XHeight:               uint64(a.height),
			}
			receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TySetConsensusThresholdLog, Log: chain33types.Encode(consensusThreshold)})
		} else {
			return nil, err
		}
	} else {
		var mc types2.ReceiptQueryConsensusThreshold
		_ = proto.Unmarshal(consensusNeededBytes, &mc)
		_, _, err = a.keeper.ProcessSetConsensusNeeded(mc.ConsensusThreshold)
		if err != nil {
			return nil, err
		}
	}

	status, err := a.keeper.ProcessClaim(*withdrawEth)
	if err != nil {
		return nil, err
	}

	ID := strconv.Itoa(int(withdrawEth.EthereumChainID)) + strconv.Itoa(int(withdrawEth.Nonce)) + withdrawEth.EthereumSender + withdrawEth.TokenContractAddress + "burn"

	//记录ethProphecy
	bz, err := a.db.Get(types2.CalProphecyPrefix(ID))
	if err != nil {
		return nil, types2.ErrProphecyGet
	}

	var dbProphecy types2.ReceiptEthProphecy
	err = proto.Unmarshal(bz, &dbProphecy)
	if err != nil {
		return nil, chain33types.ErrUnmarshal
	}

	receipt.KV = append(receipt.KV, &chain33types.KeyValue{
		Key:   types2.CalProphecyPrefix(ID),
		Value: bz,
	})
	receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TyProphecyLog, Log: chain33types.Encode(&dbProphecy)})

	if status.Text == types2.EthBridgeStatus_SuccessStatusText {
		accDB := account.NewCoinsAccount(a.api.GetConfig())
		accDB.SetDB(a.db)
		r, err := a.keeper.ProcessSuccessfulClaimForBurn(status.FinalClaim, a.execaddr, withdrawEth.LocalCoinSymbol, accDB)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)

		msgWithdrawEthBytes, err := proto.Marshal(withdrawEth)
		if err != nil {
			return nil, chain33types.ErrMarshal
		}
		receipt.KV = append(receipt.KV, &chain33types.KeyValue{Key: types2.CalWithdrawEthPrefix(), Value: msgWithdrawEthBytes})

		execlog := &chain33types.ReceiptLog{Ty: types2.TyWithdrawEthLog, Log: chain33types.Encode(&types2.ReceiptEth2Chain33{
			EthereumChainID:       withdrawEth.EthereumChainID,
			BridgeContractAddress: withdrawEth.BridgeContractAddress,
			Nonce:                 withdrawEth.Nonce,
			LocalCoinSymbol:       withdrawEth.LocalCoinSymbol,
			LocalCoinExec:         withdrawEth.LocalCoinExec,
			TokenContractAddress:  withdrawEth.TokenContractAddress,
			EthereumSender:        withdrawEth.EthereumSender,
			Chain33Receiver:       withdrawEth.Chain33Receiver,
			ValidatorAddress:      withdrawEth.ValidatorAddress,
			Amount:                withdrawEth.Amount,
			ClaimType:             withdrawEth.ClaimType,
			XTxHash:               a.txhash,
			XHeight:               uint64(a.height),
			ProphecyID:            ID,
			Decimals:              withdrawEth.Decimals,
		})}
		receipt.Logs = append(receipt.Logs, execlog)

	}

	receipt.Ty = chain33types.ExecOk
	return receipt, nil
}

func (a *action) procMsgTransfer(msgTransfer *chain33types.AssetsTransfer) (*chain33types.Receipt, error) {
	token := msgTransfer.GetCointoken()
	receipt := new(chain33types.Receipt)

	consensusNeededBytes, err := a.db.Get(types2.CalConsensusThresholdPrefix())
	if err != nil {
		if err == chain33types.ErrNotFound {
			setConsensusThreshold := &types2.ReceiptQueryConsensusThreshold{ConsensusThreshold: types2.DefaultConsensusNeeded}
			msgSetConsensusThresholdBytes, err := proto.Marshal(setConsensusThreshold)
			if err != nil {
				return nil, chain33types.ErrMarshal
			}
			receipt.KV = append(receipt.KV, &chain33types.KeyValue{
				Key:   types2.CalConsensusThresholdPrefix(),
				Value: msgSetConsensusThresholdBytes,
			})
			consensusThreshold := &types2.ReceiptSetConsensusThreshold{
				PreConsensusThreshold: int64(0),
				NowConsensusThreshold: types2.DefaultConsensusNeeded,
				XTxHash:               a.txhash,
				XHeight:               uint64(a.height),
			}
			receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TySetConsensusThresholdLog, Log: chain33types.Encode(consensusThreshold)})
		} else {
			return nil, err
		}
	} else {
		var mc types2.ReceiptQueryConsensusThreshold
		_ = proto.Unmarshal(consensusNeededBytes, &mc)
		_, _, err = a.keeper.ProcessSetConsensusNeeded(mc.ConsensusThreshold)
		if err != nil {
			return nil, err
		}
	}

	accDB, err := account.NewAccountDB(a.api.GetConfig(), types2.X2ethereumX, token, a.db)
	if err != nil {
		return nil, err
	}
	receipt, err = accDB.ExecTransfer(a.fromaddr, msgTransfer.To, address.ExecAddress(types2.X2ethereumX), msgTransfer.Amount)
	if err != nil {
		return nil, err
	}

	receipt.Ty = chain33types.ExecOk
	return receipt, nil
}

//需要一笔交易来注册validator
//这里注册的validator的power之和可能不为1，需要在内部进行加权
//返回的回执中，KV包含所有validator的power值，Log中包含本次注册的validator的power值
func (a *action) procAddValidator(msgAddValidator *types2.MsgValidator) (*chain33types.Receipt, error) {
	elog.Info("procAddValidator", "start", msgAddValidator)
	receipt := new(chain33types.Receipt)

	consensusNeededBytes, err := a.db.Get(types2.CalConsensusThresholdPrefix())
	if err != nil {
		if err == chain33types.ErrNotFound {
			setConsensusThreshold := &types2.ReceiptQueryConsensusThreshold{ConsensusThreshold: types2.DefaultConsensusNeeded}
			msgSetConsensusThresholdBytes, err := proto.Marshal(setConsensusThreshold)
			if err != nil {
				return nil, chain33types.ErrMarshal
			}
			receipt.KV = append(receipt.KV, &chain33types.KeyValue{
				Key:   types2.CalConsensusThresholdPrefix(),
				Value: msgSetConsensusThresholdBytes,
			})
			consensusThreshold := &types2.ReceiptSetConsensusThreshold{
				PreConsensusThreshold: int64(0),
				NowConsensusThreshold: types2.DefaultConsensusNeeded,
				XTxHash:               a.txhash,
				XHeight:               uint64(a.height),
			}
			receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TySetConsensusThresholdLog, Log: chain33types.Encode(consensusThreshold)})
		} else {
			return nil, err
		}
	} else {
		var mc types2.ReceiptQueryConsensusThreshold
		_ = proto.Unmarshal(consensusNeededBytes, &mc)
		_, _, err = a.keeper.ProcessSetConsensusNeeded(mc.ConsensusThreshold)
		if err != nil {
			return nil, err
		}
	}

	if !types2.CheckPower(msgAddValidator.Power) {
		return nil, types2.ErrInvalidPower
	}

	receipt, err = a.keeper.ProcessAddValidator(msgAddValidator.Address, msgAddValidator.Power)
	if err != nil {
		return nil, err
	}

	execlog := &chain33types.ReceiptLog{Ty: types2.TyAddValidatorLog, Log: chain33types.Encode(&types2.ReceiptValidator{
		Address: msgAddValidator.Address,
		Power:   msgAddValidator.Power,
		XTxHash: a.txhash,
		XHeight: uint64(a.height),
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	receipt.Ty = chain33types.ExecOk
	return receipt, nil
}

func (a *action) procRemoveValidator(msgRemoveValidator *types2.MsgValidator) (*chain33types.Receipt, error) {
	receipt := new(chain33types.Receipt)

	consensusNeededBytes, err := a.db.Get(types2.CalConsensusThresholdPrefix())
	if err != nil {
		if err == chain33types.ErrNotFound {
			setConsensusThreshold := &types2.ReceiptQueryConsensusThreshold{ConsensusThreshold: types2.DefaultConsensusNeeded}
			msgSetConsensusThresholdBytes, err := proto.Marshal(setConsensusThreshold)
			if err != nil {
				return nil, chain33types.ErrMarshal
			}
			receipt.KV = append(receipt.KV, &chain33types.KeyValue{
				Key:   types2.CalConsensusThresholdPrefix(),
				Value: msgSetConsensusThresholdBytes,
			})
			consensusThreshold := &types2.ReceiptSetConsensusThreshold{
				PreConsensusThreshold: int64(0),
				NowConsensusThreshold: types2.DefaultConsensusNeeded,
				XTxHash:               a.txhash,
				XHeight:               uint64(a.height),
			}
			receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TySetConsensusThresholdLog, Log: chain33types.Encode(consensusThreshold)})
		} else {
			return nil, err
		}
	} else {
		var mc types2.ReceiptQueryConsensusThreshold
		_ = proto.Unmarshal(consensusNeededBytes, &mc)
		_, _, err = a.keeper.ProcessSetConsensusNeeded(mc.ConsensusThreshold)
		if err != nil {
			return nil, err
		}
	}

	receipt, err = a.keeper.ProcessRemoveValidator(msgRemoveValidator.Address)
	if err != nil {
		return nil, err
	}

	execlog := &chain33types.ReceiptLog{Ty: types2.TyRemoveValidatorLog, Log: chain33types.Encode(&types2.ReceiptValidator{
		Address: msgRemoveValidator.Address,
		Power:   msgRemoveValidator.Power,
		XTxHash: a.txhash,
		XHeight: uint64(a.height),
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	receipt.Ty = chain33types.ExecOk
	return receipt, nil
}

func (a *action) procModifyValidator(msgModifyValidator *types2.MsgValidator) (*chain33types.Receipt, error) {
	receipt := new(chain33types.Receipt)

	consensusNeededBytes, err := a.db.Get(types2.CalConsensusThresholdPrefix())
	if err != nil {
		if err == chain33types.ErrNotFound {
			setConsensusThreshold := &types2.ReceiptQueryConsensusThreshold{ConsensusThreshold: types2.DefaultConsensusNeeded}
			msgSetConsensusThresholdBytes, err := proto.Marshal(setConsensusThreshold)
			if err != nil {
				return nil, chain33types.ErrMarshal
			}
			receipt.KV = append(receipt.KV, &chain33types.KeyValue{
				Key:   types2.CalConsensusThresholdPrefix(),
				Value: msgSetConsensusThresholdBytes,
			})
			consensusThreshold := &types2.ReceiptSetConsensusThreshold{
				PreConsensusThreshold: int64(0),
				NowConsensusThreshold: types2.DefaultConsensusNeeded,
				XTxHash:               a.txhash,
				XHeight:               uint64(a.height),
			}
			receipt.Logs = append(receipt.Logs, &chain33types.ReceiptLog{Ty: types2.TySetConsensusThresholdLog, Log: chain33types.Encode(consensusThreshold)})
		} else {
			return nil, err
		}
	} else {
		var mc types2.ReceiptQueryConsensusThreshold
		_ = proto.Unmarshal(consensusNeededBytes, &mc)
		_, _, err = a.keeper.ProcessSetConsensusNeeded(mc.ConsensusThreshold)
		if err != nil {
			return nil, err
		}
	}

	if !types2.CheckPower(msgModifyValidator.Power) {
		return nil, types2.ErrInvalidPower
	}

	receipt, err = a.keeper.ProcessModifyValidator(msgModifyValidator.Address, msgModifyValidator.Power)
	if err != nil {
		return nil, err
	}

	execlog := &chain33types.ReceiptLog{Ty: types2.TyModifyPowerLog, Log: chain33types.Encode(&types2.ReceiptValidator{
		Address: msgModifyValidator.Address,
		Power:   msgModifyValidator.Power,
		XTxHash: a.txhash,
		XHeight: uint64(a.height),
	})}
	receipt.Logs = append(receipt.Logs, execlog)

	receipt.Ty = chain33types.ExecOk
	return receipt, nil
}

func (a *action) procMsgSetConsensusThreshold(msgSetConsensusThreshold *types2.MsgConsensusThreshold) (*chain33types.Receipt, error) {
	receipt := new(chain33types.Receipt)
	if !types2.CheckPower(msgSetConsensusThreshold.ConsensusThreshold) {
		return nil, types2.ErrInvalidPower
	}

	preConsensusNeeded, nowConsensusNeeded, err := a.keeper.ProcessSetConsensusNeeded(msgSetConsensusThreshold.ConsensusThreshold)
	if err != nil {
		return nil, err
	}

	setConsensusThreshold := &types2.ReceiptSetConsensusThreshold{
		PreConsensusThreshold: preConsensusNeeded,
		NowConsensusThreshold: nowConsensusNeeded,
		XTxHash:               a.txhash,
		XHeight:               uint64(a.height),
	}
	execlog := &chain33types.ReceiptLog{Ty: types2.TySetConsensusThresholdLog, Log: chain33types.Encode(setConsensusThreshold)}
	receipt.Logs = append(receipt.Logs, execlog)

	msgSetConsensusThresholdBytes, err := proto.Marshal(&types2.ReceiptQueryConsensusThreshold{
		ConsensusThreshold: nowConsensusNeeded,
	})
	if err != nil {
		return nil, chain33types.ErrMarshal
	}
	receipt.KV = append(receipt.KV, &chain33types.KeyValue{Key: types2.CalConsensusThresholdPrefix(), Value: msgSetConsensusThresholdBytes})

	receipt.Ty = chain33types.ExecOk
	return receipt, nil
}
