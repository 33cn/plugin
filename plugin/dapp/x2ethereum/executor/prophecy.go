package executor

import (
	"fmt"
	"strconv"

	"github.com/33cn/chain33/types"
	x2eTy "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
)

//NewProphecy ...
func NewProphecy(id string) *x2eTy.ReceiptEthProphecy {

	status := new(x2eTy.ProphecyStatus)
	status.Text = x2eTy.EthBridgeStatus_PendingStatusText

	return &x2eTy.ReceiptEthProphecy{
		ID:              id,
		Status:          status,
		ClaimValidators: *new([]*x2eTy.ClaimValidators),
		ValidatorClaims: *new([]*x2eTy.ValidatorClaims),
	}
}

// AddClaim adds a given claim to this prophecy
func AddClaim(prophecy *x2eTy.ReceiptEthProphecy, validator string, claim string) {
	if len(prophecy.ClaimValidators) == 0 {
		prophecy.ClaimValidators = append(prophecy.ClaimValidators, &x2eTy.ClaimValidators{
			Claim: claim,
			Validators: &x2eTy.StringMap{
				Validators: []string{validator},
			},
		})
	} else {
		for _, cv := range prophecy.ClaimValidators {
			if cv.Claim == claim {
				cv.Validators.Validators = append(cv.Validators.Validators, validator)
				break
			}
		}
	}

	prophecy.ValidatorClaims = append(prophecy.ValidatorClaims, &x2eTy.ValidatorClaims{
		Validator: validator,
		Claim:     claim,
	})

}

//FindHighestClaim 遍历该prophecy所有claim，找出获得最多票数的claim
func FindHighestClaim(prophecy *x2eTy.ReceiptEthProphecy, validators map[string]int64) (string, int64, int64) {
	totalClaimsPower := int64(0)
	highestClaimPower := int64(-1)
	highestClaim := ""
	for _, claimValidators := range prophecy.ClaimValidators {
		claimPower := int64(0)
		for _, validatorAddr := range claimValidators.Validators.Validators {
			validatorPower := validators[validatorAddr]
			claimPower += validatorPower
		}
		totalClaimsPower += claimPower
		if claimPower > highestClaimPower {
			highestClaimPower = claimPower
			highestClaim = claimValidators.Claim
		}
	}
	return highestClaim, highestClaimPower, totalClaimsPower
}

//NewOracleClaimContent ...
func NewOracleClaimContent(chain33Receiver string, amount string, claimType, decimals int64) x2eTy.OracleClaimContent {
	return x2eTy.OracleClaimContent{
		Chain33Receiver: chain33Receiver,
		Amount:          amount,
		ClaimType:       claimType,
		Decimals:        decimals,
	}
}

//NewClaim ...
func NewClaim(id string, validatorAddress string, content string) x2eTy.OracleClaim {
	return x2eTy.OracleClaim{
		ID:               id,
		ValidatorAddress: validatorAddress,
		Content:          content,
	}
}

//CreateOracleClaimFromEthClaim 通过ethchain33结构构造一个OracleClaim结构，包括生成唯一的ID
func CreateOracleClaimFromEthClaim(ethClaim x2eTy.Eth2Chain33) (x2eTy.OracleClaim, error) {
	if ethClaim.ClaimType != int64(x2eTy.LockClaimType) && ethClaim.ClaimType != int64(x2eTy.BurnClaimType) {
		return x2eTy.OracleClaim{}, x2eTy.ErrInvalidClaimType
	}
	oracleID := strconv.Itoa(int(ethClaim.EthereumChainID)) + strconv.Itoa(int(ethClaim.Nonce)) + ethClaim.EthereumSender + ethClaim.TokenContractAddress
	if ethClaim.ClaimType == int64(x2eTy.LockClaimType) {
		oracleID = oracleID + "lock"
	} else if ethClaim.ClaimType == int64(x2eTy.BurnClaimType) {
		oracleID = oracleID + "burn"
	}
	claimContent := NewOracleClaimContent(ethClaim.Chain33Receiver, ethClaim.Amount, ethClaim.ClaimType, ethClaim.Decimals)
	claimBytes := types.Encode(&claimContent)

	claimString := string(claimBytes)
	claim := NewClaim(oracleID, ethClaim.ValidatorAddress, claimString)
	return claim, nil
}

//CreateOracleClaimFromOracleString --
func CreateOracleClaimFromOracleString(oracleClaimString string) (x2eTy.OracleClaimContent, error) {
	var oracleClaimContent x2eTy.OracleClaimContent

	bz := []byte(oracleClaimString)
	if err := types.Decode(bz, &oracleClaimContent); err != nil {
		return x2eTy.OracleClaimContent{}, fmt.Errorf("failed to parse claim: %s", err.Error())
	}

	return oracleClaimContent, nil
}
