package rollup

import (
	"encoding/hex"
	"testing"

	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/stretchr/testify/require"

	_ "github.com/33cn/chain33/system/dapp/init"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/crypto/bls"
)

func TestRollup(t *testing.T) {

	cfg := types.NewChain33Config(types.GetDefaultCfgstring())

	blsDrv := &bls.Driver{}

	//
	//for i:=0;i<100;i++{
	priv, _ := blsDrv.GenKey()

	println("blsPriv", hex.EncodeToString(priv.Bytes()))
	println("blsPub", hex.EncodeToString(priv.PubKey().Bytes()))
	addr, priv1 := util.Genaddress()

	println("secpPriv", hex.EncodeToString(priv1.Bytes()), addr)

	//pk := priv.PubKey()
	_, priv = util.Genaddress()
	tx := util.CreateCoinsTx(cfg, priv, addr, 1)

	//}
	sign := tx.GetSignature()
	tx.Signature = nil
	println("txsize", tx.Size(), len(sign.Pubkey), len(sign.GetSignature()))

}

func TestCommitCache(t *testing.T) {

	c := newCommitCache(1)

	cp := &rtypes.CheckPoint{CommitRound: 10}
	c.addCommitInfo(&commitInfo{cp: cp})
	cp.CommitRound = 11
	c.addCommitInfo(&commitInfo{cp: cp})
	require.Equal(t, 2, len(c.commitList))
	require.Equal(t, int64(11), c.getCheckPoint(11).CommitRound)

	sign := &rtypes.ValidatorSignMsg{
		CommitRound: 10,
		PubKey:      []byte("test"),
	}
	c.addValidatorSign(true, sign)
	c.addValidatorSign(false, sign)
	c.addValidatorSign(false, sign)

	sign.CommitRound = 11
	c.addValidatorSign(false, sign)
	c.addValidatorSign(true, sign)
	c.addValidatorSign(false, &rtypes.ValidatorSignMsg{CommitRound: 11})

	require.Equal(t, 2, len(c.signList))
	require.Equal(t, 1, len(c.signList[10].others))
	require.Equal(t, 2, len(c.signList[11].others))

	testAggre := func(set *validatorSignMsgSet) (pubs [][]byte, aggreSign []byte) {
		return make([][]byte, 1), nil
	}
	info := c.getPreparedCommit(11, testAggre)
	require.Equal(t, int64(11), info.cp.CommitRound)

	c.cleanHistory(20)
	require.Equal(t, 1, len(c.commitList))
	require.Equal(t, 1, len(c.signList))
}
