package executor

import (
	"bytes"
	"hash"
	"math/big"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/zksnark"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/pkg/errors"
)

func makeSetVerifyKeyReceipt(oldKey, newKey *zt.ZkVerifyKey) *types.Receipt {
	key := getVerifyKey(new(big.Int).SetUint64(newKey.GetChainTitleId()).String())
	log := &zt.ReceiptSetVerifyKey{
		Prev:    oldKey,
		Current: newKey,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(newKey)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: zt.TySetVerifyKeyLog, Log: types.Encode(log)},
		},
	}

}

func makeCommitProofReceipt(old, newState *zt.CommitProofState) *types.Receipt {
	key := getLastProofIdKey(new(big.Int).SetUint64(newState.GetChainTitleId()).String())
	log := &zt.ReceiptCommitProof{
		Prev:    old,
		Current: newState,
	}
	r := &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(newState)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: zt.TyCommitProofLog, Log: types.Encode(log)},
		},
	}
	//只在onChainProof 有效时候保存
	if newState.OnChainProofId > 0 {
		onChainIdKey := getLastOnChainProofIdKey(new(big.Int).SetUint64(newState.GetChainTitleId()).String())
		r.KV = append(r.KV, &types.KeyValue{Key: onChainIdKey,
			Value: types.Encode(&zt.LastOnChainProof{ChainTitleId: newState.ChainTitleId, ProofId: newState.ProofId, OnChainProofId: newState.OnChainProofId})})
	}
	return r
}

func makeCommitProofRecordReceipt(proof *zt.CommitProofState, maxRecordId uint64) *types.Receipt {
	key := getProofIdKey(new(big.Int).SetUint64(proof.ChainTitleId).String(), proof.ProofId)
	log := &zt.ReceiptCommitProofRecord{
		Proof: proof,
	}
	r := &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(proof)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: zt.TyCommitProofRecordLog, Log: types.Encode(log)},
		},
	}

	//如果此proofId 比maxRecordId更大，记录下来
	if proof.ProofId > maxRecordId {
		r.KV = append(r.KV, &types.KeyValue{Key: getMaxRecordProofIdKey(new(big.Int).SetUint64(proof.ChainTitleId).String()),
			Value: types.Encode(&types.Int64{Data: int64(proof.ProofId)})})
	}

	return r
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

func isVerifier(statedb dbm.KV, chainTitleId, addr string) bool {
	verifier, err := getVerifierData(statedb, chainTitleId)
	if err != nil {
		if isNotFound(errors.Cause(err)) {
			return false
		} else {
			panic(err)
		}
	}
	for _, v := range verifier.Verifiers {
		if addr == v {
			return true
		}
	}
	return false
}

func getVerifyKeyData(db dbm.KV, chainTitleId string) (*zt.ZkVerifyKey, error) {
	key := getVerifyKey(chainTitleId)
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
	v, ok := new(big.Int).SetString(zt.ZkParaChainInnerTitleId, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "innertitleId=%s", zt.ZkParaChainInnerTitleId)
	}
	chainId := v.Uint64()
	if !cfg.IsPara() {
		if payload.GetChainTitleId() <= 0 {
			return nil, errors.Wrapf(types.ErrInvalidParam, "chain title not set")
		}
		chainId = payload.GetChainTitleId()
	}
	if !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
	}

	oldKey, err := getVerifyKeyData(a.statedb, new(big.Int).SetUint64(chainId).String())
	if isNotFound(errors.Cause(err)) {
		key := &zt.ZkVerifyKey{
			Key:          payload.Key,
			ChainTitleId: chainId,
		}
		return makeSetVerifyKeyReceipt(nil, key), nil
	}

	if err != nil {
		return nil, errors.Wrap(err, "setVerifyKey.getVerifyKeyData")
	}
	newKey := &zt.ZkVerifyKey{
		Key:          payload.Key,
		ChainTitleId: chainId,
	}
	return makeSetVerifyKeyReceipt(oldKey, newKey), nil
}

func getLastCommitProofData(db dbm.KV, titleId string) (*zt.CommitProofState, error) {
	key := getLastProofIdKey(titleId)
	v, err := db.Get(key)
	if err != nil {
		if isNotFound(err) {
			return &zt.CommitProofState{
				ProofId:     0,
				BlockStart:  0,
				BlockEnd:    0,
				IndexStart:  0,
				IndexEnd:    0,
				OldTreeRoot: "0",
				NewTreeRoot: "0",
			}, nil
		} else {
			return nil, errors.Wrapf(err, "get db")
		}
	}
	var data zt.CommitProofState
	err = types.Decode(v, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db")
	}

	return &data, nil
}

func getLastOnChainProofData(db dbm.KV, chainTitle string) (*zt.LastOnChainProof, error) {
	key := getLastOnChainProofIdKey(chainTitle)
	v, err := db.Get(key)
	if err != nil {
		if isNotFound(err) {
			return &zt.LastOnChainProof{OnChainProofId: 0}, nil
		}
		return nil, err
	}
	var data zt.LastOnChainProof
	err = types.Decode(v, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db")
	}
	return &data, nil
}

func getMaxRecordProofIdData(db dbm.KV, chainTitle string) (*types.Int64, error) {
	key := getMaxRecordProofIdKey(chainTitle)
	v, err := db.Get(key)
	if err != nil {
		if isNotFound(err) {
			return &types.Int64{Data: 0}, nil
		}
		return nil, err
	}
	var data types.Int64
	err = types.Decode(v, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db")
	}
	return &data, nil
}

type commitProofCircuit struct {
	//SeqNum, blockStart,blockEnd, oldTreeRoot, newRootHash, full pubData[...]
	PubDataCommitment frontend.Variable `gnark:",public"`

	//SeqNum, blockStart,blockEnd, oldTreeRoot, newRootHash, deposit, partialExit... pubData[...]
	OnChainPubDataCommitment frontend.Variable `gnark:",public"`
}

func (circuit *commitProofCircuit) Define(curveID ecc.ID, api frontend.API) error {
	return nil
}

func getByteBuff(input string) (*bytes.Buffer, error) {
	var buffInput bytes.Buffer
	res, err := common.FromHex(input)
	if err != nil {
		return nil, errors.Wrapf(err, "getByteBuff to %s", input)
	}
	_, err = buffInput.Write(res)
	if err != nil {
		return nil, errors.Wrapf(err, "write buff %s", input)
	}
	return &buffInput, nil

}

func (a *Action) verifyInitRoot(payload *zt.ZkCommitProof) error {
	//只在第一个proof检查
	if payload.ProofId != 1 {
		return nil
	}
	if len(payload.GetCfgFeeAddrs().EthFeeAddr) <= 0 || len(payload.GetCfgFeeAddrs().L2FeeAddr) <= 0 {
		return errors.Wrapf(types.ErrInvalidParam, "1st proofId fee Addr nil, eth=%s,l2=%s", payload.GetCfgFeeAddrs().EthFeeAddr, payload.GetCfgFeeAddrs().L2FeeAddr)
	}

	initRoot := getInitTreeRoot(a.api.GetConfig(), payload.GetCfgFeeAddrs().EthFeeAddr, payload.GetCfgFeeAddrs().L2FeeAddr)
	if initRoot != payload.OldTreeRoot {
		return errors.Wrapf(types.ErrInvalidParam, "calcInitRoot=%s, proof's oldRoot=%s, EthFeeAddrDecimal=%s, L2FeeAddrDecimal=%s",
			initRoot, payload.OldTreeRoot, payload.GetCfgFeeAddrs().EthFeeAddr, payload.GetCfgFeeAddrs().L2FeeAddr)
	}
	return nil
}

func (a *Action) commitProof(payload *zt.ZkCommitProof) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	chainId := zt.ZkParaChainInnerTitleId
	if !cfg.IsPara() {
		if payload.GetChainTitleId() <= 0 {
			return nil, errors.Wrapf(types.ErrInvalidParam, "chainTitle is null")
		}
		chainId = new(big.Int).SetUint64(payload.GetChainTitleId()).String()
	}

	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, chainId, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not validator")
	}
	//如果系统配置了无效证明，则相应证明失效，相继的证明也失效。在exodus mode场景下使用，系统回滚到此证明则丢弃，后面重新接受上一个证明基础上的新的证明
	if isInvalidProof(a.api.GetConfig(), payload.NewTreeRoot) {
		return nil, errors.Wrapf(types.ErrNotAllow, "system cfg invalid proof")
	}
	//基本检查
	/* len(onChainPubdata)     OnChainProofId
	   =0                      =0
	   >0                      >0
	*/
	if (len(payload.OnChainPubDatas) == 0 && payload.GetOnChainProofId() != 0) ||
		(len(payload.OnChainPubDatas) > 0 && payload.GetOnChainProofId() <= 0) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "OnChainData, proofId=%d,onChainProofId=%d,lenOnChain=%d", payload.ProofId, payload.OnChainProofId, len(payload.OnChainPubDatas))
	}
	//验证proofId=1时候的initRoot
	err := a.verifyInitRoot(payload)
	if err != nil {
		zklog.Error("commitProof.verifyInitRoot", "chainId", chainId, "err", err)
		return nil, err
	}

	//1. 先验证proof是否ok
	//get verify key
	verifyKey, err := getVerifyKeyData(a.statedb, chainId)
	if err != nil {
		return nil, errors.Wrapf(err, "get verify key")
	}
	err = verifyProof(verifyKey.Key, payload)
	if err != nil {
		return nil, errors.Wrapf(err, "verify proof")
	}

	//更新数据库, public and proof, pubdata 不上链，存localdb
	chainTitleVal, ok := new(big.Int).SetString(chainId, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "chainTitleVal=%s", chainId)
	}
	newProof := &zt.CommitProofState{
		BlockStart:        payload.BlockStart,
		BlockEnd:          payload.BlockEnd,
		IndexStart:        payload.IndexStart,
		IndexEnd:          payload.IndexEnd,
		OpIndex:           payload.OpIndex,
		ProofId:           payload.ProofId,
		OldTreeRoot:       payload.OldTreeRoot,
		NewTreeRoot:       payload.NewTreeRoot,
		OnChainProofId:    payload.OnChainProofId,
		CommitBlockHeight: a.height,
		ChainTitleId:      chainTitleVal.Uint64(),
	}

	//2. 验证proof是否连续，不连续则暂时保存(考虑交易顺序被打散的场景)
	lastProof, err := getLastCommitProofData(a.statedb, chainId)
	if err != nil {
		return nil, errors.Wrap(err, "get last commit Proof")
	}
	if payload.ProofId < lastProof.ProofId+1 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "commitedId=%d less or equal  lastProofId=%d", payload.ProofId, lastProof.ProofId)
	}
	//get未处理的证明的最大id
	maxRecordId, err := getMaxRecordProofIdData(a.statedb, chainId)
	if err != nil {
		return nil, errors.Wrapf(err, "getMaxRecordProofId for id=%d", payload.ProofId)
	}
	//不连续，先保存数据库,连续时候再验证
	if payload.ProofId > lastProof.ProofId+1 {
		return makeCommitProofRecordReceipt(newProof, uint64(maxRecordId.Data)), nil
	}
	lastOnChainProof, err := getLastOnChainProofData(a.statedb, chainId)
	if err != nil {
		return nil, errors.Wrap(err, "getLastOnChainProof")
	}
	lastOnChainProofId, err := checkNewProof(lastProof, newProof, lastOnChainProof.OnChainProofId)
	if err != nil {
		return nil, errors.Wrapf(err, "checkNewProof id=%d", newProof.ProofId)
	}
	receipt := makeCommitProofReceipt(lastProof, newProof)

	//循环检查可能未处理的recordProof
	lastProof = newProof
	for i := lastProof.ProofId + 1; i < uint64(maxRecordId.Data); i++ {
		recordProof, _ := getRecordProof(a.statedb, chainId, i)
		if recordProof == nil {
			break
		}
		lastOnChainProofId, err = checkNewProof(lastProof, recordProof, lastOnChainProofId)
		if err != nil {
			zklog.Error("commitProof.checkRecordProof", "lastProofId", lastProof.ProofId, "recordProofId", recordProof.ProofId, "err", err)
			//record检查出错，不作为本交易的错误，待下次更新错误的proofId
			break
		}
		mergeReceipt(receipt, makeCommitProofReceipt(lastProof, newProof))
		lastProof = recordProof
	}
	return receipt, nil
}

func getRecordProof(db dbm.KV, title string, id uint64) (*zt.CommitProofState, error) {
	key := getProofIdKey(title, id)
	v, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	var data zt.CommitProofState
	err = types.Decode(v, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db")
	}
	return &data, nil
}

func checkNewProof(lastProof, newProof *zt.CommitProofState, lastOnChainProofId uint64) (uint64, error) {
	//proofId需要连续,区块高度需要衔接
	if lastProof.ProofId+1 != newProof.ProofId || (lastProof.ProofId > 0 && lastProof.BlockEnd != newProof.BlockStart) {
		return lastOnChainProofId, errors.Wrapf(types.ErrInvalidParam, "lastProofId=%d,newProofId=%d, lastBlockEnd=%d,newBlockStart=%d",
			lastProof.ProofId, newProof.ProofId, lastProof.BlockEnd, newProof.BlockStart)
	}

	//tree root 需要衔接, 从proofId=1开始校验
	if lastProof.ProofId > 0 && lastProof.NewTreeRoot != newProof.OldTreeRoot {
		return lastOnChainProofId, errors.Wrapf(types.ErrInvalidParam, "last proof treeRoot=%s, commit oldTreeRoot=%s",
			lastProof.NewTreeRoot, newProof.OldTreeRoot)
	}

	//如果包含OnChainPubData, OnChainProofId需要连续
	if newProof.OnChainProofId > 0 {
		if lastOnChainProofId+1 != newProof.OnChainProofId {
			return lastOnChainProofId, errors.Wrapf(types.ErrInvalidParam, "lastOnChainId not match, lastOnChainId=%d, commit=%d", lastOnChainProofId, newProof.OnChainProofId)
		}
		//更新新的onChainProofId
		return newProof.OnChainProofId, nil
	}
	return lastOnChainProofId, nil
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
	//计算pubData hash 需要和commit的一致
	mimcHash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	commitPubDataHash := proofCircuit.PubDataCommitment.GetWitnessValue(ecc.BN254)
	calcPubDataHash := calcPubDataCommitHash(mimcHash, proof.BlockStart, proof.BlockEnd, proof.ChainTitleId, proof.OldTreeRoot, proof.NewTreeRoot, proof.PubDatas)
	if commitPubDataHash.String() != calcPubDataHash {
		return errors.Wrapf(types.ErrInvalidParam, "pubData hash not match, PI=%s,calc=%s", commitPubDataHash.String(), calcPubDataHash)
	}

	//计算onChain pubData hash 需要和commit的一致
	commitOnChainPubDataHash := proofCircuit.OnChainPubDataCommitment.GetWitnessValue(ecc.BN254)
	calcOnChainPubDataHash := calcOnChainPubDataCommitHash(mimcHash, proof.ChainTitleId, proof.NewTreeRoot, proof.OnChainPubDatas)
	if commitOnChainPubDataHash.String() != calcOnChainPubDataHash {
		return errors.Wrapf(types.ErrInvalidParam, "onChain pubData hash not match, PI=%s,calc=%s", commitOnChainPubDataHash.String(), calcOnChainPubDataHash)
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

func calcPubDataCommitHash(mimcHash hash.Hash, blockStart, blockEnd, chainTitleId uint64, oldRoot, newRoot string, pubDatas []string) string {
	mimcHash.Reset()

	var f fr.Element
	t := f.SetUint64(blockStart).Bytes()
	mimcHash.Write(t[:])

	t = f.SetUint64(blockEnd).Bytes()
	mimcHash.Write(t[:])

	t = f.SetString(oldRoot).Bytes()
	mimcHash.Write(t[:])

	t = f.SetString(newRoot).Bytes()
	mimcHash.Write(t[:])

	t = f.SetUint64(chainTitleId).Bytes()
	mimcHash.Write(t[:])

	for _, r := range pubDatas {
		t = f.SetString(r).Bytes()
		mimcHash.Write(t[:])
	}
	ret := mimcHash.Sum(nil)

	return f.SetBytes(ret).String()
}

func calcOnChainPubDataCommitHash(mimcHash hash.Hash, chainTitleId uint64, newRoot string, pubDatas []string) string {
	mimcHash.Reset()
	var f fr.Element

	t := f.SetString(newRoot).Bytes()
	mimcHash.Write(t[:])
	t = f.SetUint64(chainTitleId).Bytes()
	mimcHash.Write(t[:])

	sum := mimcHash.Sum(nil)

	for _, p := range pubDatas {
		mimcHash.Reset()
		t = f.SetString(p).Bytes()
		mimcHash.Write(sum)
		mimcHash.Write(t[:])
		sum = mimcHash.Sum(nil)
	}

	return f.SetBytes(sum).String()
}

//合约管理员或管理员设置在链上的管理员才可设置
func (a *Action) setVerifier(payload *zt.ZkVerifier) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
	}
	if len(payload.Verifiers) == 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "verifier nil")
	}
	//chainTitle只是在主链有作用，区分不同平行链， 平行链默认为0
	v, ok := new(big.Int).SetString(zt.ZkParaChainInnerTitleId, 10)
	if !ok {
		return nil, errors.Wrapf(types.ErrInvalidParam, "innertitleId=%s", zt.ZkParaChainInnerTitleId)
	}
	chainTitleId := v.Uint64()
	if !cfg.IsPara() {
		if payload.GetChainTitleId() <= 0 {
			return nil, errors.Wrap(types.ErrInvalidParam, "chainTitle or verifier nil")
		}
		chainTitleId = payload.ChainTitleId
	}

	oldKey, err := getVerifierData(a.statedb, new(big.Int).SetUint64(chainTitleId).String())
	if isNotFound(errors.Cause(err)) {
		key := &zt.ZkVerifier{ChainTitleId: chainTitleId, Verifiers: payload.Verifiers}
		return makeSetVerifierReceipt(nil, key), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "setVerifyKey.getVerifyKeyData")
	}
	newKey := &zt.ZkVerifier{ChainTitleId: chainTitleId, Verifiers: payload.Verifiers}
	return makeSetVerifierReceipt(oldKey, newKey), nil
}

func getVerifierData(db dbm.KV, chainTitleId string) (*zt.ZkVerifier, error) {
	key := getVerifier(chainTitleId)
	v, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db verify key")
	}
	var data zt.ZkVerifier
	err = types.Decode(v, &data)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db verify key")
	}

	return &data, nil
}

func makeSetVerifierReceipt(old, newData *zt.ZkVerifier) *types.Receipt {
	key := getVerifier(new(big.Int).SetUint64(newData.ChainTitleId).String())
	log := &zt.ReceiptSetVerifier{
		Prev:    old,
		Current: newData,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(newData)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: zt.TySetVerifierLog, Log: types.Encode(log)},
		},
	}
}
