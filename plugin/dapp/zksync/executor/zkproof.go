package executor

import (
	"bytes"
	"encoding/hex"
	"math/big"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/consensys/gnark-crypto/ecc"

	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"

	"github.com/33cn/plugin/plugin/dapp/mix/executor/zksnark"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/pkg/errors"
)

func makeSetVerifyKeyReceipt(old, new *zt.ZkVerifyKey) *types.Receipt {
	key := getVerifyKey()
	log := &zt.ReceiptSetVerifyKey{
		Prev:    old,
		Current: new,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(new)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: zt.TySetVerifyKey, Log: types.Encode(log)},
		},
	}

}

func makeCommitProofReceipt(old, new *zt.CommitProofState) *types.Receipt {
	key := getVerifyKey()
	heightKey := getHeightCommitProofKey(new.BlockStart)
	log := &zt.ReceiptCommitProof{
		Prev:    old,
		Current: new,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(new)},
			{Key: heightKey, Value: types.Encode(new)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: zt.TyCommitProof, Log: types.Encode(log)},
		},
	}

}

func isNotFound(err error) bool {
	if err != nil && (err == dbm.ErrNotFoundInDb || err == types.ErrNotFound) {
		return true
	}
	return false
}

// IsSuperManager is supper manager or not
func isSuperManager(cfg *types.Chain33Config, addr string) bool {
	confManager := types.ConfSub(cfg, zt.Zksync)
	for _, m := range confManager.GStrList(zt.ZkManagerKey) {
		if addr == m {
			return true
		}
	}
	return false
}

func getVerifyKeyData(db dbm.KV) (*zt.ZkVerifyKey, error) {
	key := getVerifyKey()
	v, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db verify key")
	}
	var data zt.ZkVerifyKey
	err = types.Decode(v, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db verify key")
	}

	return &data, nil
}

//合约管理员或管理员设置在链上的管理员才可设置
func (a *Action) setVerifyKey(payload *zt.ZkVerifyKey) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
	}

	oldKey, err := getVerifyKeyData(a.statedb)
	if isNotFound(errors.Cause(err)) {
		key := &zt.ZkVerifyKey{Key: payload.Key}

		return makeSetVerifyKeyReceipt(nil, key), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "setVerifyKey.getVerifyKeyData")
	}
	newKey := &zt.ZkVerifyKey{Key: payload.Key}
	return makeSetVerifyKeyReceipt(oldKey, newKey), nil
}

func getLastCommitProofData(db dbm.KV) (*zt.CommitProofState, error) {
	key := getLastCommitProofKey()
	v, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db")
	}
	var data zt.CommitProofState
	err = types.Decode(v, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db")
	}

	return &data, nil
}

//manager 设置validator
func isValidator(cfg *types.Chain33Config, addr string) bool {
	return true //TODO

}

type commitProofCircuit struct {
	NewTreeRoot frontend.Variable `gnark:",public"`
	//SeqNum, blockStart,blockEnd, oldTreeRoot, newRootHash, deposit, partialExit... pubData[...]
	OnChainPubDataCommitment frontend.Variable `gnark:",public"`
	//SeqNum, blockStart,blockEnd, oldTreeRoot, newRootHash, full pubData[...]
	PubDataCommitment frontend.Variable `gnark:",public"`
}

func (circuit *commitProofCircuit) Define(curveID ecc.ID, api frontend.API) error {
	return nil
}

func getByteBuff(input string) (*bytes.Buffer, error) {
	var buffInput bytes.Buffer
	res, err := hex.DecodeString(input)
	if err != nil {
		return nil, errors.Wrapf(err, "getByteBuff to %s", input)
	}
	_, err = buffInput.Write(res)
	if err != nil {
		return nil, errors.Wrapf(err, "write buff %s", input)
	}
	return &buffInput, nil

}

//
func (a *Action) commitProof(payload *zt.ZkCommitProof) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if !isValidator(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not validator")
	}

	//按高度索引 proof的 oldTreeRoot和newTreeRoot需要和链上的一致  //TODO

	lastProof, err := getLastCommitProofData(a.statedb)
	if err != nil && !isNotFound(errors.Cause(err)) {
		return nil, errors.Wrap(err, "get last commit Proof")
	}

	//高度需要连续
	if lastProof != nil && lastProof.BlockEnd+1 != payload.BlockStart {
		return nil, errors.Wrapf(types.ErrInvalidParam, "last proof block end=%d, new proof=%d",
			lastProof.BlockEnd, payload.BlockStart)
	}

	//get verify key
	verifyKey, err := getVerifyKeyData(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "get verify key")
	}
	err = verifyProof(verifyKey.Key, payload)
	if err != nil {
		return nil, errors.Wrapf(err, "verify proof")
	}

	//更新数据库, public and proof上链， pubdata 不上链，存localdb
	newProof := &zt.CommitProofState{
		BlockStart:  payload.BlockStart,
		BlockEnd:    payload.BlockEnd,
		OldTreeRoot: payload.OldTreeRoot,
		NewTreeRoot: payload.NewTreeRoot,
		PublicInput: payload.PublicInput,
		Proof:       payload.Proof,
	}
	return makeCommitProofReceipt(lastProof, newProof), nil

}

func verifyProof(verifyKey string, proof *zt.ZkCommitProof) error {

	//decode public inputs
	pBuff, err := getByteBuff(proof.PublicInput)
	if err != nil {
		return errors.Wrapf(err, "read public input str")
	}
	var proofCircuit commitProofCircuit
	_, err = witness.ReadPublicFrom(pBuff, ecc.BN254, &proofCircuit)
	if err != nil {
		return errors.Wrapf(err, "read public input")
	}

	//newTreeRoot， public 和 提交信息一致
	commitNewTreeRoot := proofCircuit.NewTreeRoot.GetWitnessValue(ecc.BN254)
	if proof.NewTreeRoot != commitNewTreeRoot.String() {
		return errors.Wrapf(types.ErrInvalidParam, "PI.newTreeRoot PI=%s", commitNewTreeRoot.String())
	}

	//计算pubData hash 需要和commit的一致
	commitPubDataHash := proofCircuit.PubDataCommitment.GetWitnessValue(ecc.BN254)
	calcPubDataHash := calcPubDataCommitHash(proof.BlockStart, proof.BlockEnd, proof.OldTreeRoot, proof.NewTreeRoot, proof.PubDatas)
	if commitPubDataHash.String() != calcPubDataHash {
		return errors.Wrapf(types.ErrInvalidParam, "PI.pubData PI=%s,calc=%s", commitPubDataHash.String(), calcPubDataHash)
	}

	//验证证明
	ok, err := zksnark.Verify(verifyKey, proof.Proof, proof.PublicInput)
	if err != nil {
		return errors.Wrapf(err, "proof verify error")
	}
	if !ok {
		return errors.Wrapf(types.ErrInvalidParam, "proof verify fail")
	}
	return nil

}

func calcPubDataCommitHash(blockStart, blockEnd uint64, oldRoot, newRoot string, pubDatas []string) string {
	mimcHash := mimc.NewMiMC(zt.ZkMimcHashSeed)

	var v big.Int
	v.SetUint64(blockStart)
	mimcHash.Write(v.Bytes())

	v.SetUint64(blockEnd)
	mimcHash.Write(v.Bytes())

	v.SetString(oldRoot, 10)
	mimcHash.Write(v.Bytes())

	v.SetString(newRoot, 10)
	mimcHash.Write(v.Bytes())

	for _, r := range pubDatas {
		v.SetString(r, 10)
		mimcHash.Write(v.Bytes())
	}
	ret := mimcHash.Sum(nil)

	var f fr.Element
	f.SetBytes(ret)
	return f.String()

}
