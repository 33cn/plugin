package types

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
)

// TestEvmType_CreateTx 测试RPC创建交易逻辑
func TestEvmType_CreateTx(t *testing.T) {
	evm := &EvmType{}
	errMap := map[int]string{2: "code must be set in create tx",
		4: "encoding/hex: invalid byte: U+0078 'x'"}
	for idx, test := range []CreateCallTx{
		{
			Code:     "abddee",
			Abi:      "[{}]",
			IsCreate: true,
			Name:     "user.evm.xxx",
			Note:     "test",
			Alias:    "mycon",
			Fee:      5000000,
			Amount:   100000000,
		},
		{
			Code:     "abddee",
			Abi:      "",
			IsCreate: true,
			Name:     "user.evm.xxx",
			Note:     "test",
			Alias:    "mycon",
			Fee:      5000000,
			Amount:   100000000,
		},
		{
			Code:     "",
			Abi:      "[{}]",
			IsCreate: true,
			Name:     "user.evm.xxx",
			Note:     "test",
			Alias:    "mycon",
			Fee:      5000000,
			Amount:   100000000,
		},
		{
			Code:     "abccdd",
			Abi:      "[{}]",
			IsCreate: false,
			Name:     "user.evm.xxx",
			Note:     "test",
			Alias:    "mycon",
			Fee:      0,
			Amount:   100000000,
		},
		{
			Code:     "xyz",
			Abi:      "[{}]",
			IsCreate: true,
			Name:     "user.evm.xxx",
			Note:     "test",
			Alias:    "mycon",
			Fee:      5000000,
			Amount:   100000000,
		},
	} {

		data, err := json.Marshal(&test)
		assert.NoError(t, err)

		tx, err := evm.CreateTx("CreateCall", data)
		if er, ok := errMap[idx]; ok {
			assert.EqualError(t, err, er)
			continue
		} else {
			assert.NoError(t, err)
		}

		var action EVMContractAction
		types.Decode(tx.Payload, &action)

		assert.EqualValues(t, test.Amount, action.Amount)
		assert.EqualValues(t, test.Abi, action.Abi)
		assert.EqualValues(t, test.Alias, action.Alias)
		assert.EqualValues(t, test.Note, action.Note)

		if tx.Fee < test.Fee {
			assert.Fail(t, "tx fee low")
		}

		if len(test.Code) > 0 {
			bcode, err := hex.DecodeString(test.Code)
			assert.NoError(t, err)
			assert.EqualValues(t, bcode, action.Code)
		}
		if test.IsCreate {
			assert.EqualValues(t, address.ExecAddress(types.ExecName(ExecutorName)), tx.To)
		} else {
			assert.EqualValues(t, address.ExecAddress(types.ExecName(test.Name)), tx.To)
		}
	}

}
