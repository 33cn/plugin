package executor

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	vcomm "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"

	dbmock "github.com/33cn/chain33/common/db/mocks"
	ctypes "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/stretchr/testify/mock"

	"github.com/33cn/chain33/common"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"gotest.tools/assert"
)

const abiStr = "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"ApprovalForAll\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"values\",\"type\":\"uint256[]\"}],\"name\":\"TransferBatch\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"TransferSingle\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TransferToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"string\",\"name\":\"value\",\"type\":\"string\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"URI\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address[]\",\"name\":\"accounts\",\"type\":\"address[]\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"}],\"name\":\"balanceOfBatch\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"burnToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"tokenIds\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"}],\"name\":\"burnTokenBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"getTokenInfo\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"getUserTokenIds\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"isApprovedForAll\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"mintToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"tokenIds\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"}],\"name\":\"mintTokenBatch\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256[]\",\"name\":\"ids\",\"type\":\"uint256[]\"},{\"internalType\":\"uint256[]\",\"name\":\"amounts\",\"type\":\"uint256[]\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeBatchTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"id\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"safeTransferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"name\":\"setApprovalForAll\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"tokenInfo\",\"type\":\"string\"}],\"name\":\"setTokenInfo\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"transferToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"uri\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// mintTokenBatch(address owner, uint256[] calldata tokenIds, uint256[] calldata amounts)
func Test_Pack(t *testing.T) {
	_, packData, err := evmAbi.Pack("mintTokenBatch(14KEKbYtKKQm4wMthSK9J4La4nAiidGozt, [1,2,3], [100, 200, 300])", abiStr, false)
	assert.Equal(t, nil, err)
	packStr := common.ToHex(packData)
	fmt.Println("packStr = ", packStr)
	//0xa08fad67000000000000000000000000245afbf176934ccdd7ca291a8dddaa13c8184822000000000000000000000000000000000000000000000000000000000000006000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000000006400000000000000000000000000000000000000000000000000000000000000c8000000000000000000000000000000000000000000000000000000000000012c
}

func Test_Unpack(t *testing.T) {
	_, packData, err := evmAbi.Pack("mintTokenBatch(14KEKbYtKKQm4wMthSK9J4La4nAiidGozt, [1,2,3], [100, 200, 300])", abiStr, false)
	assert.Equal(t, nil, err)
	packStr := common.ToHex(packData)
	fmt.Println("packStr = ", packStr)

	outputs, err := evmAbi.Unpack(packData, "mintTokenBatch", abiStr)
	assert.Equal(t, nil, err)
	for i, info := range outputs {
		fmt.Println(i, "th info = ", info)
	}
}

func Test_UnpackEvent(t *testing.T) {

	//event TransferToken(address owner, address to, uint256 tokenId, uint256 amount);
	packData, err := common.FromHex("0x000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000c00000000000000000000000000000000000000000000000000000000000000003000000000000000000000000000000000000000000000000000000000098968100000000000000000000000000000000000000000000000000000000009896820000000000000000000000000000000000000000000000000000000000989683000000000000000000000000000000000000000000000000000000000000000300000000000000000000000000000000000000000000000000000000000000c800000000000000000000000000000000000000000000000000000000000000c800000000000000000000000000000000000000000000000000000000000000c8")
	assert.Equal(t, nil, err)

	outputs, err := evmAbi.UnpackAllTypes(packData, "TransferToken", abiStr)
	assert.Equal(t, nil, err)
	for i, info := range outputs {
		fmt.Println(i, "th info = ", info)
	}
}

const BridgeBankABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_operatorAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_oracleAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_ethereumBridgeAddress\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_beneficiary\",\"type\":\"address\"}],\"name\":\"LogBridgeTokenMint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_ownerFrom\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"_ethereumReceiver\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"LogEthereumTokenBurn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"_to\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"LogLock\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"LogNewBridgeToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"LogUnlock\",\"type\":\"event\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"addToken2LockList\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"bridgeTokenCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"bridgeTokenCreated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"bridgeTokenWhitelist\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_ethereumReceiver\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"_ethereumTokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"burnBridgeTokens\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_percents\",\"type\":\"uint8\"}],\"name\":\"configLockedTokenOfflineSave\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_offlineSave\",\"type\":\"address\"}],\"name\":\"configOfflineSaveAccount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"createNewBridgeToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"ethereumBridge\",\"outputs\":[{\"internalType\":\"contractEthereumBridge\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_id\",\"type\":\"bytes32\"}],\"name\":\"getEthereumDepositStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"getLockedTokenAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"getToken2address\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"getofflineSaveCfg\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"hasBridgeTokenCreated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"highThreshold\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_recipient\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"lock\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"lockedFunds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lowThreshold\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_ethereumSender\",\"type\":\"bytes\"},{\"internalType\":\"addresspayable\",\"name\":\"_intendedRecipient\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_bridgeTokenAddress\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"mintBridgeTokens\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"offlineSave\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"offlineSaveCfgs\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_percents\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"operator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"oracle\",\"outputs\":[{\"internalType\":\"contractOracle\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"token2address\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"tokenAllow2Lock\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_recipient\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"unlock\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_id\",\"type\":\"bytes32\"}],\"name\":\"viewEthereumDeposit\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

func Test_UnpackEventLock(t *testing.T) {

	//event TransferToken(address owner, address to, uint256 tokenId, uint256 amount);
	packData, err := common.FromHex("0x000000000000000000000000809b923eea05ad681a0ec1459e613cd7af7f5f1800000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000001dcd650000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000014bc333839e37bc7faad0137abae2275030555101f00000000000000000000000000000000000000000000000000000000000000000000000000000000000000034254590000000000000000000000000000000000000000000000000000000000")
	assert.Equal(t, nil, err)

	outputs, err := evmAbi.UnpackAllTypes(packData, "LogLock", BridgeBankABI)
	assert.Equal(t, nil, err)
	for i, info := range outputs {
		fmt.Println(i, "th info = ", info)
	}
}

//UnpackLogLock module=cross2eth_relayer
//value=500000000
//symbol=BTY
//token addr on chain33 evm=1111111111111111111114oLvT2
//chain33 sender=1Cj1rqUenPmkeD6A8MGEzkBKQFN2H9yL3x
//ethereum recipient=0xbc333839E37bc7fAAD0137aBaE2275030555101f
//nonce=1

const BridgeBankABIBridgevmxgo = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_operatorAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_oracleAddress\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_goAssetBridgeAddress\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_beneficiary\",\"type\":\"address\"}],\"name\":\"LogBridgeTokenMint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_ownerFrom\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_goAssetReceiver\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"LogGoAssetTokenBurn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"}],\"name\":\"LogLock\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"LogNewBridgeToken\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_value\",\"type\":\"uint256\"}],\"name\":\"LogUnlock\",\"type\":\"event\"},{\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"fallback\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"addToken2LockList\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"bridgeTokenCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"bridgeTokenCreated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"bridgeTokenWhitelist\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_goAssetReceiver\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_goAssetTokenAddress\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"burnBridgeTokens\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_percents\",\"type\":\"uint8\"}],\"name\":\"configLockedTokenOfflineSave\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_offlineSave\",\"type\":\"address\"}],\"name\":\"configOfflineSaveAccount\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"createNewBridgeToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_id\",\"type\":\"bytes32\"}],\"name\":\"getGoAssetDepositStatus\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"getLockedTokenAddress\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"getToken2address\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"getofflineSaveCfg\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"goAssetBridge\",\"outputs\":[{\"internalType\":\"contractGoAssetBridge\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"}],\"name\":\"hasBridgeTokenCreated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"highThreshold\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_recipient\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"lock\",\"outputs\":[],\"payable\":true,\"stateMutability\":\"payable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lockNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"lockedFunds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"lowThreshold\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"address\",\"name\":\"_goAssetSender\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"_intendedRecipient\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_bridgeTokenAddress\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"mintBridgeTokens\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"offlineSave\",\"outputs\":[{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"offlineSaveCfgs\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_threshold\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"_percents\",\"type\":\"uint8\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"operator\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"oracle\",\"outputs\":[{\"internalType\":\"contractOracle\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"token2address\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"tokenAllow2Lock\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"internalType\":\"addresspayable\",\"name\":\"_recipient\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"_symbol\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"unlock\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_id\",\"type\":\"bytes32\"}],\"name\":\"viewGoAssetDeposit\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"addresspayable\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"}]"

// 'lock(1BCGLhdcdthNutQowV2YShuuN9fJRRGLxu, 195ycckxqnxyaQeLYTQvgETr1vsEp4NYmP, 500000000)'
func Test_UnpackEventLockOfBridgevmxgo(t *testing.T) {
	//event TransferToken(address owner, address to, uint256 tokenId, uint256 amount);
	packData, err := common.FromHex("0x000000000000000000000000143b825fde7a42043b2cedef6b4c6f57c04d71b70000000000000000000000006fd2a3693c289b6c3c211f5d2f85ce145d1afaeb00000000000000000000000058b1edde0fc37c0f7f4a23d1c2df488ec5df0fe100000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000000000000000000000000000000000001dcd6500000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000034554480000000000000000000000000000000000000000000000000000000000")
	assert.Equal(t, nil, err)

	outputs, err := evmAbi.UnpackAllTypes(packData, "LogLock", BridgeBankABIBridgevmxgo)
	assert.Equal(t, nil, err)
	for i, info := range outputs {
		fmt.Println(i, "th info = ", info)
	}
}

// 'lock(1BCGLhdcdthNutQowV2YShuuN9fJRRGLxu, 195ycckxqnxyaQeLYTQvgETr1vsEp4NYmP, 500000000)'
// "7750c9f0": "lock(address,address,uint256)",
func Test_UnpackInputLockOfBridgevmxgo(t *testing.T) {
	methodName, packDataOrigin, err := evmAbi.Pack("lock(1BCGLhdcdthNutQowV2YShuuN9fJRRGLxu, 195ycckxqnxyaQeLYTQvgETr1vsEp4NYmP, 500000000)", BridgeBankABIBridgevmxgo, false)
	assert.Equal(t, nil, err)
	fmt.Println("methodName", methodName)
	fmt.Println("packDataOrigin", common.ToHex(packDataOrigin))

	//event TransferToken(address owner, address to, uint256 tokenId, uint256 amount);
	packData, err := common.FromHex("0x7750c9f00000000000000000000000006fd2a3693c289b6c3c211f5d2f85ce145d1afaeb00000000000000000000000058b1edde0fc37c0f7f4a23d1c2df488ec5df0fe1000000000000000000000000000000000000000000000000000000001dcd6500")
	//7750c9f0
	//0000000000000000000000006fd2a3693c289b6c3c211f5d2f85ce145d1afaeb
	//00000000000000000000000058b1edde0fc37c0f7f4a23d1c2df488ec5df0fe1
	//000000000000000000000000000000000000000000000000000000001dcd6500
	assert.Equal(t, nil, err)

	outputs, err := evmAbi.UnpackAllTypes(packData, "lock", BridgeBankABIBridgevmxgo)
	assert.Equal(t, nil, err)
	for i, info := range outputs {
		fmt.Println(i, "th info = ", info)
	}

	correct := 0
	for _, para := range outputs {
		switch para.Name {
		case "_recipient":
			assert.Equal(t, para.Value, "1BCGLhdcdthNutQowV2YShuuN9fJRRGLxu")
			correct++
		case "_amount":
			assert.Equal(t, para.Value.(*big.Int).Int64(), int64(500000000))
			correct++
		case "_token":
			assert.Equal(t, para.Value, "195ycckxqnxyaQeLYTQvgETr1vsEp4NYmP")
			correct++
		}
	}
	assert.Equal(t, correct, 3)
}

func TestEVMExecutor_Query_GetCode(t *testing.T) {
	api := new(apimock.QueueProtocolAPI)
	cfg := ctypes.NewChain33Config(ctypes.GetDefaultCfgstring())
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	exec := initEvmExeccutor(t, api)
	var contractorAddr = strings.ToLower("0xDe79A84DD3A16BB91044167075dE17a1CA4b1d6b")
	var creator = strings.ToLower("0xd83b69C56834E85e023B1738E69BFA2F0dd52905")
	exec.mStateDB.CreateAccount(contractorAddr, creator, "evm", "ABC")
	cacc := exec.mStateDB.GetAccount(contractorAddr)
	code := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8}
	exec.mStateDB.SetCode(contractorAddr, code)
	exec.mStateDB.SetAbi(contractorAddr, "abidata")
	t.Log("data", cacc.Data.GetCreator(), "abi", exec.mStateDB.GetAbi(contractorAddr))
	kv := cacc.GetDataKV()
	exec.GetMStateDB().StateDB.Set(kv[0].GetKey(), kv[0].GetValue())
	msg, err := exec.Query_GetCode(&types.CheckEVMAddrReq{Addr: contractorAddr})
	assert.Equal(t, nil, err)
	ret, ok := msg.(*types.EVMContractData)
	assert.Equal(t, true, ok)
	t.Log("ret", ret.Code)
	ok = bytes.Equal(code, ret.Code)
	assert.Equal(t, true, ok)
}

func TestNewEVMExecutor_Query_GetNonce(t *testing.T) {
	api := new(apimock.QueueProtocolAPI)
	cfg := ctypes.NewChain33Config(ctypes.GetDefaultCfgstring())
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	exec := initEvmExeccutor(t, api)
	localDB := new(dbmock.KVDB)
	exec.SetLocalDB(localDB)
	var creator = "0xd83b69C56834E85e023B1738E69BFA2F0dd52905"
	var evmAccountNonce ctypes.EvmAccountNonce
	evmAccountNonce.Nonce = 123
	evmAccountNonce.Addr = creator
	localDB.On("Get", mock.Anything).Return(ctypes.Encode(&evmAccountNonce), nil)
	msg, err := exec.Query_GetNonce(&types.EvmGetNonceReq{
		Address: creator,
	})
	assert.Equal(t, nil, err)
	resp, ok := msg.(*types.EvmGetNonceRespose)
	assert.Equal(t, true, ok)
	t.Log("getnonce", resp.GetNonce())
	assert.Equal(t, evmAccountNonce.Nonce, resp.GetNonce())
}
func initEvmExeccutor(t *testing.T, api *apimock.QueueProtocolAPI) *EVMExecutor {
	localDB := new(dbmock.KVDB)
	driver, err := address.LoadDriver(2, -1)
	assert.Equal(t, nil, err)

	vcomm.InitEvmAddressDriver(driver)
	var exec = NewEVMExecutor()
	exec.SetAPI(api)
	statDB, err := dbm.NewGoMemDB("state", "state", 1024)
	assert.Equal(t, nil, err)
	exec.SetLocalDB(localDB)
	exec.SetStateDB(statDB)
	exec.SetEnv(0, 0, 0)
	exec.CheckInit()
	exec.mStateDB.StateDB = statDB

	return exec
}

func TestEVMExecutor_Check(t *testing.T) {
	api := new(apimock.QueueProtocolAPI)

	cfg := ctypes.NewChain33Config(ctypes.GetDefaultCfgstring())
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	pub := vcomm.FromHex("0x04715e4e07d983c2d98eeac7018bce6e68ef9de25835340f6455f1b1c9686132ac54904f5e04b07966a256140a5f487c4aef3ddc461e02d58f90cc8baa49f9c7ca")
	sig := &ctypes.Signature{
		Ty:     8452,
		Pubkey: pub,
	}
	toAddr := "0x55d398326f99059fF775485246999027B3197955"
	v := &cty.CoinsAction_Transfer{Transfer: &ctypes.AssetsTransfer{Amount: 111}}
	transfer := &cty.CoinsAction{Value: v, Ty: cty.CoinsActionTransfer}
	//相同的NONCE，tx2手续费更高
	tx1 := &ctypes.Transaction{ChainID: 3999, Execer: []byte("coins"), Payload: ctypes.Encode(transfer), Fee: 100000, Expire: 0, To: toAddr, Nonce: 1, Signature: sig}
	tx2 := &ctypes.Transaction{ChainID: 3999, Execer: []byte("coins"), Payload: ctypes.Encode(transfer), Fee: 110002, Expire: 0, To: toAddr, Nonce: 1, Signature: sig}
	tx3 := &ctypes.Transaction{ChainID: 3999, Execer: []byte("coins"), Payload: ctypes.Encode(transfer), Fee: 100003, Expire: 0, To: toAddr, Nonce: 3, Signature: sig}
	tx4 := &ctypes.Transaction{ChainID: 3999, Execer: []byte("coins"), Payload: ctypes.Encode(transfer), Fee: 100003, Expire: 0, To: toAddr, Nonce: 4, Signature: sig}

	var details []*ctypes.TransactionDetail
	details = append(details, &ctypes.TransactionDetail{Tx: tx1, Index: 0}, &ctypes.TransactionDetail{Tx: tx2, Index: 1})
	detailTxs := &ctypes.TransactionDetails{Txs: details}
	exec := initEvmExeccutor(t, api)

	api.On("GetTxListByAddr", mock.Anything).Return(detailTxs, nil)
	api.On("RemoveTxsByHashList", mock.Anything).Return(nil)
	err := exec.CheckTx(tx2, 0)
	assert.Equal(t, nil, err)

	detailTxs.Txs = append(detailTxs.Txs, &ctypes.TransactionDetail{Tx: tx3, Index: 2})
	err = exec.CheckTx(nil, 0)
	assert.Equal(t, err.Error(), "tx empty")
	detailTxs.Txs = append(detailTxs.Txs, &ctypes.TransactionDetail{Tx: tx4, Index: 3})
	err = exec.CheckTx(tx4, 0)
	assert.Equal(t, nil, err)

}
