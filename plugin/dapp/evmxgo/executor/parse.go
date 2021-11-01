package executor

import (
	"errors"

	"github.com/33cn/chain33/types"
	bridgevmxgo "github.com/33cn/plugin/plugin/dapp/bridgevmxgo/contracts/generated"
	chain33Abi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
)

const (
	LockMethod = "lock"
)

//solidity interface: function lock(address _recipient, address _token, uint256 _amount)
//铸币交易的接收人必须与发起lock交易时填写的接收地址一致
func checkMinePara(mint *evmxgotypes.EvmxgoMint, tx2lock *types.Transaction) error {
	unpack, err := chain33Abi.Unpack(tx2lock.Payload, LockMethod, bridgevmxgo.BridgeBankABI)
	if err != nil {
		return err
	}
	for _, para := range unpack {
		switch para.Name {
		case "_recipient":
			if mint.Recipient != para.Value {
				return errors.New("Not consitent recipient address")
			}
		case "_amount":
			if mint.Amount != para.Value {
				return errors.New("Not consitent Amount")
			}

		case "_token":
			if mint.BridgeToken != para.Value {
				return errors.New("Not consitent token Address")
			}
		}
	}
	return nil
}
