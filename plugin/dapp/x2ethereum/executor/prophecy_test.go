package executor

import (
	"testing"

	x2eTy "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
	"github.com/stretchr/testify/assert"
)

func TestAddClaim(t *testing.T) {
	valis := &x2eTy.StringMap{}
	valis.Validators = append(valis.Validators, "aa")
	validator := &x2eTy.ClaimValidators{Claim: "alice", Validators: valis}
	prophecy := &x2eTy.ReceiptEthProphecy{
		ClaimValidators: []*x2eTy.ClaimValidators{validator},
	}
	AddClaim(prophecy, "bb", "alice")
	assert.Contains(t, prophecy.ClaimValidators[0].Validators.Validators, "aa")
	assert.Contains(t, prophecy.ClaimValidators[0].Validators.Validators, "bb")

	valis2 := &x2eTy.StringMap{}
	valis2.Validators = append(valis.Validators, "zz")
	validator2 := &x2eTy.ClaimValidators{Claim: "bob", Validators: valis2}
	prophecy.ClaimValidators = append(prophecy.ClaimValidators, validator2)

	AddClaim(prophecy, "bb", "bob")
	assert.Contains(t, prophecy.ClaimValidators[0].Validators.Validators, "aa")
	assert.Contains(t, prophecy.ClaimValidators[0].Validators.Validators, "bb")
	assert.Contains(t, prophecy.ClaimValidators[1].Validators.Validators, "zz")
	assert.Contains(t, prophecy.ClaimValidators[1].Validators.Validators, "bb")

}

// TestCalcProphecyIDSeparator 确认分叉后 ID 带分隔符，不同(chainID,nonce)组合不再碰撞
func TestCalcProphecyIDSeparator(t *testing.T) {
	claimA := &x2eTy.Eth2Chain33{EthereumChainID: 1, Nonce: 23, EthereumSender: "0xsender", TokenContractAddress: "0xtoken", ClaimType: int64(x2eTy.LockClaimType)}
	claimB := &x2eTy.Eth2Chain33{EthereumChainID: 12, Nonce: 3, EthereumSender: "0xsender", TokenContractAddress: "0xtoken", ClaimType: int64(x2eTy.LockClaimType)}

	// 旧格式无分隔符，两种组合拼出相同 ID
	assert.Equal(t, calcProphecyID(claimA, false), calcProphecyID(claimB, false))
	// 分叉后带分隔符，ID 不再碰撞
	assert.NotEqual(t, calcProphecyID(claimA, true), calcProphecyID(claimB, true))
	assert.NotEqual(t, calcProphecyID(claimA, true), calcProphecyID(claimA, false))
}
