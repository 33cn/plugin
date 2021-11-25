package executor

import (
	"encoding/json"
	"errors"
	"math/big"

	dbm "github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/types"
	bridgevmxgo "github.com/33cn/plugin/plugin/dapp/bridgevmxgo/contracts/generated"
	chain33Abi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
)

const (
	LockMethod = "lock"
)

//solidity interface: function lock(address _recipient, address _token, uint256 _amount)
//铸币交易的接收人必须与发起lock交易时填写的接收地址一致
func checkMintPara(mint *evmxgotypes.EvmxgoMint, tx2lock *types.Transaction, db dbm.KV) error {
	bridgevmxgoConfig, err := loadBridgevmxgoAddr(db)
	if nil != err {
		return err
	}

	var action evmtypes.EVMContractAction
	if err := types.Decode(tx2lock.Payload, &action); nil != err {
		return err
	}
	//确认合约地址的正确性
	if action.ContractAddr != bridgevmxgoConfig.Address {
		return errors.New("Not consistent bridgevmxgo address configured by manager")
	}

	unpack, err := chain33Abi.UnpackAllTypes(action.Para, LockMethod, bridgevmxgo.BridgeBankABI)
	if err != nil {
		return err
	}

	var recipient, bridgeToken string
	var amount int64
	correct := 0
	for _, para := range unpack {
		switch para.Name {
		case "_recipient":
			recipient = para.Value.(string)
			if mint.Recipient != recipient {
				return errors.New("Not consitent recipient address")
			}
			correct++

		case "_amount":
			amount = para.Value.(*big.Int).Int64()
			if mint.Amount != amount {
				return errors.New("Not consitent Amount")
			}
			correct++

		case "_token":
			bridgeToken = para.Value.(string)
			if mint.BridgeToken != bridgeToken {
				return errors.New("Not consitent token Address")
			}
			correct++
		}
	}
	elog.Info("checkMintPara", "lock parameter unpacked _recipient ", recipient, "bridgeToken", bridgeToken,
		"amount", amount)
	if correct != 3 {
		return errors.New("not check all the points: _recipient, _amount, _token")
	}
	return nil
}

func loadBridgevmxgoAddr(db dbm.KV) (*evmxgotypes.BridgevmxgoConfig, error) {
	key := bridgevmxgoAddrPrefix
	value, err := getManageKey(key, db)
	if err != nil {
		elog.Info("loadBridgevmxgoAddr", "get db key", "not found", "key", key)
		return nil, err
	}
	if value == nil {
		elog.Info("loadBridgevmxgoAddr", "get db key", "  found nil value", "key", key)
		return nil, nil
	}
	elog.Info("loadBridgevmxgoAddr", "value", string(value))

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		elog.Error("loadBridgevmxgoAddr load loadEvmxgoMintConfig", "Can't decode ConfigItem due to", err.Error())
		return nil, err // types.ErrBadConfigValue
	}

	configValue := item.GetArr().Value
	if len(configValue) <= 0 {
		return nil, evmxgotypes.ErrEvmxgoSymbolNotConfigValue
	}

	var e evmxgotypes.BridgevmxgoConfig
	err = json.Unmarshal([]byte(configValue[0]), &e)

	if err != nil {
		elog.Error("loadBridgevmxgoAddr load", "Can't decode token info due to:", err.Error())
		return nil, err
	}
	return &e, nil
}
