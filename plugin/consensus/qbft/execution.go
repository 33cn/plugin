// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbft

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	ttypes "github.com/33cn/plugin/plugin/consensus/qbft/types"
	tmtypes "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
)

//-----------------------------------------------------------------------------
// BlockExecutor handles block execution and state updates.
// It exposes ApplyBlock(), which validates & executes the block, updates state w/ ABCI responses,
// then commits and updates the mempool atomically, then saves state.

// BlockExecutor provides the context and accessories for properly executing a block.
type BlockExecutor struct {
	// save state, validators, consensus params, abci responses here
	db *CSStateDB
}

// NewBlockExecutor returns a new BlockExecutor with a NopEventBus.
// Call SetEventBus to provide one.
func NewBlockExecutor(db *CSStateDB) *BlockExecutor {
	return &BlockExecutor{
		db: db,
	}
}

// ValidateBlock validates the given block against the given state.
// If the block is invalid, it returns an error.
// Validation does not mutate state, but does require historical information from the stateDB,
// ie. to verify evidence from a validator at an old height.
func (blockExec *BlockExecutor) ValidateBlock(s State, block *ttypes.QbftBlock) error {
	return validateBlock(blockExec.db, s, block)
}

// ApplyBlock validates the block against the state, executes it against the app,
// fires the relevant events, commits the app, and saves the new state and responses.
// It's the only function that needs to be called
// from outside this package to process and commit an entire block.
// It takes a blockID to avoid recomputing the parts hash.
func (blockExec *BlockExecutor) ApplyBlock(s State, blockID ttypes.BlockID, block *ttypes.QbftBlock) (State, error) {
	if err := blockExec.ValidateBlock(s, block); err != nil {
		return s, fmt.Errorf("Commit failed for invalid block: %v", err)
	}

	// update the state with the block and responses
	s, err := updateState(s, blockID, block)
	if err != nil {
		return s, fmt.Errorf("Commit failed for application: %v", err)
	}

	blockExec.db.SaveState(s)
	return s, nil
}

// updateState returns a new QbftState updated according to the header and responses.
func updateState(s State, blockID ttypes.BlockID, block *ttypes.QbftBlock) (State, error) {

	// copy the valset so we can apply changes from EndBlock
	// and update s.LastValidators and s.Validators
	prevValSet := s.Validators.Copy()
	nextValSet := prevValSet.Copy()

	// update the validator set with the latest abciResponses
	lastHeightValsChanged := s.LastHeightValidatorsChanged

	seq := s.Sequence + 1
	// include situation multiBlock=1
	if seq == multiBlocks.Load().(int64) {
		// Update validator accums and set state variables
		nextValSet.IncrementAccum(1)
		seq = 0
	}

	// update the params with the latest abciResponses
	nextParams := s.ConsensusParams
	lastHeightParamsChanged := s.LastHeightConsensusParamsChanged

	// NOTE: the AppHash has not been populated.
	// It will be filled on state.Save.
	return State{
		ChainID:                          s.ChainID,
		LastBlockHeight:                  block.Header.Height,
		LastBlockTotalTx:                 s.LastBlockTotalTx + block.Header.NumTxs,
		LastBlockID:                      blockID,
		LastBlockTime:                    block.Header.Time,
		Validators:                       nextValSet,
		LastValidators:                   s.Validators.Copy(),
		LastHeightValidatorsChanged:      lastHeightValsChanged,
		ConsensusParams:                  nextParams,
		LastHeightConsensusParamsChanged: lastHeightParamsChanged,
		LastResultsHash:                  nil,
		AppHash:                          nil,
		Sequence:                         seq,
		LastSequence:                     s.Sequence,
		LastCommitRound:                  block.Header.Round,
	}, nil
}

func updateValidators(currentSet *ttypes.ValidatorSet, updates []*tmtypes.QbftNode) error {
	// If more or equal than 1/3 of total voting power changed in one block, then
	// a light client could never prove the transition externally. See
	// ./lite/doc.go for details on how a light client tracks validators.
	vp23, err := changeInVotingPowerMoreOrEqualToOneThird(currentSet, updates)
	if err != nil {
		return err
	}
	if vp23 {
		return errors.New("the change in voting power must be strictly less than 1/3")
	}

	for _, v := range updates {
		pubkeyBytes, err := hex.DecodeString(v.PubKey)
		if err != nil {
			return err
		}
		pubkey, err := ttypes.ConsensusCrypto.PubKeyFromBytes(pubkeyBytes)
		if err != nil {
			return err
		}

		address := ttypes.GenAddressByPubKey(pubkey)
		power := v.Power
		// mind the overflow from int64
		if power < 0 {
			return fmt.Errorf("Power (%d) overflows int64", v.Power)
		}

		_, val := currentSet.GetByAddress(address)
		if val == nil {
			// add val
			added := currentSet.Add(ttypes.NewValidator(pubkey, power))
			if !added {
				return fmt.Errorf("Failed to add new validator %X with voting power %d", address, power)
			}
		} else if v.Power == 0 {
			// remove val
			_, removed := currentSet.Remove(address)
			if !removed {
				return fmt.Errorf("Failed to remove validator %X", address)
			}
		} else {
			// update val
			val.VotingPower = power
			updated := currentSet.Update(val)
			if !updated {
				return fmt.Errorf("Failed to update validator %X with voting power %d", address, power)
			}
		}
	}
	return nil
}

func changeInVotingPowerMoreOrEqualToOneThird(currentSet *ttypes.ValidatorSet, updates []*tmtypes.QbftNode) (bool, error) {
	threshold := currentSet.TotalVotingPower() * 1 / 3
	acc := int64(0)

	for _, v := range updates {
		pubkeyBytes, err := hex.DecodeString(v.PubKey)
		if err != nil {
			return false, err
		}
		pubkey, err := ttypes.ConsensusCrypto.PubKeyFromBytes(pubkeyBytes)
		if err != nil {
			return false, err
		}

		address := ttypes.GenAddressByPubKey(pubkey)
		power := v.Power
		// mind the overflow from int64
		if power < 0 {
			return false, fmt.Errorf("Power (%d) overflows int64", v.Power)
		}

		_, val := currentSet.GetByAddress(address)
		if val == nil {
			acc += power
		} else {
			np := val.VotingPower - power
			if np < 0 {
				np = -np
			}
			acc += np
		}

		if acc >= threshold {
			return true, nil
		}
	}

	return false, nil
}

func validateBlock(stateDB *CSStateDB, s State, b *ttypes.QbftBlock) error {
	// Validate internal consistency.
	if err := b.ValidateBasic(); err != nil {
		return err
	}

	// validate basic info
	if b.Header.ChainID != s.ChainID {
		return fmt.Errorf("Wrong Block.Header.ChainID. Expected %v, got %v", s.ChainID, b.Header.ChainID)
	}
	if b.Header.Height != s.LastBlockHeight+1 {
		return fmt.Errorf("Wrong Block.Header.Height. Expected %v, got %v", s.LastBlockHeight+1, b.Header.Height)
	}

	// validate prev block info
	if !bytes.Equal(b.Header.LastBlockID.Hash, s.LastBlockID.Hash) {
		return fmt.Errorf("Wrong Block.Header.LastBlockID.  Expected %X, got %X", s.LastBlockID.Hash, b.Header.LastBlockID.Hash)
	}

	newTxs := b.Header.NumTxs
	if b.Header.TotalTxs != s.LastBlockTotalTx+newTxs {
		return fmt.Errorf("Wrong Block.Header.TotalTxs. Expected %v, got %v", s.LastBlockTotalTx+newTxs, b.Header.TotalTxs)
	}

	// validate app info
	if !bytes.Equal(b.Header.AppHash, s.AppHash) {
		return fmt.Errorf("Wrong Block.Header.AppHash.  Expected %X, got %X", s.AppHash, b.Header.AppHash)
	}
	if !bytes.Equal(b.Header.ConsensusHash, s.ConsensusParams.Hash()) {
		return fmt.Errorf("Wrong Block.Header.ConsensusHash.  Expected %X, got %X", s.ConsensusParams.Hash(), b.Header.ConsensusHash)
	}
	if !bytes.Equal(b.Header.LastResultsHash, s.LastResultsHash) {
		return fmt.Errorf("Wrong Block.Header.LastResultsHash.  Expected %X, got %X", s.LastResultsHash, b.Header.LastResultsHash)
	}
	if !bytes.Equal(b.Header.ValidatorsHash, s.Validators.Hash()) {
		return fmt.Errorf("Wrong Block.Header.ValidatorsHash.  Expected %X, got %X", s.Validators.Hash(), b.Header.ValidatorsHash)
	}

	// Validate block LastCommit.
	if b.Header.Height == 1 {
		if len(b.LastCommit.Precommits) != 0 {
			return errors.New("Block at height 1 (first block) not have LastCommit precommits")
		}
	} else {
		if (b.Header.LastSequence == 0 && b.LastCommit.VoteType != uint32(ttypes.VoteTypePrecommit)) ||
			(b.Header.LastSequence > 0 && b.LastCommit.VoteType != uint32(ttypes.VoteTypePrevote)) {
			return fmt.Errorf("Wrong LastCommit VoteType. LastSequence %v, VoteType %v",
				b.Header.LastSequence, b.LastCommit.VoteType)
		}
		lastCommit := &ttypes.Commit{QbftCommit: b.LastCommit}
		err := s.LastValidators.VerifyCommit(s.ChainID, s.LastBlockID, b.Header.Height-1, lastCommit)
		if err != nil {
			return err
		}
	}

	return nil
}
