package ethbridge

import (
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"strconv"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

var (
	//日志
	elog = log.New("module", "ethbridge")
)

func NewOracleClaimContent(chain33Receiver string, amount string, claimType, decimals int64) types.OracleClaimContent {
	return types.OracleClaimContent{
		Chain33Receiver: chain33Receiver,
		Amount:          amount,
		ClaimType:       claimType,
		Decimals:        decimals,
	}
}

func NewClaim(id string, validatorAddress string, content string) types.OracleClaim {
	return types.OracleClaim{
		ID:               id,
		ValidatorAddress: validatorAddress,
		Content:          content,
	}
}

//通过ethchain33结构构造一个OracleClaim结构，包括生成唯一的ID
func CreateOracleClaimFromEthClaim(ethClaim types.Eth2Chain33) (types.OracleClaim, error) {
	if ethClaim.ClaimType != int64(types.LOCK_CLAIM_TYPE) && ethClaim.ClaimType != int64(types.BURN_CLAIM_TYPE) {
		return types.OracleClaim{}, types.ErrInvalidClaimType
	}
	oracleID := strconv.Itoa(int(ethClaim.EthereumChainID)) + strconv.Itoa(int(ethClaim.Nonce)) + ethClaim.EthereumSender + ethClaim.TokenContractAddress
	if ethClaim.ClaimType == int64(types.LOCK_CLAIM_TYPE) {
		oracleID = oracleID + "lock"
	} else if ethClaim.ClaimType == int64(types.BURN_CLAIM_TYPE) {
		oracleID = oracleID + "burn"
	}
	claimContent := NewOracleClaimContent(ethClaim.Chain33Receiver, ethClaim.Amount, ethClaim.ClaimType, ethClaim.Decimals)
	claimBytes, err := proto.Marshal(&claimContent)
	if err != nil {
		return types.OracleClaim{}, err
	}
	claimString := string(claimBytes)
	claim := NewClaim(oracleID, ethClaim.ValidatorAddress, claimString)
	return claim, nil
}

func CreateOracleClaimFromOracleString(oracleClaimString string) (types.OracleClaimContent, error) {
	var oracleClaimContent types.OracleClaimContent

	bz := []byte(oracleClaimString)
	if err := proto.Unmarshal(bz, &oracleClaimContent); err != nil {
		return types.OracleClaimContent{}, errors.New(fmt.Sprintf("failed to parse claim: %s", err.Error()))
	}

	return oracleClaimContent, nil
}
