// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/33cn/chain33/common/address"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
)

const (

	//ShuffleTypeNoVrf shuffle type: NoVrf, use default address order
	ShuffleTypeNoVrf = iota

	//ShuffleTypeVrf shuffle type: Vrf
	ShuffleTypeVrf

	//ShuffleTypePartVrf shuffle type: PartVrf
	ShuffleTypePartVrf
)

// ValidatorMgr ...
type ValidatorMgr struct {
	// Immutable
	ChainID string

	// Validators are persisted to the database separately every time they change,
	// so we can query for historical validator sets.
	// Note that if s.LastBlockHeight causes a valset change,
	// we set s.LastHeightValidatorsChanged = s.LastBlockHeight + 1
	Validators            *ttypes.ValidatorSet
	VrfValidators         *ttypes.ValidatorSet
	NoVrfValidators       *ttypes.ValidatorSet
	LastCycleBoundaryInfo *dty.DposCBInfo
	ShuffleCycle          int64
	ShuffleType           int64 //0-no vrf 1-vrf 2-part vrf
	// The latest AppHash we've received from calling abci.Commit()
	AppHash []byte
}

// Copy makes a copy of the State for mutating.
func (s ValidatorMgr) Copy() ValidatorMgr {
	mgr := ValidatorMgr{
		ChainID:      s.ChainID,
		Validators:   s.Validators.Copy(),
		AppHash:      s.AppHash,
		ShuffleCycle: s.ShuffleCycle,
		ShuffleType:  s.ShuffleType,
		//VrfValidators: s.VrfValidators.Copy(),
		//NoVrfValidators: s.NoVrfValidators.Copy(),
		//LastCycleBoundaryInfo: &dty.DposCBInfo{
		//	Cycle: s.LastCycleBoundaryInfo.Cycle,
		//	StopHeight: s.LastCycleBoundaryInfo.StopHeight,
		//	StopHash: s.LastCycleBoundaryInfo.StopHash,
		//	Pubkey: s.LastCycleBoundaryInfo.Pubkey,
		//	Signature: s.LastCycleBoundaryInfo.Signature,
		//},
	}

	if s.LastCycleBoundaryInfo != nil {
		mgr.LastCycleBoundaryInfo = &dty.DposCBInfo{
			Cycle:      s.LastCycleBoundaryInfo.Cycle,
			StopHeight: s.LastCycleBoundaryInfo.StopHeight,
			StopHash:   s.LastCycleBoundaryInfo.StopHash,
			Pubkey:     s.LastCycleBoundaryInfo.Pubkey,
			Signature:  s.LastCycleBoundaryInfo.Signature,
		}
	}

	if s.VrfValidators != nil {
		mgr.VrfValidators = s.VrfValidators.Copy()
	}

	if s.NoVrfValidators != nil {
		mgr.NoVrfValidators = s.NoVrfValidators.Copy()
	}

	return mgr
}

// Equals returns true if the States are identical.
func (s ValidatorMgr) Equals(s2 ValidatorMgr) bool {
	return bytes.Equal(s.Bytes(), s2.Bytes())
}

// Bytes serializes the State using go-wire.
func (s ValidatorMgr) Bytes() []byte {
	sbytes, err := json.Marshal(s)
	if err != nil {
		fmt.Printf("Error reading GenesisDoc: %v", err)
		return nil
	}
	return sbytes
}

// IsEmpty returns true if the State is equal to the empty State.
func (s ValidatorMgr) IsEmpty() bool {
	return s.Validators == nil // XXX can't compare to Empty
}

// GetValidators returns the last and current validator sets.
func (s ValidatorMgr) GetValidators() (current *ttypes.ValidatorSet) {
	return s.Validators
}

// MakeGenesisValidatorMgr creates validators from ttypes.GenesisDoc.
func MakeGenesisValidatorMgr(genDoc *ttypes.GenesisDoc) (ValidatorMgr, error) {
	err := genDoc.ValidateAndComplete()
	if err != nil {
		return ValidatorMgr{}, fmt.Errorf("Error in genesis file: %v", err)
	}

	// Make validators slice
	validators := make([]*ttypes.Validator, len(genDoc.Validators))
	for i, val := range genDoc.Validators {
		pubKey, err := ttypes.PubKeyFromString(val.PubKey.Data)
		if err != nil {
			return ValidatorMgr{}, fmt.Errorf("Error validate[%v] in genesis file: %v", i, err)
		}

		// Make validator
		validators[i] = &ttypes.Validator{
			Address: address.BytesToBtcAddress(address.NormalVer, pubKey.Bytes()).Hash160[:],
			PubKey:  pubKey.Bytes(),
		}
	}

	return ValidatorMgr{
		ChainID:    genDoc.ChainID,
		Validators: ttypes.NewValidatorSet(validators),
		AppHash:    genDoc.AppHash,
	}, nil
}

// GetValidatorByIndex method
func (s *ValidatorMgr) GetValidatorByIndex(index int) (addr []byte, val *ttypes.Validator) {
	if index < 0 || index >= len(s.Validators.Validators) {
		return nil, nil
	}

	if s.ShuffleType == ShuffleTypeNoVrf {
		val = s.Validators.Validators[index]
		return val.Address, val.Copy()
	} else if s.ShuffleType == ShuffleTypeVrf {
		val = s.VrfValidators.Validators[index]
		return address.BytesToBtcAddress(address.NormalVer, val.PubKey).Hash160[:], val.Copy()
	} else if s.ShuffleType == ShuffleTypePartVrf {
		if index < len(s.VrfValidators.Validators) {
			val = s.VrfValidators.Validators[index]
			return address.BytesToBtcAddress(address.NormalVer, val.PubKey).Hash160[:], val.Copy()
		}

		val = s.NoVrfValidators.Validators[index-len(s.VrfValidators.Validators)]
		return address.BytesToBtcAddress(address.NormalVer, val.PubKey).Hash160[:], val.Copy()
	}

	return nil, nil
}

// GetIndexByPubKey method
func (s *ValidatorMgr) GetIndexByPubKey(pubkey []byte) (index int) {
	if nil == pubkey {
		return -1
	}

	index = -1

	if s.ShuffleType == ShuffleTypeNoVrf {
		for i := 0; i < s.Validators.Size(); i++ {
			if bytes.Equal(s.Validators.Validators[i].PubKey, pubkey) {
				index = i
				return index
			}
		}
	} else if s.ShuffleType == ShuffleTypeVrf {
		for i := 0; i < s.VrfValidators.Size(); i++ {
			if bytes.Equal(s.VrfValidators.Validators[i].PubKey, pubkey) {
				index = i
				return index
			}
		}
	} else if s.ShuffleType == ShuffleTypePartVrf {
		for i := 0; i < s.VrfValidators.Size(); i++ {
			if bytes.Equal(s.VrfValidators.Validators[i].PubKey, pubkey) {
				index = i
				return index
			}
		}

		for j := 0; j < s.NoVrfValidators.Size(); j++ {
			if bytes.Equal(s.NoVrfValidators.Validators[j].PubKey, pubkey) {
				index = j + s.VrfValidators.Size()
				return index
			}
		}
	}

	return index
}

// FillVoteItem method
func (s *ValidatorMgr) FillVoteItem(voteItem *ttypes.VoteItem) {
	if s.LastCycleBoundaryInfo != nil {
		voteItem.LastCBInfo = &ttypes.CycleBoundaryInfo{
			Cycle:      s.LastCycleBoundaryInfo.Cycle,
			StopHeight: s.LastCycleBoundaryInfo.StopHeight,
			StopHash:   s.LastCycleBoundaryInfo.StopHash,
		}
	}

	voteItem.ShuffleType = s.ShuffleType
	for i := 0; s.Validators != nil && i < s.Validators.Size(); i++ {
		node := &ttypes.SuperNode{
			PubKey:  s.Validators.Validators[i].PubKey,
			Address: s.Validators.Validators[i].Address,
		}
		voteItem.Validators = append(voteItem.Validators, node)
	}

	for i := 0; s.VrfValidators != nil && i < s.VrfValidators.Size(); i++ {
		node := &ttypes.SuperNode{
			PubKey:  s.VrfValidators.Validators[i].PubKey,
			Address: s.VrfValidators.Validators[i].Address,
		}
		voteItem.VrfValidators = append(voteItem.VrfValidators, node)
	}

	for i := 0; s.NoVrfValidators != nil && i < s.NoVrfValidators.Size(); i++ {
		node := &ttypes.SuperNode{
			PubKey:  s.NoVrfValidators.Validators[i].PubKey,
			Address: s.NoVrfValidators.Validators[i].Address,
		}
		voteItem.NoVrfValidators = append(voteItem.NoVrfValidators, node)
	}
}

// UpdateFromVoteItem method
func (s *ValidatorMgr) UpdateFromVoteItem(voteItem *ttypes.VoteItem) bool {
	validators := voteItem.Validators
	if len(s.Validators.Validators) != len(voteItem.Validators) {
		return false
	}

	for i := 0; i < s.Validators.Size(); i++ {
		if !bytes.Equal(validators[i].PubKey, s.Validators.Validators[i].PubKey) {
			return false
		}
	}

	if voteItem.LastCBInfo != nil {
		if s.LastCycleBoundaryInfo == nil {
			s.LastCycleBoundaryInfo = &dty.DposCBInfo{
				Cycle:      voteItem.LastCBInfo.Cycle,
				StopHeight: voteItem.LastCBInfo.StopHeight,
				StopHash:   voteItem.LastCBInfo.StopHash,
			}
		} else if voteItem.LastCBInfo.Cycle != s.LastCycleBoundaryInfo.Cycle ||
			voteItem.LastCBInfo.StopHeight != s.LastCycleBoundaryInfo.StopHeight ||
			voteItem.LastCBInfo.StopHash != s.LastCycleBoundaryInfo.StopHash {
			s.LastCycleBoundaryInfo = &dty.DposCBInfo{
				Cycle:      voteItem.LastCBInfo.Cycle,
				StopHeight: voteItem.LastCBInfo.StopHeight,
				StopHash:   voteItem.LastCBInfo.StopHash,
			}
		}
	}

	s.ShuffleType = voteItem.ShuffleType

	var vrfVals []*ttypes.Validator
	for i := 0; i < len(voteItem.VrfValidators); i++ {
		val := &ttypes.Validator{
			Address: voteItem.VrfValidators[i].Address,
			PubKey:  voteItem.VrfValidators[i].PubKey,
		}

		vrfVals = append(vrfVals, val)
	}

	s.VrfValidators = ttypes.NewValidatorSet(vrfVals)

	var noVrfVals []*ttypes.Validator
	for i := 0; i < len(voteItem.NoVrfValidators); i++ {
		val := &ttypes.Validator{
			Address: voteItem.NoVrfValidators[i].Address,
			PubKey:  voteItem.NoVrfValidators[i].PubKey,
		}

		noVrfVals = append(noVrfVals, val)
	}

	s.NoVrfValidators = ttypes.NewValidatorSet(noVrfVals)

	return true
}
