package oracle

import (
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
)

func NewProphecy(id string) *types.ReceiptEthProphecy {

	status := new(types.ProphecyStatus)
	status.Text = types.EthBridgeStatus_PendingStatusText

	return &types.ReceiptEthProphecy{
		ID:              id,
		Status:          status,
		ClaimValidators: *new([]*types.ClaimValidators),
		ValidatorClaims: *new([]*types.ValidatorClaims),
	}
}

func NewEmptyProphecy() *types.ReceiptEthProphecy {
	return NewProphecy("")
}

//
//type DBProphecy struct {
//	ID              string `json:"id"`
//	Status          Status `json:"status"`
//	ClaimValidators []byte `json:"claim_validators"`
//	ValidatorClaims []byte `json:"validator_claims"`
//}
//
//// SerializeForDB serializes a prophecy into a DBProphecy
//func (prophecy Prophecy) SerializeForDB() (DBProphecy, error) {
//	claimValidators, err := json.Marshal(prophecy.ClaimValidators)
//	if err != nil {
//		return DBProphecy{}, err
//	}
//
//	validatorClaims, err := json.Marshal(prophecy.ValidatorClaims)
//	if err != nil {
//		return DBProphecy{}, err
//	}
//
//	return DBProphecy{
//		ID:              prophecy.ID,
//		Status:          prophecy.Status,
//		ClaimValidators: claimValidators,
//		ValidatorClaims: validatorClaims,
//	}, nil
//}
//
//// DeserializeFromDB deserializes a DBProphecy into a prophecy
//func (dbProphecy DBProphecy) DeserializeFromDB() (Prophecy, error) {
//	claimValidators := new([]*types.ClaimValidators)
//	err := json.Unmarshal(dbProphecy.ClaimValidators, &claimValidators)
//	if err != nil {
//		return Prophecy{}, err
//	}
//
//	validatorClaims := new([]*types.ValidatorClaims)
//	err = json.Unmarshal(dbProphecy.ValidatorClaims, &validatorClaims)
//	if err != nil {
//		return Prophecy{}, err
//	}
//
//	return Prophecy{
//		ID:              dbProphecy.ID,
//		Status:          dbProphecy.Status,
//		ClaimValidators: *claimValidators,
//		ValidatorClaims: *validatorClaims,
//	}, nil
//}

// AddClaim adds a given claim to this prophecy
func AddClaim(prophecy *types.ReceiptEthProphecy, validator string, claim string) {
	claimValidators := new(types.StringMap)
	if len(prophecy.ClaimValidators) == 0 {
		prophecy.ClaimValidators = append(prophecy.ClaimValidators, &types.ClaimValidators{
			Claim: claim,
			Validators: &types.StringMap{
				Validators: []string{validator},
			},
		})
	} else {
		for index, cv := range prophecy.ClaimValidators {
			if cv.Claim == claim {
				claimValidators = cv.Validators
				prophecy.ClaimValidators[index].Validators = types.AddToStringMap(claimValidators, validator)
				break
			}
		}
	}

	//todo
	// validator不可能相同？
	//if len(prophecy.ValidatorClaims) == 0 {
	//	prophecy.ValidatorClaims = append(prophecy.ValidatorClaims, &types.ValidatorClaims{
	//		Validator: validator,
	//		Claim:     claim,
	//	})
	//} else {
	//	for index, vc := range prophecy.ValidatorClaims {
	//		if vc.Validator == validator {
	//			prophecy.ValidatorClaims[index].Claim = claim
	//			break
	//		} else {
	//			prophecy.ValidatorClaims = append(prophecy.ValidatorClaims, &types.ValidatorClaims{
	//				Validator: validator,
	//				Claim:     claim,
	//			})
	//		}
	//	}
	//}

	prophecy.ValidatorClaims = append(prophecy.ValidatorClaims, &types.ValidatorClaims{
		Validator: validator,
		Claim:     claim,
	})

}

// 遍历该prophecy所有claim，找出获得最多票数的claim
func FindHighestClaim(prophecy *types.ReceiptEthProphecy, validators map[string]int64) (string, int64, int64) {
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
