package executor

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/executor/oracle"
	types2 "github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

func (x *x2ethereum) Query_GetEthProphecy(in *types2.QueryEthProphecyParams) (types.Message, error) {
	prophecy := &types2.ReceiptEthProphecy{}
	prophecyKey := types2.CalProphecyPrefix()

	var dbProphecy []oracle.DBProphecy
	val, err := x.GetStateDB().Get(prophecyKey)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(val, &dbProphecy)
	if err != nil {
		return nil, types.ErrUnmarshal
	}

	for _, dbP := range dbProphecy {
		if dbP.ID == in.ID {
			dbPD, err := dbP.DeserializeFromDB()
			if err != nil {
				return nil, err
			}
			prophecy = &types2.ReceiptEthProphecy{
				ID: in.ID,
				Status: &types2.ProphecyStatus{
					Text:       types2.EthBridgeStatus(dbP.Status.Text),
					FinalClaim: dbP.Status.FinalClaim,
				},
				ClaimValidators: dbPD.ClaimValidators,
				ValidatorClaims: dbPD.ValidatorClaims,
			}
			return prophecy, nil
		}
	}
	return nil, types2.ErrInvalidProphecyID
}

func (x *x2ethereum) Query_GetValidators(in *types2.QueryValidatorsParams) (types.Message, error) {
	validatorsKey := types2.CalValidatorMapsPrefix()

	var v []*types2.MsgValidator
	vBytes, err := x.GetStateDB().Get(validatorsKey)
	if err != nil {
		elog.Error("Query_GetValidators", "GetValidators Err", err)
		return nil, err
	}

	err = json.Unmarshal(vBytes, &v)
	if err != nil {
		return nil, types.ErrUnmarshal
	}

	if in.Validator != "" {
		validatorsRes := new(types2.ReceiptQueryValidator)
		for _, vv := range v {
			if vv.Address == in.Validator {
				val := make([]*types2.MsgValidator, 1)
				val[0] = vv
				validatorsRes = &types2.ReceiptQueryValidator{
					Validators: val,
					TotalPower: vv.Power,
				}
				return validatorsRes, nil
			}
		}
		// 未知的地址
		return nil, types2.ErrInvalidValidator
	} else {
		validatorsRes := new(types2.ReceiptQueryValidator)
		var totalPower int64
		for _, vv := range v {
			totalPower += vv.Power
		}
		validatorsRes.Validators = v
		validatorsRes.TotalPower = totalPower
		return validatorsRes, nil
	}
}

func (x *x2ethereum) Query_GetTotalPower(in *types2.QueryTotalPowerParams) (types.Message, error) {
	totalPower := &types2.ReceiptQueryTotalPower{}
	totalPowerKey := types2.CalLastTotalPowerPrefix()

	totalPowerBytes, err := x.GetStateDB().Get(totalPowerKey)
	if err != nil {
		elog.Error("Query_GetTotalPower", "GetTotalPower Err", err)
		return nil, err
	}
	err = json.Unmarshal(totalPowerBytes, &totalPower)
	if err != nil {
		return nil, types.ErrUnmarshal
	}
	return totalPower, nil
}

func (x *x2ethereum) Query_GetConsensusThreshold(in *types2.QueryConsensusThresholdParams) (types.Message, error) {
	consensus := &types2.ReceiptSetConsensusThreshold{}
	consensusKey := types2.CalConsensusThresholdPrefix()

	consensusBytes, err := x.GetStateDB().Get(consensusKey)
	if err != nil {
		elog.Error("Query_GetConsensusNeeded", "GetConsensusNeeded Err", err)
		return nil, err
	}
	err = json.Unmarshal(consensusBytes, &consensus)
	if err != nil {
		return nil, types.ErrUnmarshal
	}
	return consensus, nil
}

func (x *x2ethereum) Query_GetSymbolTotalAmountByTxType(in *types2.QuerySymbolAssetsByTxTypeParams) (types.Message, error) {
	symbolAmount := &types2.ReceiptQuerySymbolAssets{}

	if in.TokenAddr != "" {
		var r types2.ReceiptQuerySymbolAssetsByTxType
		symbolAmountKey := types2.CalTokenSymbolTotalLockOrBurnAmount(in.TokenSymbol, in.TokenAddr, types2.DirectionType[in.Direction], in.TxType)

		totalAmountBytes, err := x.GetLocalDB().Get(symbolAmountKey)
		if err != nil {
			elog.Error("Query_GetSymbolTotalAmountByTxType", "GetSymbolTotalAmountByTxType Err", err)
			return nil, err
		}
		err = types.Decode(totalAmountBytes, &r)
		if err != nil {
			return nil, types.ErrUnmarshal
		}

		r.TotalAmount = types2.TrimZeroAndDot(strconv.FormatFloat(types2.Toeth(r.TotalAmount, in.Decimal), 'f', 4, 64))

		symbolAmount.Res = append(symbolAmount.Res, &r)
	} else {
		tokenAddressesBytes, err := x.GetLocalDB().Get(types2.CalTokenSymbolToTokenAddress(in.TokenSymbol))
		if err != nil && err != types.ErrNotFound {
			return nil, err
		}
		var tokenAddresses types2.ReceiptTokenToTokenAddress
		err = types.Decode(tokenAddressesBytes, &tokenAddresses)
		if err != nil {
			return nil, err
		}

		for _, addr := range tokenAddresses.TokenAddress {
			var r types2.ReceiptQuerySymbolAssetsByTxType
			symbolAmountKey := types2.CalTokenSymbolTotalLockOrBurnAmount(in.TokenSymbol, addr, types2.DirectionType[in.Direction], in.TxType)

			totalAmountBytes, err := x.GetLocalDB().Get(symbolAmountKey)
			if err != nil {
				elog.Error("Query_GetSymbolTotalAmountByTxType", "GetSymbolTotalAmountByTxType Err", err)
				return nil, err
			}
			err = types.Decode(totalAmountBytes, &r)
			if err != nil {
				return nil, types.ErrUnmarshal
			}

			r.TotalAmount = types2.TrimZeroAndDot(strconv.FormatFloat(types2.Toeth(r.TotalAmount, in.Decimal), 'f', 4, 64))

			symbolAmount.Res = append(symbolAmount.Res, &r)
		}
	}

	return symbolAmount, nil
}

func (x *x2ethereum) Query_GetRelayerBalance(in *types2.QueryRelayerBalance) (types.Message, error) {
	symbolAmount := &types2.ReceiptQueryRelayerBalance{}

	// 要查询特定的tokenAddr
	if in.TokenAddr != "" {
		accDB, err := account.NewAccountDB(x.GetAPI().GetConfig(), types2.X2ethereumX, strings.ToLower(in.TokenSymbol+in.TokenAddr), x.GetStateDB())
		if err != nil {
			return nil, err
		}

		acc := accDB.LoadExecAccount(in.Address, address.ExecAddress(types2.X2ethereumX))
		res := new(types2.ReceiptQueryRelayerBalanceForOneToken)
		res.TokenAddr = in.TokenAddr
		res.TokenSymbol = in.TokenSymbol
		res.Balance = types2.TrimZeroAndDot(strconv.FormatFloat(float64(acc.Balance)/1e8, 'f', 4, 64))
		symbolAmount.Res = append(symbolAmount.Res, res)

	} else {

		tokenAddressesBytes, err := x.GetLocalDB().Get(types2.CalTokenSymbolToTokenAddress(in.TokenSymbol))
		if err != nil && err != types.ErrNotFound {
			return nil, err
		}
		var tokenAddresses types2.ReceiptTokenToTokenAddress
		err = types.Decode(tokenAddressesBytes, &tokenAddresses)
		if err != nil {
			return nil, err
		}

		for _, addr := range tokenAddresses.TokenAddress {
			accDB, err := account.NewAccountDB(x.GetAPI().GetConfig(), types2.X2ethereumX, strings.ToLower(in.TokenSymbol+addr), x.GetStateDB())
			if err != nil {
				return nil, err
			}

			acc := accDB.LoadExecAccount(in.Address, address.ExecAddress(types2.X2ethereumX))
			res := new(types2.ReceiptQueryRelayerBalanceForOneToken)
			res.TokenAddr = addr
			res.TokenSymbol = in.TokenSymbol
			res.Balance = types2.TrimZeroAndDot(strconv.FormatFloat(float64(acc.Balance)/1e8, 'f', 4, 64))
			symbolAmount.Res = append(symbolAmount.Res, res)
		}
	}

	return symbolAmount, nil
}
