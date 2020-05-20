package executor

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/33cn/chain33/types"
	x2eTy "github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

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
	claimValidators := new(x2eTy.StringMap)
	if len(prophecy.ClaimValidators) == 0 {
		prophecy.ClaimValidators = append(prophecy.ClaimValidators, &x2eTy.ClaimValidators{
			Claim: claim,
			Validators: &x2eTy.StringMap{
				Validators: []string{validator},
			},
		})
	} else {
		for index, cv := range prophecy.ClaimValidators {
			if cv.Claim == claim {
				claimValidators = cv.Validators
				prophecy.ClaimValidators[index].Validators = x2eTy.AddToStringMap(claimValidators, validator)
				break
			}
		}
	}

	prophecy.ValidatorClaims = append(prophecy.ValidatorClaims, &x2eTy.ValidatorClaims{
		Validator: validator,
		Claim:     claim,
	})

}

// 遍历该prophecy所有claim，找出获得最多票数的claim
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

func NewOracleClaimContent(chain33Receiver string, amount string, claimType, decimals int64) x2eTy.OracleClaimContent {
	return x2eTy.OracleClaimContent{
		Chain33Receiver: chain33Receiver,
		Amount:          amount,
		ClaimType:       claimType,
		Decimals:        decimals,
	}
}

func NewClaim(id string, validatorAddress string, content string) x2eTy.OracleClaim {
	return x2eTy.OracleClaim{
		ID:               id,
		ValidatorAddress: validatorAddress,
		Content:          content,
	}
}

//通过ethchain33结构构造一个OracleClaim结构，包括生成唯一的ID
func CreateOracleClaimFromEthClaim(ethClaim x2eTy.Eth2Chain33) (x2eTy.OracleClaim, error) {
	if ethClaim.ClaimType != int64(x2eTy.LOCK_CLAIM_TYPE) && ethClaim.ClaimType != int64(x2eTy.BURN_CLAIM_TYPE) {
		return x2eTy.OracleClaim{}, x2eTy.ErrInvalidClaimType
	}
	oracleID := strconv.Itoa(int(ethClaim.EthereumChainID)) + strconv.Itoa(int(ethClaim.Nonce)) + ethClaim.EthereumSender + ethClaim.TokenContractAddress
	if ethClaim.ClaimType == int64(x2eTy.LOCK_CLAIM_TYPE) {
		oracleID = oracleID + "lock"
	} else if ethClaim.ClaimType == int64(x2eTy.BURN_CLAIM_TYPE) {
		oracleID = oracleID + "burn"
	}
	claimContent := NewOracleClaimContent(ethClaim.Chain33Receiver, ethClaim.Amount, ethClaim.ClaimType, ethClaim.Decimals)
	claimBytes := types.Encode(&claimContent)

	claimString := string(claimBytes)
	claim := NewClaim(oracleID, ethClaim.ValidatorAddress, claimString)
	return claim, nil
}

func CreateOracleClaimFromOracleString(oracleClaimString string) (x2eTy.OracleClaimContent, error) {
	var oracleClaimContent x2eTy.OracleClaimContent

	bz := []byte(oracleClaimString)
	if err := types.Decode(bz, &oracleClaimContent); err != nil {
		return x2eTy.OracleClaimContent{}, errors.New(fmt.Sprintf("failed to parse claim: %s", err.Error()))
	}

	return oracleClaimContent, nil
}
