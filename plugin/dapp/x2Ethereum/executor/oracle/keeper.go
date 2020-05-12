package oracle

import (
	"strings"

	"github.com/golang/protobuf/proto"

	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	types2 "github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

var (
	//日志
	olog = log.New("module", "oracle")
)

type OracleKeeper struct {
	db dbm.KV
	// 通过审核的最低阈值
	ConsensusThreshold int64
}

func NewOracleKeeper(db dbm.KV, ConsensusThreshold int64) *OracleKeeper {
	if ConsensusThreshold <= 0 || ConsensusThreshold > 100 {
		return nil
	}
	return &OracleKeeper{
		db:                 db,
		ConsensusThreshold: ConsensusThreshold,
	}
}

func (k *OracleKeeper) GetProphecy(id string) (*types.ReceiptEthProphecy, error) {
	if id == "" {
		return NewEmptyProphecy(), types.ErrInvalidIdentifier
	}

	bz, err := k.db.Get(types.CalProphecyPrefix(id))
	if err != nil && err != types2.ErrNotFound {
		return NewEmptyProphecy(), types.ErrProphecyGet
	} else if err == types2.ErrNotFound {
		return NewEmptyProphecy(), types.ErrProphecyNotFound
	}

	var dbProphecy types.ReceiptEthProphecy
	err = proto.Unmarshal(bz, &dbProphecy)
	if err != nil {
		return NewEmptyProphecy(), types2.ErrUnmarshal
	}
	return &dbProphecy, nil
}

// setProphecy saves a prophecy with an initial claim
func (k *OracleKeeper) setProphecy(prophecy *types.ReceiptEthProphecy) error {
	err := k.checkProphecy(prophecy)
	if err != nil {
		return err
	}

	bz, err := k.db.Get(types.CalProphecyPrefix(prophecy.ID))
	if err != nil && err != types2.ErrNotFound {
		return types.ErrProphecyGet
	}

	var dbProphecy types.ReceiptEthProphecy
	if err != types2.ErrNotFound {
		err = proto.Unmarshal(bz, &dbProphecy)
		if err != nil {
			return types2.ErrUnmarshal
		}
	}

	dbProphecy = *prophecy

	serializedProphecyBytes, err := proto.Marshal(&dbProphecy)
	if err != nil {
		return types2.ErrMarshal
	}

	err = k.db.Set(types.CalProphecyPrefix(prophecy.ID), serializedProphecyBytes)
	if err != nil {
		return types.ErrSetKV
	}
	return nil
}

func (k *OracleKeeper) checkProphecy(prophecy *types.ReceiptEthProphecy) error {
	if prophecy.ID == "" {
		return types.ErrInvalidIdentifier
	}
	if len(prophecy.ClaimValidators) == 0 {
		return types.ErrNoClaims
	}
	return nil
}

func (k *OracleKeeper) ProcessClaim(claim types.OracleClaim) (*types.ProphecyStatus, error) {
	activeValidator := k.checkActiveValidator(claim.ValidatorAddress)
	if !activeValidator {
		return nil, types.ErrInvalidValidator
	}
	if strings.TrimSpace(claim.Content) == "" {
		return nil, types.ErrInvalidClaim
	}
	var claimContent types.OracleClaimContent
	err := proto.Unmarshal([]byte(claim.Content), &claimContent)
	if err != nil {
		return nil, types2.ErrUnmarshal
	}
	prophecy, err := k.GetProphecy(claim.ID)
	if err != nil {
		if err != types.ErrProphecyNotFound {
			return nil, err
		}
		prophecy = NewProphecy(claim.ID)
	} else {
		var exist bool
		for _, vc := range prophecy.ValidatorClaims {
			if vc.Claim == claim.Content {
				exist = true
			}
		}
		if !exist {
			prophecy.Status.Text = types.EthBridgeStatus_FailedStatusText
			return nil, types.ErrClaimInconsist
		}
		if prophecy.Status.Text == types.EthBridgeStatus_FailedStatusText {
			return nil, types.ErrProphecyFinalized
		}
		for _, vc := range prophecy.ValidatorClaims {
			if vc.Validator == claim.ValidatorAddress && vc.Claim != "" {
				return nil, types.ErrDuplicateMessage
			}
		}
	}
	AddClaim(prophecy, claim.ValidatorAddress, claim.Content)
	prophecy, err = k.processCompletion(prophecy, claimContent.ClaimType)
	if err != nil {
		return nil, err
	}
	err = k.setProphecy(prophecy)
	if err != nil {
		return nil, err
	}
	return prophecy.Status, nil
}

func (k *OracleKeeper) checkActiveValidator(validatorAddress string) bool {
	validatorMap, err := k.GetValidatorArray()
	if err != nil {
		return false
	}

	for _, v := range validatorMap.Validators {
		if v.Address == validatorAddress {
			return true
		}
	}
	return false
}

// 计算该prophecy是否达标
func (k *OracleKeeper) processCompletion(prophecy *types.ReceiptEthProphecy, claimType int64) (*types.ReceiptEthProphecy, error) {
	address2power := make(map[string]int64)
	validatorArrays, err := k.GetValidatorArray()
	if err != nil {
		return nil, err
	}
	for _, validator := range validatorArrays.Validators {
		address2power[validator.Address] = validator.Power
	}
	highestClaim, highestClaimPower, totalClaimsPower := FindHighestClaim(prophecy, address2power)
	totalPower, err := k.GetLastTotalPower()
	if err != nil {
		return nil, err
	}
	highestConsensusRatio := highestClaimPower * 100
	remainingPossibleClaimPower := totalPower - totalClaimsPower
	highestPossibleClaimPower := highestClaimPower + remainingPossibleClaimPower
	highestPossibleConsensusRatio := highestPossibleClaimPower * 100
	olog.Info("processCompletion", "highestConsensusRatio", highestConsensusRatio/totalPower, "ConsensusThreshold", k.ConsensusThreshold, "highestPossibleConsensusRatio", highestPossibleConsensusRatio/totalPower)
	if highestConsensusRatio >= k.ConsensusThreshold*totalPower {
		prophecy.Status.Text = types.EthBridgeStatus_SuccessStatusText

		prophecy.Status.FinalClaim = highestClaim
	} else if highestPossibleConsensusRatio < k.ConsensusThreshold*totalPower {
		prophecy.Status.Text = types.EthBridgeStatus_FailedStatusText
	}
	return prophecy, nil
}

// Load the last total validator power.
func (k *OracleKeeper) GetLastTotalPower() (int64, error) {
	b, err := k.db.Get(types.CalLastTotalPowerPrefix())
	if err != nil && err != types2.ErrNotFound {
		return 0, err
	} else if err == types2.ErrNotFound {
		return 0, nil
	}
	var powers types.ReceiptQueryTotalPower
	err = proto.Unmarshal(b, &powers)
	if err != nil {
		return 0, types2.ErrUnmarshal
	}
	return powers.TotalPower, nil
}

// Set the last total validator power.
func (k *OracleKeeper) SetLastTotalPower() error {
	var totalPower int64
	validatorArrays, err := k.GetValidatorArray()
	if err != nil {
		return err
	}
	for _, validator := range validatorArrays.Validators {
		totalPower += validator.Power
	}
	totalP := types.ReceiptQueryTotalPower{
		TotalPower: totalPower,
	}
	totalPBytes, _ := proto.Marshal(&totalP)
	err = k.db.Set(types.CalLastTotalPowerPrefix(), totalPBytes)
	if err != nil {
		return types.ErrSetKV
	}
	return nil
}

func (k *OracleKeeper) GetValidatorArray() (*types.ValidatorList, error) {
	validatorsBytes, err := k.db.Get(types.CalValidatorMapsPrefix())
	if err != nil {
		return nil, err
	}
	var validatorArrays types.ValidatorList
	err = proto.Unmarshal(validatorsBytes, &validatorArrays)
	if err != nil {
		return nil, types2.ErrUnmarshal
	}
	return &validatorArrays, nil
}

func (k *OracleKeeper) SetConsensusThreshold(ConsensusThreshold int64) {
	k.ConsensusThreshold = ConsensusThreshold
	olog.Info("SetConsensusNeeded", "nowConsensusNeeded", k.ConsensusThreshold)
	return
}

func (k *OracleKeeper) GetConsensusThreshold() int64 {
	return k.ConsensusThreshold
}
