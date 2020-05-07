package ethbridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

var (
	//日志
	elog = log.New("module", "ethbridge")
)

func NewEthBridgeClaim(ethereumChainID int64, bridgeContract string, nonce int64, localCoinSymbol, localCoinExec string, tokenContact string, ethereumSender string, chain33Receiver string, validator string, amount string, claimType int64) types.Eth2Chain33 {
	return types.Eth2Chain33{
		EthereumChainID:       ethereumChainID,
		BridgeContractAddress: bridgeContract,
		Nonce:                nonce,
		TokenContractAddress: tokenContact,
		EthereumSender:       ethereumSender,
		Chain33Receiver:      chain33Receiver,
		ValidatorAddress:     validator,
		Amount:               amount,
		ClaimType:            claimType,
		LocalCoinSymbol:      localCoinSymbol,
		LocalCoinExec:        localCoinExec,
	}
}

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
	oracleID := strconv.Itoa(int(ethClaim.EthereumChainID)) + strconv.Itoa(int(ethClaim.Nonce)) + ethClaim.EthereumSender
	if ethClaim.ClaimType == int64(types.LOCK_CLAIM_TYPE) {
		oracleID = oracleID + "lock"
	} else if ethClaim.ClaimType == int64(types.BURN_CLAIM_TYPE) {
		oracleID = oracleID + "burn"
	}
	claimContent := NewOracleClaimContent(ethClaim.Chain33Receiver, ethClaim.Amount, ethClaim.ClaimType, ethClaim.Decimals)
	claimBytes, err := json.Marshal(claimContent)
	if err != nil {
		return types.OracleClaim{}, err
	}
	claimString := string(claimBytes)
	claim := NewClaim(oracleID, ethClaim.ValidatorAddress, claimString)
	return claim, nil
}

// 通过oracleclaim反向构造ethchain33结构
func CreateEthClaimFromOracleString(ethereumChainID int64, bridgeContract string, nonce int64, localCoinSymbol, localCoinExec string, tokenContract string, ethereumAddress string, validator string, oracleClaimString string) (types.Eth2Chain33, error) {
	oracleClaim, err := CreateOracleClaimFromOracleString(oracleClaimString)
	if err != nil {
		elog.Error("CreateEthClaimFromOracleString", "CreateOracleClaimFromOracleString error", err)
		return types.Eth2Chain33{}, err
	}

	return NewEthBridgeClaim(
		ethereumChainID,
		bridgeContract,
		nonce,
		localCoinSymbol,
		localCoinExec,
		tokenContract,
		ethereumAddress,
		oracleClaim.Chain33Receiver,
		validator,
		oracleClaim.Amount,
		oracleClaim.ClaimType,
	), nil
}

func CreateOracleClaimFromOracleString(oracleClaimString string) (types.OracleClaimContent, error) {
	var oracleClaimContent types.OracleClaimContent

	bz := []byte(oracleClaimString)
	if err := json.Unmarshal(bz, &oracleClaimContent); err != nil {
		return types.OracleClaimContent{}, errors.New(fmt.Sprintf("failed to parse claim: %s", err.Error()))
	}

	return oracleClaimContent, nil
}
