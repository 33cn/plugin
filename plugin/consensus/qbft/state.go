// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbft

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/qbft/types"
	tmtypes "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
)

const fee = 1e6

// QbftState is a short description of the latest committed block of the Qbft consensus.
// It keeps all information necessary to validate new blocks,
// including the last validator set and the consensus params.
// All fields are exposed so the struct can be easily serialized,
// but none of them should be mutated directly.
// Instead, use state.Copy() or state.NextState(...).
// NOTE: not goroutine-safe.
type State struct {
	// Immutable
	ChainID string

	// LastBlockHeight=0 at genesis (ie. block(H=0) does not exist)
	LastBlockHeight  int64
	LastBlockTotalTx int64
	LastBlockID      ttypes.BlockID
	LastBlockTime    int64

	// LastValidators is used to validate block.LastCommit.
	// Validators are persisted to the database separately every time they change,
	// so we can query for historical validator sets.
	// Note that if s.LastBlockHeight causes a valset change,
	// we set s.LastHeightValidatorsChanged = s.LastBlockHeight + 1
	Validators                  *ttypes.ValidatorSet
	LastValidators              *ttypes.ValidatorSet
	LastHeightValidatorsChanged int64

	// Consensus parameters used for validating blocks.
	// Changes returned by EndBlock and updated after Commit.
	ConsensusParams                  ttypes.ConsensusParams
	LastHeightConsensusParamsChanged int64

	// Merkle root of the results from executing prev block
	LastResultsHash []byte

	// The latest AppHash we've received from calling abci.Commit()
	AppHash         []byte
	Sequence        int64
	LastSequence    int64
	LastCommitRound int64
}

// Copy makes a copy of the QbftState for mutating.
func (s State) Copy() State {
	if &s == nil {
		return State{}
	}
	return State{
		ChainID: s.ChainID,

		LastBlockHeight:  s.LastBlockHeight,
		LastBlockTotalTx: s.LastBlockTotalTx,
		LastBlockID:      s.LastBlockID,
		LastBlockTime:    s.LastBlockTime,

		Validators:                  s.Validators.Copy(),
		LastValidators:              s.LastValidators.Copy(),
		LastHeightValidatorsChanged: s.LastHeightValidatorsChanged,

		ConsensusParams:                  s.ConsensusParams,
		LastHeightConsensusParamsChanged: s.LastHeightConsensusParamsChanged,

		AppHash: s.AppHash,

		LastResultsHash: s.LastResultsHash,
		Sequence:        s.Sequence,
		LastSequence:    s.LastSequence,
		LastCommitRound: s.LastCommitRound,
	}
}

// Equals returns true if the States are identical.
func (s State) Equals(s2 State) bool {
	return bytes.Equal(s.Bytes(), s2.Bytes())
}

// Bytes serializes the QbftState using go-wire.
func (s State) Bytes() []byte {
	sbytes, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return sbytes
}

// IsEmpty returns true if the QbftState is equal to the empty QbftState.
func (s State) IsEmpty() bool {
	return s.Validators == nil // XXX can't compare to Empty
}

// GetValidators returns the last and current validator sets.
func (s State) GetValidators() (last *ttypes.ValidatorSet, current *ttypes.ValidatorSet) {
	return s.LastValidators, s.Validators
}

// String returns a string
func (s State) String() string {
	return s.StringIndented("")
}

// StringIndented returns a string
func (s State) StringIndented(indent string) string {
	return fmt.Sprintf(`State{
%s  ChainID:           %v
%s  LastBlockHeight:   %v
%s  LastBlockTotalTx:  %v
%s  LastBlockID:       %X
%s  Validators:        %v
%s  LastProposer:      %v
%s  Sequence:          %v
%s  LastSequence:      %v
%s  LastCommitRound:   %v
%s}`,
		indent, s.ChainID,
		indent, s.LastBlockHeight,
		indent, s.LastBlockTotalTx,
		indent, s.LastBlockID,
		indent, s.Validators.StringIndented(indent),
		indent, s.LastValidators.GetProposer().String(),
		indent, s.Sequence,
		indent, s.LastSequence,
		indent, s.LastCommitRound,
		indent)
}

//------------------------------------------------------------------------
// Create a block from the latest state

// MakeBlock builds a block with the given txs and commit from the current state.
func (s State) MakeBlock(height int64, round int64, pblock *types.Block, commit *tmtypes.QbftCommit, proposerAddr []byte) *ttypes.QbftBlock {
	// build base block
	block := ttypes.MakeBlock(height, round, pblock, commit)

	// fill header with state data
	block.Header.ChainID = s.ChainID
	block.Header.TotalTxs = s.LastBlockTotalTx + block.Header.NumTxs
	block.Header.LastBlockID = s.LastBlockID.QbftBlockID
	block.Header.ValidatorsHash = s.Validators.Hash()
	block.Header.AppHash = s.AppHash
	block.Header.ConsensusHash = s.ConsensusParams.Hash()
	block.Header.LastResultsHash = s.LastResultsHash
	block.Header.ProposerAddr = proposerAddr
	block.Header.Sequence = s.Sequence
	block.Header.LastSequence = s.LastSequence

	return block
}

//------------------------------------------------------------------------
// Genesis

// MakeGenesisStateFromFile reads and unmarshals state from the given
// file.
//
// Used during replay and in tests.
func MakeGenesisStateFromFile(genDocFile string) (State, error) {
	genDoc, err := MakeGenesisDocFromFile(genDocFile)
	if err != nil {
		return State{}, err
	}
	return MakeGenesisState(genDoc)
}

// MakeGenesisDocFromFile reads and unmarshals genesis doc from the given file.
func MakeGenesisDocFromFile(genDocFile string) (*ttypes.GenesisDoc, error) {
	genDocJSON, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, fmt.Errorf("Couldn't read GenesisDoc file: %v", err)
	}
	genDoc, err := ttypes.GenesisDocFromJSON(genDocJSON)
	if err != nil {
		return nil, fmt.Errorf("Error reading GenesisDoc: %v", err)
	}
	return genDoc, nil
}

// MakeGenesisState creates state from ttypes.GenesisDoc.
func MakeGenesisState(genDoc *ttypes.GenesisDoc) (State, error) {
	err := genDoc.ValidateAndComplete()
	if err != nil {
		return State{}, fmt.Errorf("Error in genesis file: %v", err)
	}

	// Make validators slice
	validators := make([]*ttypes.Validator, len(genDoc.Validators))
	for i, val := range genDoc.Validators {
		pubKey, err := ttypes.PubKeyFromString(val.PubKey.Data)
		if err != nil {
			return State{}, fmt.Errorf("Error validate[%v] in genesis file: %v", i, err)
		}

		// Make validator
		validators[i] = &ttypes.Validator{
			Address:     ttypes.GenAddressByPubKey(pubKey),
			PubKey:      pubKey.Bytes(),
			VotingPower: val.Power,
		}
	}

	return State{

		ChainID: genDoc.ChainID,

		LastBlockHeight: 0,
		LastBlockID:     ttypes.BlockID{QbftBlockID: &tmtypes.QbftBlockID{}},
		LastBlockTime:   genDoc.GenesisTime.UnixNano(),

		Validators:                  ttypes.NewValidatorSet(validators),
		LastValidators:              ttypes.NewValidatorSet(nil),
		LastHeightValidatorsChanged: 0,

		ConsensusParams:                  *genDoc.ConsensusParams,
		LastHeightConsensusParamsChanged: 0,

		AppHash:         genDoc.AppHash,
		Sequence:        0,
		LastSequence:    0,
		LastCommitRound: 0,
	}, nil
}

// CSStateDB just for EvidencePool and BlockExecutor
type CSStateDB struct {
	client *Client
	state  State
	mtx    sync.Mutex
}

// NewStateDB make a new one
func NewStateDB(client *Client, state State) *CSStateDB {
	return &CSStateDB{
		client: client,
		state:  state,
	}
}

// SaveState to state cache
func (csdb *CSStateDB) SaveState(state State) {
	csdb.mtx.Lock()
	defer csdb.mtx.Unlock()
	csdb.state = state.Copy()
}

// LoadState convert external state to internal state
func LoadState(state *tmtypes.QbftState) State {
	stateTmp := State{
		ChainID:                          state.GetChainID(),
		LastBlockHeight:                  state.GetLastBlockHeight(),
		LastBlockTotalTx:                 state.GetLastBlockTotalTx(),
		LastBlockID:                      ttypes.BlockID{QbftBlockID: state.LastBlockID},
		LastBlockTime:                    state.LastBlockTime,
		Validators:                       &ttypes.ValidatorSet{Validators: make([]*ttypes.Validator, 0), Proposer: &ttypes.Validator{}},
		LastValidators:                   &ttypes.ValidatorSet{Validators: make([]*ttypes.Validator, 0), Proposer: &ttypes.Validator{}},
		LastHeightValidatorsChanged:      state.LastHeightValidatorsChanged,
		ConsensusParams:                  ttypes.ConsensusParams{BlockSize: ttypes.BlockSize{}, TxSize: ttypes.TxSize{}, BlockGossip: ttypes.BlockGossip{}, EvidenceParams: ttypes.EvidenceParams{}},
		LastHeightConsensusParamsChanged: state.LastHeightConsensusParamsChanged,
		LastResultsHash:                  state.LastResultsHash,
		AppHash:                          state.AppHash,
		Sequence:                         state.Sequence,
		LastSequence:                     state.LastSequence,
		LastCommitRound:                  state.LastCommitRound,
	}
	if validators := state.GetValidators(); validators != nil {
		if array := validators.GetValidators(); array != nil {
			targetArray := make([]*ttypes.Validator, len(array))
			LoadValidators(targetArray, array)
			stateTmp.Validators.Validators = targetArray
		}
		if proposer := validators.GetProposer(); proposer != nil {
			if val, err := LoadProposer(proposer); err == nil {
				stateTmp.Validators.Proposer = val
			}
		}
	}
	if lastValidators := state.GetLastValidators(); lastValidators != nil {
		if array := lastValidators.GetValidators(); array != nil {
			targetArray := make([]*ttypes.Validator, len(array))
			LoadValidators(targetArray, array)
			stateTmp.LastValidators.Validators = targetArray
		}
		if proposer := lastValidators.GetProposer(); proposer != nil {
			if val, err := LoadProposer(proposer); err == nil {
				stateTmp.LastValidators.Proposer = val
			}
		}
	}
	if consensusParams := state.GetConsensusParams(); consensusParams != nil {
		if consensusParams.GetBlockSize() != nil {
			stateTmp.ConsensusParams.BlockSize.MaxBytes = int(consensusParams.BlockSize.MaxBytes)
			stateTmp.ConsensusParams.BlockSize.MaxGas = consensusParams.BlockSize.MaxGas
			stateTmp.ConsensusParams.BlockSize.MaxTxs = int(consensusParams.BlockSize.MaxTxs)
		}
		if consensusParams.GetTxSize() != nil {
			stateTmp.ConsensusParams.TxSize.MaxGas = consensusParams.TxSize.MaxGas
			stateTmp.ConsensusParams.TxSize.MaxBytes = int(consensusParams.TxSize.MaxBytes)
		}
		if consensusParams.GetBlockGossip() != nil {
			stateTmp.ConsensusParams.BlockGossip.BlockPartSizeBytes = int(consensusParams.BlockGossip.BlockPartSizeBytes)
		}
		if consensusParams.GetEvidenceParams() != nil {
			stateTmp.ConsensusParams.EvidenceParams.MaxAge = consensusParams.EvidenceParams.MaxAge
		}
	}

	return stateTmp
}

func saveConsensusParams(dest *tmtypes.QbftConsensusParams, source ttypes.ConsensusParams) {
	dest.BlockSize.MaxBytes = int32(source.BlockSize.MaxBytes)
	dest.BlockSize.MaxTxs = int32(source.BlockSize.MaxTxs)
	dest.BlockSize.MaxGas = source.BlockSize.MaxGas
	dest.TxSize.MaxGas = source.TxSize.MaxGas
	dest.TxSize.MaxBytes = int32(source.TxSize.MaxBytes)
	dest.BlockGossip.BlockPartSizeBytes = int32(source.BlockGossip.BlockPartSizeBytes)
	dest.EvidenceParams.MaxAge = source.EvidenceParams.MaxAge
}

func saveValidators(dest []*tmtypes.QbftValidator, source []*ttypes.Validator) []*tmtypes.QbftValidator {
	for _, item := range source {
		if item == nil {
			dest = append(dest, &tmtypes.QbftValidator{})
		} else {
			validator := &tmtypes.QbftValidator{
				Address:     fmt.Sprintf("%X", item.Address),
				PubKey:      fmt.Sprintf("%X", item.PubKey),
				VotingPower: item.VotingPower,
				Accum:       item.Accum,
			}
			dest = append(dest, validator)
		}
	}
	return dest
}

func saveProposer(dest *tmtypes.QbftValidator, source *ttypes.Validator) {
	if source != nil {
		dest.Address = fmt.Sprintf("%X", source.Address)
		dest.PubKey = fmt.Sprintf("%X", source.PubKey)
		dest.VotingPower = source.VotingPower
		dest.Accum = source.Accum
	}
}

// SaveState convert internal state to external state
func SaveState(state State) *tmtypes.QbftState {
	newState := tmtypes.QbftState{
		ChainID:                          state.ChainID,
		LastBlockHeight:                  state.LastBlockHeight,
		LastBlockTotalTx:                 state.LastBlockTotalTx,
		LastBlockID:                      state.LastBlockID.QbftBlockID,
		LastBlockTime:                    state.LastBlockTime,
		Validators:                       &tmtypes.QbftValidatorSet{Validators: make([]*tmtypes.QbftValidator, 0), Proposer: &tmtypes.QbftValidator{}},
		LastValidators:                   &tmtypes.QbftValidatorSet{Validators: make([]*tmtypes.QbftValidator, 0), Proposer: &tmtypes.QbftValidator{}},
		LastHeightValidatorsChanged:      state.LastHeightValidatorsChanged,
		ConsensusParams:                  &tmtypes.QbftConsensusParams{BlockSize: &tmtypes.QbftBlockSize{}, TxSize: &tmtypes.QbftTxSize{}, BlockGossip: &tmtypes.QbftBlockGossip{}, EvidenceParams: &tmtypes.QbftEvidenceParams{}},
		LastHeightConsensusParamsChanged: state.LastHeightConsensusParamsChanged,
		LastResultsHash:                  state.LastResultsHash,
		AppHash:                          state.AppHash,
		Sequence:                         state.Sequence,
		LastSequence:                     state.LastSequence,
		LastCommitRound:                  state.LastCommitRound,
	}
	if state.Validators != nil {
		newState.Validators.Validators = saveValidators(newState.Validators.Validators, state.Validators.Validators)
		saveProposer(newState.Validators.Proposer, state.Validators.Proposer)
	}
	if state.LastValidators != nil {
		newState.LastValidators.Validators = saveValidators(newState.LastValidators.Validators, state.LastValidators.Validators)
		saveProposer(newState.LastValidators.Proposer, state.LastValidators.Proposer)
	}
	saveConsensusParams(newState.ConsensusParams, state.ConsensusParams)
	return &newState
}

// LoadValidators convert all external validators to internal validators
func LoadValidators(des []*ttypes.Validator, source []*tmtypes.QbftValidator) {
	for i, item := range source {
		if item.GetAddress() == "" {
			qbftlog.Warn("LoadValidators address is empty")
			continue
		}
		if item.GetPubKey() == "" {
			qbftlog.Warn("LoadValidators pubkey is empty")
			continue
		}
		addr, err := hex.DecodeString(item.GetAddress())
		if err != nil {
			qbftlog.Warn("LoadValidators address is invalid")
			continue
		}
		pubkey, err := hex.DecodeString(item.GetPubKey())
		if err != nil {
			qbftlog.Warn("LoadValidators pubkey is invalid")
			continue
		}

		des[i] = &ttypes.Validator{}
		des[i].Address = addr
		des[i].PubKey = pubkey
		des[i].VotingPower = item.VotingPower
		des[i].Accum = item.Accum
	}
}

// LoadProposer convert external proposer to internal proposer
func LoadProposer(source *tmtypes.QbftValidator) (*ttypes.Validator, error) {
	if source.GetAddress() == "" {
		qbftlog.Warn("LoadProposer address is empty")
		return nil, errors.New("LoadProposer address is empty")
	}
	if source.GetPubKey() == "" {
		qbftlog.Warn("LoadProposer pubkey is empty")
		return nil, errors.New("LoadProposer pubkey is empty")
	}
	addr, err := hex.DecodeString(source.GetAddress())
	if err != nil {
		qbftlog.Warn("LoadProposer address is invalid")
		return nil, errors.New("LoadProposer address is invalid")
	}
	pubkey, err := hex.DecodeString(source.GetPubKey())
	if err != nil {
		qbftlog.Warn("LoadProposer pubkey is invalid")
		return nil, errors.New("LoadProposer pubkey is invalid")
	}

	des := &ttypes.Validator{}
	des.Address = addr
	des.PubKey = pubkey
	des.VotingPower = source.VotingPower
	des.Accum = source.Accum
	return des, nil
}

// CreateBlockInfoTx make blockInfo to the first transaction of the block and execer is qbftNode
func CreateBlockInfoTx(priv crypto.PrivKey, state *tmtypes.QbftState, block *tmtypes.QbftBlock) *types.Transaction {
	blockSave := *block
	blockSave.Data = nil
	blockInfo := &tmtypes.QbftBlockInfo{
		State: state,
		Block: &blockSave,
	}
	qbftlog.Debug("CreateBlockInfoTx", "blockInfo", blockInfo)

	nput := &tmtypes.QbftNodeAction_BlockInfo{BlockInfo: blockInfo}
	action := &tmtypes.QbftNodeAction{Value: nput, Ty: tmtypes.QbftNodeActionBlockInfo}
	tx := &types.Transaction{Execer: []byte("qbftNode"), Payload: types.Encode(action), Fee: fee}
	tx.To = address.ExecAddress("qbftNode")
	tx.Nonce = random.Int63()
	tx.Sign(int32(ttypes.SignMap[signName.Load().(string)]), priv)

	return tx
}
