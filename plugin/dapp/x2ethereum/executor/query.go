package executor

import (
	"strconv"
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/types"
	x2eTy "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
)

func (x *x2ethereum) Query_GetEthProphecy(in *x2eTy.QueryEthProphecyParams) (types.Message, error) {
	prophecyKey := x2eTy.CalProphecyPrefix(in.ID)

	var dbProphecy x2eTy.ReceiptEthProphecy

	val, err := x.GetStateDB().Get(prophecyKey)
	if err != nil {
		return nil, err
	}

	err = types.Decode(val, &dbProphecy)
	if err != nil {
		return nil, types.ErrUnmarshal
	}

	return &dbProphecy, nil
}

func (x *x2ethereum) Query_GetValidators(in *x2eTy.QueryValidatorsParams) (types.Message, error) {
	validatorsKey := x2eTy.CalValidatorMapsPrefix()

	var v x2eTy.ValidatorList
	vBytes, err := x.GetStateDB().Get(validatorsKey)
	if err != nil {
		elog.Error("Query_GetValidators", "GetValidators Err", err)
		return nil, err
	}

	err = types.Decode(vBytes, &v)
	if err != nil {
		return nil, types.ErrUnmarshal
	}

	if in.Validator != "" {
		for _, vv := range v.Validators {
			if vv.Address == in.Validator {
				return &x2eTy.ReceiptQueryValidator{
					Validators: []*x2eTy.MsgValidator{vv},
					TotalPower: vv.Power,
				}, nil
			}
		}
		// 未知的地址
		return nil, x2eTy.ErrInvalidValidator
	}

	validatorsRes := new(x2eTy.ReceiptQueryValidator)
	var totalPower int64
	for _, vv := range v.Validators {
		totalPower += vv.Power
	}
	validatorsRes.Validators = v.Validators
	validatorsRes.TotalPower = totalPower
	return validatorsRes, nil

}

func (x *x2ethereum) Query_GetTotalPower(in *x2eTy.QueryTotalPowerParams) (types.Message, error) {
	totalPower := &x2eTy.ReceiptQueryTotalPower{}
	totalPowerKey := x2eTy.CalLastTotalPowerPrefix()

	totalPowerBytes, err := x.GetStateDB().Get(totalPowerKey)
	if err != nil {
		elog.Error("Query_GetTotalPower", "GetTotalPower Err", err)
		return nil, err
	}
	err = types.Decode(totalPowerBytes, totalPower)
	if err != nil {
		return nil, types.ErrUnmarshal
	}
	return totalPower, nil
}

func (x *x2ethereum) Query_GetConsensusThreshold(in *x2eTy.QueryConsensusThresholdParams) (types.Message, error) {
	consensus := &x2eTy.ReceiptQueryConsensusThreshold{}
	consensusKey := x2eTy.CalConsensusThresholdPrefix()

	consensusBytes, err := x.GetStateDB().Get(consensusKey)
	if err != nil {
		elog.Error("Query_GetConsensusNeeded", "GetConsensusNeeded Err", err)
		return nil, err
	}
	err = types.Decode(consensusBytes, consensus)
	if err != nil {
		return nil, types.ErrUnmarshal
	}
	return consensus, nil
}

func (x *x2ethereum) Query_GetSymbolTotalAmountByTxType(in *x2eTy.QuerySymbolAssetsByTxTypeParams) (types.Message, error) {
	symbolAmount := &x2eTy.ReceiptQuerySymbolAssets{}

	if in.TokenAddr != "" {
		var r x2eTy.ReceiptQuerySymbolAssetsByTxType
		symbolAmountKey := x2eTy.CalTokenSymbolTotalLockOrBurnAmount(in.TokenSymbol, in.TokenAddr, x2eTy.DirectionType[in.Direction], in.TxType)

		totalAmountBytes, err := x.GetLocalDB().Get(symbolAmountKey)
		if err != nil {
			elog.Error("Query_GetSymbolTotalAmountByTxType", "GetSymbolTotalAmountByTxType Err", err)
			return nil, err
		}
		err = types.Decode(totalAmountBytes, &r)
		if err != nil {
			return nil, types.ErrUnmarshal
		}

		r.TotalAmount = x2eTy.TrimZeroAndDot(strconv.FormatFloat(x2eTy.Toeth(r.TotalAmount, in.Decimal), 'f', 4, 64))

		symbolAmount.Res = append(symbolAmount.Res, &r)
	} else {
		tokenAddressesBytes, err := x.GetLocalDB().Get(x2eTy.CalTokenSymbolToTokenAddress(in.TokenSymbol))
		if err != nil && err != types.ErrNotFound {
			return nil, err
		}
		var tokenAddresses x2eTy.ReceiptTokenToTokenAddress
		err = types.Decode(tokenAddressesBytes, &tokenAddresses)
		if err != nil {
			return nil, err
		}

		for _, addr := range tokenAddresses.TokenAddress {
			var r x2eTy.ReceiptQuerySymbolAssetsByTxType
			symbolAmountKey := x2eTy.CalTokenSymbolTotalLockOrBurnAmount(in.TokenSymbol, addr, x2eTy.DirectionType[in.Direction], in.TxType)

			totalAmountBytes, err := x.GetLocalDB().Get(symbolAmountKey)
			if err != nil {
				elog.Error("Query_GetSymbolTotalAmountByTxType", "GetSymbolTotalAmountByTxType Err", err)
				return nil, err
			}
			err = types.Decode(totalAmountBytes, &r)
			if err != nil {
				return nil, types.ErrUnmarshal
			}

			r.TotalAmount = x2eTy.TrimZeroAndDot(strconv.FormatFloat(x2eTy.Toeth(r.TotalAmount, in.Decimal), 'f', 4, 64))

			symbolAmount.Res = append(symbolAmount.Res, &r)
		}
	}

	return symbolAmount, nil
}

func (x *x2ethereum) Query_GetRelayerBalance(in *x2eTy.QueryRelayerBalance) (types.Message, error) {
	symbolAmount := &x2eTy.ReceiptQueryRelayerBalance{}

	// 要查询特定的tokenAddr
	if in.TokenAddr != "" {
		accDB, err := account.NewAccountDB(x.GetAPI().GetConfig(), x2eTy.X2ethereumX, strings.ToLower(in.TokenSymbol+in.TokenAddr), x.GetStateDB())
		if err != nil {
			return nil, err
		}

		acc := accDB.LoadAccount(in.Address)
		res := new(x2eTy.ReceiptQueryRelayerBalanceForOneToken)
		res.TokenAddr = in.TokenAddr
		res.TokenSymbol = in.TokenSymbol
		res.Balance = x2eTy.TrimZeroAndDot(strconv.FormatFloat(float64(acc.Balance)/1e8, 'f', 4, 64))
		symbolAmount.Res = append(symbolAmount.Res, res)
	} else {
		tokenAddressesBytes, err := x.GetLocalDB().Get(x2eTy.CalTokenSymbolToTokenAddress(in.TokenSymbol))
		if err != nil && err != types.ErrNotFound {
			return nil, err
		}
		var tokenAddresses x2eTy.ReceiptTokenToTokenAddress
		err = types.Decode(tokenAddressesBytes, &tokenAddresses)
		if err != nil {
			return nil, err
		}

		for _, addr := range tokenAddresses.TokenAddress {
			accDB, err := account.NewAccountDB(x.GetAPI().GetConfig(), x2eTy.X2ethereumX, strings.ToLower(in.TokenSymbol+addr), x.GetStateDB())
			if err != nil {
				return nil, err
			}

			acc := accDB.LoadAccount(in.Address)
			res := new(x2eTy.ReceiptQueryRelayerBalanceForOneToken)
			res.TokenAddr = addr
			res.TokenSymbol = in.TokenSymbol
			res.Balance = x2eTy.TrimZeroAndDot(strconv.FormatFloat(float64(acc.Balance)/1e8, 'f', 4, 64))
			symbolAmount.Res = append(symbolAmount.Res, res)
		}
	}

	return symbolAmount, nil
}
