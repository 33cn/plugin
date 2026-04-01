package runtime

import (
	"bytes"
	"math/big"
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	vmcommon "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/state"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/require"
)

func TestCallContractEmitsLogWithCalleeAddress(t *testing.T) {
	cfg := ctypes.NewChain33Config(ctypes.GetDefaultCfgstring())
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig").Return(cfg)

	dbDir, stateDB, localDB := util.CreateTestDB()
	defer util.CloseTestDB(dbDir, stateDB)

	// Use a height above ForkEVMState so TyLogEVMEventData is exported from snapshots.
	const blockHeight = int64(700000)
	mdb := state.NewMemoryStateDB(stateDB, localDB, nil, blockHeight, api)
	txHash := vmcommon.BytesToHash([]byte("tx-a-call-b-log"))
	mdb.Prepare(txHash, 0)

	callerAddr := vmcommon.BytesToAddress([]byte{0x01})
	contractA := vmcommon.BytesToAddress([]byte{0xa1})
	contractB := vmcommon.BytesToAddress([]byte{0xb1})
	logTopic := vmcommon.BytesToHash([]byte("callee-b-topic"))

	mdb.CreateAccount(contractA.String(), callerAddr.String(), "evm.A", "A")
	mdb.CreateAccount(contractB.String(), callerAddr.String(), "evm.B", "B")
	mdb.SetCode(contractA.String(), buildCallCode(contractB))
	mdb.SetCode(contractB.String(), buildEmitLogCode(logTopic))

	evm := NewEVM(Context{
		CanTransfer: func(state.EVMStateDB, vmcommon.Address, uint64) bool { return true },
		Transfer:    func(state.EVMStateDB, vmcommon.Address, vmcommon.Address, uint64) bool { return true },
		GetHash:     func(uint64) vmcommon.Hash { return vmcommon.Hash{} },
		BlockNumber: big.NewInt(blockHeight),
	}, mdb, Config{}, cfg)

	var err error
	_, _, _, err = evm.Call(AccountRef(callerAddr), contractA, nil, 5_000_000, 0)
	require.NoError(t, err)

	ver := mdb.GetLastSnapshot()
	require.NotNil(t, ver)
	_, logs := mdb.GetChangedData(ver.GetID())

	found := false
	for _, lg := range logs {
		if lg.Ty != evmtypes.TyLogEVMEventData {
			continue
		}
		var evmLog ctypes.EVMLog
		err = ctypes.Decode(lg.Log, &evmLog)
		require.NoError(t, err)

		if len(evmLog.GetTopic()) == 1 && bytes.Equal(evmLog.GetTopic()[0], logTopic.Bytes()) {
			found = true
			require.Equal(t, contractB.String(), evmLog.GetAddress())
		}
	}
	require.True(t, found, "expected at least one event log emitted by contract B")
}

func buildCallCode(callee vmcommon.Address) []byte {
	code := []byte{
		0x60, 0x00, // retSize
		0x60, 0x00, // retOffset
		0x60, 0x00, // inSize
		0x60, 0x00, // inOffset
		0x60, 0x00, // value
		0x73, // PUSH20 <callee>
	}
	code = append(code, callee.Bytes()...)
	code = append(code,
		0x61, 0xff, 0xff, // gas
		0xf1, // CALL
		0x00, // STOP
	)
	return code
}

func buildEmitLogCode(topic vmcommon.Hash) []byte {
	code := []byte{0x7f} // PUSH32 <topic>
	code = append(code, topic.Bytes()...)
	code = append(code,
		0x60, 0x00, // mSize
		0x60, 0x00, // mStart
		0xa1, // LOG1
		0x00, // STOP
	)
	return code
}
