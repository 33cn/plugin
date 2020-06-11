package ethtxs

import (
	"fmt"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	chain33Common "github.com/33cn/chain33/common"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	chain33Types "github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	ebrelayerTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	chainTestCfg = chain33Types.NewChain33Config(chain33Types.GetDefaultCfgstring())
)

func Test_RelayToChain33(t *testing.T) {
	var tx chain33Types.Transaction
	var ret chain33Types.Reply
	ret.IsOk = true

	mockapi := &mocks.QueueProtocolAPI{}
	// 这里对需要mock的方法打桩,Close是必须的，其它方法根据需要
	mockapi.On("Close").Return()
	mockapi.On("AddPushSubscribe", mock.Anything).Return(&ret, nil)
	mockapi.On("CreateTransaction", mock.Anything).Return(&tx, nil)
	mockapi.On("SendTx", mock.Anything).Return(&ret, nil)
	mockapi.On("SendTransaction", mock.Anything).Return(&ret, nil)
	mockapi.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)

	mock33 := testnode.New("", mockapi)
	defer mock33.Close()
	rpcCfg := mock33.GetCfg().RPC
	// 这里必须设置监听端口，默认的是无效值
	rpcCfg.JrpcBindAddr = "127.0.0.1:8801"
	mock33.GetRPC().Listen()

	chain33PrivateKeyStr := "0xd627968e445f2a41c92173225791bae1ba42126ae96c32f28f97ff8f226e5c68"
	var driver secp256k1.Driver
	privateKeySli, err := chain33Common.FromHex(chain33PrivateKeyStr)
	require.Nil(t, err)

	priKey, err := driver.PrivKeyFromBytes(privateKeySli)
	require.Nil(t, err)

	claim := &ebrelayerTypes.EthBridgeClaim{}

	fmt.Println("======================= testRelayLockToChain33 =======================")
	_, err = RelayLockToChain33(priKey, claim, "http://127.0.0.1:8801")
	require.Nil(t, err)

	fmt.Println("======================= testRelayBurnToChain33 =======================")
	_, err = RelayBurnToChain33(priKey, claim, "http://127.0.0.1:8801")
	require.Nil(t, err)
}
