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
