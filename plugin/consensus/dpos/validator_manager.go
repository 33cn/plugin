// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"

	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
)

const fee = 1e6

var (
	r *rand.Rand
)

// ValidatorMgr ...
type ValidatorMgr struct {
	// Immutable
	ChainID string

	// Validators are persisted to the database separately every time they change,
	// so we can query for historical validator sets.
	// Note that if s.LastBlockHeight causes a valset change,
	// we set s.LastHeightValidatorsChanged = s.LastBlockHeight + 1
	Validators *ttypes.ValidatorSet

	// The latest AppHash we've received from calling abci.Commit()
	AppHash []byte
}

// Copy makes a copy of the State for mutating.
func (s ValidatorMgr) Copy() ValidatorMgr {
	return ValidatorMgr{
		ChainID: s.ChainID,

		Validators: s.Validators.Copy(),

		AppHash: s.AppHash,
	}
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
			Address: ttypes.GenAddressByPubKey(pubKey),
			PubKey:  pubKey.Bytes(),
		}
	}

	return ValidatorMgr{
		ChainID:    genDoc.ChainID,
		Validators: ttypes.NewValidatorSet(validators),
		AppHash:    genDoc.AppHash,
	}, nil
}
