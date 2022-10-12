package executor

import (
	"bytes"
	"fmt"
	"hash"
	"math/big"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/zksnark"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
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

	//if lastProof.ProofId+1 != newProof.ProofId {
	//	return lastOnChainProofId, errors.Wrapf(types.ErrInvalidParam, "lastProofId=%d,newProofId=%d, lastBlockEnd=%d,newBlockStart=%d",
	//		lastProof.ProofId, newProof.ProofId, lastProof.BlockEnd, newProof.BlockStart)
	//}

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

func getInitHistoryLeaf(ethFeeAddr, chain33FeeAddr string) []*zt.HistoryLeaf {
	leaves := getInitAccountLeaf(ethFeeAddr, chain33FeeAddr)
	var historyLeaf []*zt.HistoryLeaf
	for _, l := range leaves {
		history := &zt.HistoryLeaf{
			EthAddress:  l.EthAddress,
			AccountId:   l.AccountId,
			Chain33Addr: l.Chain33Addr,
			TokenHash:   zt.Byte2Str(l.TokenHash),
		}
		historyLeaf = append(historyLeaf, history)
	}
	return historyLeaf
}

//暂时保存到全局变量里面
func getHistoryAccountProofFromDb(localdb dbm.KV, chainTitleId uint64, targetRootHash string) *zt.HistoryAccountProofInfo {
	if historyProof.RootHash == targetRootHash {
		return &historyProof
	}
	return nil
}

func setHistoryAccountProofToDb(localdb dbm.KV, chainTitleId uint64, proof *zt.HistoryAccountProofInfo) error {
	historyProof = *proof
	return nil
}

func getHistoryAccountByRoot(localdb dbm.KV, chainTitleId uint64, targetRootHash string) (*zt.HistoryAccountProofInfo, error) {
	info := getHistoryAccountProofFromDb(localdb, chainTitleId, targetRootHash)
	if info != nil {
		return info, nil
	}

	proofTable := NewCommitProofTable(localdb)
	accountMap := make(map[uint64]*zt.HistoryLeaf)
	maxAccountId := uint64(0)
	chainTitleIdStr := new(big.Int).SetUint64(chainTitleId).String()
	rows, err := proofTable.ListIndex("root", getRootCommitProofKey(chainTitleIdStr, targetRootHash), nil, 1, zt.ListASC)
	if err != nil {
		return nil, errors.Wrapf(err, "proofTable.ListIndex")
	}
	proof := rows[0].Data.(*zt.ZkCommitProof)
	if proof == nil {
		return nil, errors.New("proof not exist")
	}

	for i := uint64(1); i <= proof.ProofId; i++ {
		row, err := proofTable.GetData(getProofIdCommitProofKey(chainTitleIdStr, i))
		if err != nil {
			return nil, err
		}
		data := row.Data.(*zt.ZkCommitProof)
		//从第一个proof获取cfgFeeAddr
		if i == uint64(1) {
			initLeaves := getInitHistoryLeaf(data.GetCfgFeeAddrs().EthFeeAddr, data.GetCfgFeeAddrs().L2FeeAddr)
			for _, l := range initLeaves {
				accountMap[l.AccountId] = l
				if maxAccountId < l.AccountId {
					maxAccountId = l.AccountId
				}
			}
		}

		operations := transferPubDatasToOption(data.PubDatas)
		for _, op := range operations {
			switch op.Ty {
			case zt.TyDepositAction:
				operation := op.Op.GetDeposit()
				fromLeaf, ok := accountMap[operation.AccountID]
				if !ok {
					fromLeaf = &zt.HistoryLeaf{
						AccountId:   operation.GetAccountID(),
						EthAddress:  operation.GetEthAddress(),
						Chain33Addr: operation.GetLayer2Addr(),
						Tokens: []*zt.TokenBalance{
							{
								TokenId: operation.TokenID,
								Balance: operation.GetAmount(),
							},
						},
					}
				} else {
					var tokenBalance *zt.TokenBalance
					//找到token
					for _, token := range fromLeaf.Tokens {
						if token.TokenId == operation.TokenID {
							tokenBalance = token
						}
					}
					if tokenBalance == nil {
						tokenBalance = &zt.TokenBalance{
							TokenId: operation.TokenID,
							Balance: operation.Amount,
						}
						fromLeaf.Tokens = append(fromLeaf.Tokens, tokenBalance)
					} else {
						balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
						change, _ := new(big.Int).SetString(operation.Amount, 10)
						tokenBalance.Balance = new(big.Int).Add(balance, change).String()
					}
				}
				accountMap[operation.AccountID] = fromLeaf
				if operation.AccountID > maxAccountId {
					maxAccountId = operation.AccountID
				}
			case zt.TyWithdrawAction:
				operation := op.Op.GetWithdraw()
				fromLeaf, ok := accountMap[operation.AccountID]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
					//add fee
					change = new(big.Int).Add(change, fee)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				accountMap[operation.AccountID] = fromLeaf

			case zt.TyTransferAction:
				operation := op.Op.GetTransfer()
				fromLeaf, ok := accountMap[operation.FromAccountID]
				if !ok {
					return nil, errors.New("account not exist")
				}
				toLeaf, ok := accountMap[operation.ToAccountID]
				if !ok {
					return nil, errors.New("account not exist")
				}

				var fromTokenBalance, toTokenBalance *zt.TokenBalance
				//找到fromToken
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						fromTokenBalance = token
					}
				}
				if fromTokenBalance == nil {
					return nil, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
					//add fee
					change = new(big.Int).Add(change, fee)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("balance not enough")
					}
					fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}

				//找到toToken
				for _, token := range toLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						toTokenBalance = token
					}
				}
				change, _ := new(big.Int).SetString(operation.Amount, 10)
				toFee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
				newToChange := new(big.Int).Sub(change, toFee)
				if toTokenBalance == nil {
					toTokenBalance = &zt.TokenBalance{
						TokenId: operation.TokenID,
						Balance: newToChange.String(),
					}
					toLeaf.Tokens = append(toLeaf.Tokens, toTokenBalance)
				} else {
					balance, _ := new(big.Int).SetString(toTokenBalance.GetBalance(), 10)
					toTokenBalance.Balance = new(big.Int).Add(balance, newToChange).String()
				}
				accountMap[operation.FromAccountID] = fromLeaf
				accountMap[operation.ToAccountID] = toLeaf
			case zt.TyTransferToNewAction:
				operation := op.Op.GetTransferToNew()
				fromLeaf, ok := accountMap[operation.FromAccountID]
				if !ok {
					return nil, errors.New("account not exist")
				}

				var fromTokenBalance *zt.TokenBalance
				//找到fromToken
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						fromTokenBalance = token
					}
				}
				if fromTokenBalance == nil {
					return nil, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
					//add fee
					change = new(big.Int).Add(change, fee)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("balance not enough")
					}
					fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}

				change, _ := new(big.Int).SetString(operation.Amount, 10)
				toFee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
				newToChange := new(big.Int).Sub(change, toFee)
				toLeaf := &zt.HistoryLeaf{
					AccountId:   operation.GetToAccountID(),
					EthAddress:  operation.GetEthAddress(),
					Chain33Addr: operation.GetLayer2Addr(),
					Tokens: []*zt.TokenBalance{
						{
							TokenId: operation.TokenID,
							Balance: newToChange.String(),
						},
					},
				}

				accountMap[operation.FromAccountID] = fromLeaf
				accountMap[operation.ToAccountID] = toLeaf
				if operation.ToAccountID > maxAccountId {
					maxAccountId = operation.ToAccountID
				}
			case zt.TyProxyExitAction:
				operation := op.Op.GetProxyExit()
				//proxy
				fromLeaf, ok := accountMap[operation.ProxyID]
				if !ok {
					return nil, errors.New(fmt.Sprintf("account=%d not exist", operation.ProxyID))
				}

				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New(fmt.Sprintf("token=%d not exist", operation.TokenID))
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.Balance, 10)
					fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
					tokenBalance.Balance = new(big.Int).Sub(balance, fee).String()
				}
				accountMap[operation.ProxyID] = fromLeaf

				//target account
				targetLeaf, ok := accountMap[operation.TargetID]
				if !ok {
					return nil, errors.New(fmt.Sprintf("proxy account=%d not exist", operation.TargetID))
				}
				//找到token
				for _, token := range targetLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New(fmt.Sprintf("proxy target token=%d not exist", operation.TokenID))
				} else {
					if tokenBalance.Balance != operation.Amount {
						return nil, errors.New(fmt.Sprintf("proxy target tokenBalance different"))
					}
					tokenBalance.Balance = "0"
				}
				accountMap[operation.TargetID] = targetLeaf
			case zt.TySetPubKeyAction:
				operation := op.Op.GetSetPubKey()
				fromLeaf, ok := accountMap[operation.AccountID]
				if !ok {
					return nil, errors.New("account not exist")
				}

				pubKey := &zt.ZkPubKey{
					X: operation.PubKey.X,
					Y: operation.PubKey.Y,
				}
				if fromLeaf.ProxyPubKeys == nil {
					fromLeaf.ProxyPubKeys = new(zt.AccountProxyPubKeys)
				}
				switch operation.PubKeyTy {
				case 0:
					fromLeaf.PubKey = pubKey
				case zt.NormalProxyPubKey:
					fromLeaf.ProxyPubKeys.Normal = pubKey
				case zt.SystemProxyPubKey:
					fromLeaf.ProxyPubKeys.System = pubKey
				case zt.SuperProxyPubKey:
					fromLeaf.ProxyPubKeys.Super = pubKey
				default:
					return nil, errors.New(fmt.Sprintf("setPubKey ty=%d not support", operation.PubKeyTy))
				}
				accountMap[operation.AccountID] = fromLeaf

			case zt.TyFullExitAction:
				operation := op.Op.GetFullExit()
				fromLeaf, ok := accountMap[operation.AccountID]
				if !ok {
					return nil, errors.New(fmt.Sprintf("account=%d not exist", operation.AccountID))
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New(fmt.Sprintf("token=%d not exist", operation.TokenID))
				} else {
					tokenBalance.Balance = "0"
				}
				accountMap[operation.AccountID] = fromLeaf

			case zt.TySwapAction:
				operation := op.Op.GetSwap()
				//token: left asset, right asset 顺序
				//电路顺序为：sell-leftAsset, buy+leftAsset, sell-rightAsset-fee, buy+rightAsset-2ndFee
				//这里考虑leaf获取方便，顺序调整为 sell-leftAsset buy+rightAsset-2ndfee, sell-rightAsset-fee,buy+leftAsset
				leftLeaf, ok := accountMap[operation.Left.AccountID]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//找到sell token
				for _, token := range leftLeaf.Tokens {
					if token.TokenId == operation.LeftTokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("taker token not exist")
				}
				balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
				change, _ := new(big.Int).SetString(operation.LeftDealAmount, 10)
				if change.Cmp(balance) > 0 {
					return nil, errors.New("balance not enough")
				}
				tokenBalance.Balance = new(big.Int).Sub(balance, change).String()

				//buy token
				tokenBalance = nil
				for _, token := range leftLeaf.Tokens {
					if token.TokenId == operation.RightTokenID {
						tokenBalance = token
					}
				}
				//taker 2nd fee
				change, _ = new(big.Int).SetString(operation.RightDealAmount, 10)
				fee, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
				if fee.Cmp(change) > 0 {
					return nil, errors.New("change not enough to fee to taker")
				}
				//sub fee
				change = new(big.Int).Sub(change, fee)
				if tokenBalance == nil {
					newToken := &zt.TokenBalance{
						TokenId: operation.RightTokenID,
						Balance: change.String(),
					}
					leftLeaf.Tokens = append(leftLeaf.Tokens, newToken)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				accountMap[operation.Left.AccountID] = leftLeaf

				//toAccount leaf
				rightLeaf, ok := accountMap[operation.Right.AccountID]
				if !ok {
					return nil, errors.New(fmt.Sprintf("right account=%d not exist", operation.Right.AccountID))
				}

				//找到right asset
				tokenBalance = nil
				for _, token := range rightLeaf.Tokens {
					if token.TokenId == operation.RightTokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New(fmt.Sprintf("right sell token=%d not exist", operation.RightTokenID))
				}
				balance, _ = new(big.Int).SetString(tokenBalance.GetBalance(), 10)
				change, _ = new(big.Int).SetString(operation.RightDealAmount, 10)
				if change.Cmp(balance) > 0 {
					return nil, errors.New("maker token balance not enough")
				}
				newBalance := new(big.Int).Sub(balance, change)
				//1st fee
				fee, _ = new(big.Int).SetString(operation.Fee.Fee, 10)
				if fee.Cmp(newBalance) > 0 {
					return nil, errors.New("change not enough to fee to taker")
				}
				tokenBalance.Balance = new(big.Int).Sub(newBalance, fee).String()

				//buy token
				tokenBalance = nil
				for _, token := range rightLeaf.Tokens {
					if token.TokenId == operation.RightTokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					newToken := &zt.TokenBalance{
						TokenId: operation.RightTokenID,
						Balance: operation.RightDealAmount,
					}
					rightLeaf.Tokens = append(rightLeaf.Tokens, newToken)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.RightDealAmount, 10)
					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				accountMap[operation.Right.AccountID] = rightLeaf

			case zt.TyContractToTreeAction:
				operation := op.Op.GetContractToTree()
				fromLeaf, ok := accountMap[zt.SystemTree2ContractAcctId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					panic(fmt.Sprintf("contract2tree system acct balance nil token=%d", operation.TokenID))
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				accountMap[zt.SystemTree2ContractAcctId] = fromLeaf

				//toAccount
				toLeaf, ok := accountMap[operation.AccountID]
				if !ok {
					return nil, errors.Wrapf(types.ErrAccountNotExist, "ty=%d,toAccountId=%d not exist", zt.TyContractToTreeAction, operation.AccountID)
				}
				//找到toToken
				var toTokenBalance *zt.TokenBalance
				for _, token := range toLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						toTokenBalance = token
					}
				}
				change, _ := new(big.Int).SetString(operation.Amount, 10)
				if toTokenBalance == nil {
					toTokenBalance = &zt.TokenBalance{
						TokenId: operation.TokenID,
						Balance: change.String(),
					}
					toLeaf.Tokens = append(toLeaf.Tokens, toTokenBalance)
				} else {
					balance, _ := new(big.Int).SetString(toTokenBalance.GetBalance(), 10)
					toTokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				accountMap[operation.AccountID] = toLeaf

			case zt.TyContractToTreeNewAction:
				operation := op.Op.GetContract2TreeNew()
				fromLeaf, ok := accountMap[zt.SystemTree2ContractAcctId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					panic(fmt.Sprintf("contract2treeNew system acct balance nil token=%d", operation.TokenID))
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				accountMap[zt.SystemTree2ContractAcctId] = fromLeaf

				//to leaf
				change, _ := new(big.Int).SetString(operation.Amount, 10)
				toLeaf := &zt.HistoryLeaf{
					AccountId:   operation.ToAccountID,
					EthAddress:  operation.GetEthAddress(),
					Chain33Addr: operation.GetLayer2Addr(),
					Tokens: []*zt.TokenBalance{
						{
							TokenId: operation.TokenID,
							Balance: change.String(),
						},
					},
				}
				accountMap[operation.ToAccountID] = toLeaf
				if operation.ToAccountID > maxAccountId {
					maxAccountId = operation.ToAccountID
				}

			case zt.TyTreeToContractAction:
				operation := op.Op.GetTreeToContract()
				fromLeaf, ok := accountMap[operation.AccountID]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				accountMap[operation.AccountID] = fromLeaf

			case zt.TyFeeAction:
				operation := op.Op.GetFee()
				fromLeaf, ok := accountMap[operation.AccountID]
				if !ok {
					return nil, errors.New(fmt.Sprintf("fee accountId=%d not exist", operation.AccountID))
				}
				if operation.AccountID != zt.SystemFeeAccountId {
					return nil, errors.New(fmt.Sprintf("fee accountId=%d not systmeId=%d", operation.AccountID, zt.SystemFeeAccountId))
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					newToken := &zt.TokenBalance{
						TokenId: operation.TokenID,
						Balance: operation.Amount,
					}
					fromLeaf.Tokens = append(fromLeaf.Tokens, newToken)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)

					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				accountMap[operation.AccountID] = fromLeaf

			case zt.TyMintNFTAction:
				operation := op.Op.GetMintNFT()
				fromLeaf, ok := accountMap[operation.MintAcctID]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//1. fee
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.Fee.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("mint nft token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("mint nft fee balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}

				//2. creator systemNFTtokenId+1 get serialId
				tokenBalance = nil
				serialId := "0"
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == zt.SystemNFTTokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					newToken := &zt.TokenBalance{
						TokenId: zt.SystemNFTTokenId,
						Balance: "1",
					}
					fromLeaf.Tokens = append(fromLeaf.Tokens, newToken)
				} else {
					//before balance as the serialId
					serialId = tokenBalance.GetBalance()
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					tokenBalance.Balance = new(big.Int).Add(balance, big.NewInt(1)).String()
				}

				//3. systemNFT accountId's systemNFTTokenId balance+1 get newTokenId
				newNFTTokenId := uint64(0)
				systemNFTLeaf, ok := accountMap[zt.SystemNFTAccountId]
				if !ok {
					return nil, errors.New("SystemNFTAccountId not found")
				}
				tokenBalance = nil
				for _, token := range systemNFTLeaf.Tokens {
					if token.TokenId == zt.SystemNFTTokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					newNFTTokenId = new(big.Int).Add(big.NewInt(zt.SystemNFTTokenId), big.NewInt(1)).Uint64()
					newToken := &zt.TokenBalance{
						TokenId: zt.SystemNFTTokenId,
						Balance: new(big.Int).Add(big.NewInt(zt.SystemNFTTokenId), big.NewInt(2)).String(),
					}
					systemNFTLeaf.Tokens = append(systemNFTLeaf.Tokens, newToken)
				} else {
					//before balance as the new token Id
					b, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					newNFTTokenId = b.Uint64()
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					tokenBalance.Balance = new(big.Int).Add(balance, big.NewInt(1)).String()
				}

				//4. set new system NFT tokenId balance
				//检查systemNFTAccount 没有此newNFTTokenId
				tokenBalance = nil
				for _, token := range systemNFTLeaf.Tokens {
					if token.TokenId == newNFTTokenId {
						tokenBalance = token
					}
				}
				if tokenBalance != nil {
					return nil, errors.New(fmt.Sprintf("systemNFTAccount has the newNFT id=%d", newNFTTokenId))
				}
				mintAmount, _ := new(big.Int).SetString(operation.Amount, 10)
				newSysNFTTokenBalance, err := getNewNFTTokenBalance(operation.MintAcctID, serialId, operation.ErcProtocol, mintAmount.Uint64(),
					operation.ContentHash[0], operation.ContentHash[1])
				if err != nil {
					return nil, errors.Wrapf(err, "newNFTTokenBalance")
				}
				newToken := &zt.TokenBalance{
					TokenId: newNFTTokenId,
					Balance: newSysNFTTokenBalance,
				}
				systemNFTLeaf.Tokens = append(systemNFTLeaf.Tokens, newToken)
				accountMap[zt.SystemNFTAccountId] = systemNFTLeaf

				//5. recipient id
				toLeaf, ok := accountMap[operation.RecipientID]
				if !ok {
					return nil, errors.New("account not exist")
				}
				tokenBalance = nil
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.Fee.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance != nil {
					return nil, errors.New("nft recipient nft token  existed")
				}
				newToken = &zt.TokenBalance{
					TokenId: newNFTTokenId,
					Balance: operation.Amount,
				}
				toLeaf.Tokens = append(toLeaf.Tokens, newToken)

			case zt.TyWithdrawNFTAction:
				operation := op.Op.GetWithdrawNFT()
				fromLeaf, ok := accountMap[operation.FromAcctID]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//1. fee
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.Fee.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("withdraw nft fee token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("withdraw nft fee balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				//2. NFT token balance-amount
				tokenBalance = nil
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.NFTTokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("withdraw nft token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.WithdrawAmount, 10)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("withdraw nft  balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
			case zt.TyTransferNFTAction:
				operation := op.Op.GetTransferNFT()
				fromLeaf, ok := accountMap[operation.FromAccountID]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//1. fee
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.Fee.TokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("transfer nft fee token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Fee.Fee, 10)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("withdraw nft fee balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				//2. NFT token balance-amount
				tokenBalance = nil
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.NFTTokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("transfer nft token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("transfer nft  balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}

				toLeaf, ok := accountMap[operation.RecipientID]
				if !ok {
					return nil, errors.New("account not exist")
				}

				// NFT token balance+amount
				tokenBalance = nil
				for _, token := range toLeaf.Tokens {
					if token.TokenId == operation.NFTTokenID {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					newToken := &zt.TokenBalance{
						TokenId: operation.NFTTokenID,
						Balance: operation.Amount,
					}
					toLeaf.Tokens = append(toLeaf.Tokens, newToken)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
			default:
				return nil, errors.New(fmt.Sprintf("not support op ty=%d", op.Ty))
			}
		}
	}

	historyAccounts := &zt.HistoryAccountProofInfo{RootHash: targetRootHash}
	for i := uint64(zt.SystemFeeAccountId); i <= maxAccountId; i++ {
		if _, ok := accountMap[i]; !ok {
			return nil, errors.Wrapf(err, "accountId=%d not exist", i)
		}
		historyAccounts.Leaves = append(historyAccounts.Leaves, accountMap[i])
		historyAccounts.LeafHashes = append(historyAccounts.LeafHashes, getHistoryLeafHash(accountMap[i]))
	}

	//验证leafHash和rootHash是否匹配
	accountMerkleProof, err := getMerkleTreeProof(zt.SystemFeeAccountId, historyAccounts.LeafHashes)
	if err != nil {
		return nil, errors.Wrapf(err, "account.getMerkleTreeProof")
	}
	if accountMerkleProof.RootHash != targetRootHash {
		return nil, errors.Wrapf(types.ErrInvalidParam, "calc root=%s,expect=%s", accountMerkleProof.RootHash, targetRootHash)
	}

	err = setHistoryAccountProofToDb(localdb, chainTitleId, historyAccounts)
	if err != nil {
		zklog.Error("setHistoryAccountProofToDb", "err", err)
	}
	return historyAccounts, nil
}

func GetHistoryAccountProof(historyAccountInfo *zt.HistoryAccountProofInfo, targetAccountId, targetTokenId uint64) (*zt.ZkProofWitness, error) {
	if targetAccountId > uint64(len(historyAccountInfo.Leaves)) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "targetAccountId=%d not exist", targetAccountId)
	}
	targetLeaf := historyAccountInfo.Leaves[targetAccountId-1]

	var tokenFound bool
	var tokenIndex int
	for i, t := range targetLeaf.Tokens {
		if t.TokenId == targetTokenId {
			tokenIndex = i
			tokenFound = true
			break
		}
	}
	if !tokenFound {
		return nil, errors.Wrapf(types.ErrInvalidParam, "accountId=%d has no asset tokenId=%d", targetAccountId, targetTokenId)
	}

	accountMerkleProof, err := getMerkleTreeProof(targetAccountId-1, historyAccountInfo.LeafHashes)
	if err != nil {
		return nil, errors.Wrapf(err, "account.getMerkleTreeProof")
	}
	if accountMerkleProof.RootHash != historyAccountInfo.RootHash {
		return nil, errors.Wrapf(types.ErrInvalidParam, "calc root=%s,expect=%s", accountMerkleProof.RootHash, historyAccountInfo.RootHash)
	}

	//token proof
	var tokenHashes [][]byte
	for _, token := range targetLeaf.Tokens {
		tokenHashes = append(tokenHashes, getTokenBalanceHash(token))
	}
	tokenMerkleProof, err := getMerkleTreeProof(uint64(tokenIndex), tokenHashes)
	if err != nil {
		return nil, errors.Wrapf(err, "token.getMerkleProof")
	}

	accTreePath := &zt.SiblingPath{
		Path:   accountMerkleProof.ProofSet,
		Helper: accountMerkleProof.Helpers,
	}
	accountW := &zt.AccountWitness{
		ID:            targetAccountId,
		EthAddr:       targetLeaf.EthAddress,
		Chain33Addr:   targetLeaf.Chain33Addr,
		TokenTreeRoot: tokenMerkleProof.RootHash,
		PubKey:        targetLeaf.PubKey,
		ProxyPubKeys:  targetLeaf.ProxyPubKeys,
		Sibling:       accTreePath,
	}

	tokenTreePath := &zt.SiblingPath{
		Path:   tokenMerkleProof.ProofSet,
		Helper: tokenMerkleProof.Helpers,
	}
	tokenW := &zt.TokenWitness{
		ID:      targetTokenId,
		Balance: targetLeaf.Tokens[tokenIndex].Balance,
		Sibling: tokenTreePath,
	}
	var witness zt.ZkProofWitness
	witness.AccountWitness = accountW
	witness.TokenWitness = tokenW
	witness.TreeRoot = historyProof.RootHash

	return &witness, nil
}

//根据rootHash获取account在该root下的证明
func getAccountProofInHistory(localdb dbm.KV, req *zt.ZkReqExistenceProof) (*zt.ZkProofWitness, error) {
	historyAccountInfo, err := getHistoryAccountByRoot(localdb, req.ChainTitleId, req.RootHash)
	if err != nil {
		return nil, err
	}
	return GetHistoryAccountProof(historyAccountInfo, req.AccountId, req.TokenId)
}

func getMerkleTreeProof(index uint64, hashes [][]byte) (*zt.MerkleTreeProof, error) {
	tree := getNewTree()
	err := tree.SetIndex(index)
	if err != nil {
		return nil, errors.Wrapf(err, "tree.SetIndex")
	}
	for _, h := range hashes {
		tree.Push(h)
	}
	rootHash, proofSet, proofIndex, numLeaves := tree.Prove()
	helpers := make([]string, 0)
	proofStringSet := make([]string, 0)
	for _, v := range merkletree.GenerateProofHelper(proofSet, proofIndex, numLeaves) {
		helpers = append(helpers, big.NewInt(int64(v)).String())
	}
	for _, v := range proofSet {
		proofStringSet = append(proofStringSet, zt.Byte2Str(v))
	}
	return &zt.MerkleTreeProof{RootHash: zt.Byte2Str(rootHash), ProofSet: proofStringSet, Helpers: helpers}, nil
}

func transferPubDatasToOption(pubDatas []string) []*zt.ZkOperation {
	operations := make([]*zt.ZkOperation, 0)
	start := 0
	for start < len(pubDatas) {
		chunk := wallet.ChunkStringToByte(pubDatas[start])
		operationTy := getTyByChunk(chunk)
		chunkNum := getChunkNum(operationTy)
		if operationTy != zt.TyNoopAction {
			operation := getOperationByChunk(pubDatas[start:start+chunkNum], operationTy)
			fmt.Println("transferPubDatasToOption.op=", operation)
			operations = append(operations, operation)
		}
		start = start + chunkNum
	}
	return operations
}

func getTyByChunk(chunk []byte) uint64 {
	return new(big.Int).SetBytes(chunk[:1]).Uint64()
}

func getChunkNum(opType uint64) int {
	switch opType {
	case zt.TyNoopAction:
		return zt.NoopChunks
	case zt.TyDepositAction:
		return zt.DepositChunks
	case zt.TyWithdrawAction:
		return zt.WithdrawChunks
	case zt.TyTransferAction:
		return zt.TransferChunks
	case zt.TyTransferToNewAction:
		return zt.Transfer2NewChunks
	case zt.TyProxyExitAction:
		return zt.ProxyExitChunks
	case zt.TySetPubKeyAction:
		return zt.SetPubKeyChunks
	case zt.TyFullExitAction:
		return zt.FullExitChunks
	case zt.TySwapAction:
		return zt.SwapChunks
	case zt.TyContractToTreeAction:
		return zt.Contract2TreeChunks
	case zt.TyTreeToContractAction:
		return zt.Tree2ContractChunks
	case zt.TyFeeAction:
		return zt.FeeChunks
	case zt.TyMintNFTAction:
		return zt.MintNFTChunks
	case zt.TyWithdrawNFTAction:
		return zt.WithdrawNFTChunks
	case zt.TyTransferNFTAction:
		return zt.TransferNFTChunks

	default:
		panic(fmt.Sprintf("operation tx type=%d not support", opType))
	}

}

func getOperationByChunk(chunks []string, optionTy uint64) *zt.ZkOperation {
	totalChunk := make([]byte, 0)
	for _, chunk := range chunks {
		totalChunk = append(totalChunk, wallet.ChunkStringToByte(chunk)...)
	}
	switch optionTy {
	case zt.TyDepositAction:
		return getDepositOperationByChunk(totalChunk)
	case zt.TyWithdrawAction:
		return getWithDrawOperationByChunk(totalChunk)
	case zt.TyTransferAction:
		return getTransferOperationByChunk(totalChunk)
	case zt.TyTransferToNewAction:
		return getTransfer2NewOperationByChunk(totalChunk)
	case zt.TyProxyExitAction:
		return getProxyExitOperationByChunk(totalChunk)
	case zt.TySetPubKeyAction:
		return getSetPubKeyOperationByChunk(totalChunk)
	case zt.TyFullExitAction:
		return getFullExitOperationByChunk(totalChunk)
	case zt.TySwapAction:
		return getSwapOperationByChunk(totalChunk)
	case zt.TyContractToTreeAction:
		return getContract2TreeOptionByChunk(totalChunk)
	case zt.TyTreeToContractAction:
		return getTree2ContractOperationByChunk(totalChunk)
	case zt.TyFeeAction:
		return getFeeOperationByChunk(totalChunk)
	case zt.TyMintNFTAction:
		return getMintNFTOperationByChunk(totalChunk)
	case zt.TyWithdrawNFTAction:
		return getWithdrawNFTOperationByChunk(totalChunk)
	case zt.TyTransferNFTAction:
		return getTransferNFTOperationByChunk(totalChunk)
	default:
		panic("operationTy not support")
	}
}

//根据proofId重建merkleTree
//func saveHistoryAccountTree(localdb dbm.KV, endProofId uint64) ([]*types.KeyValue, error) {
//	var localKvs []*types.KeyValue
//	proofTable := NewCommitProofTable(localdb)
//	historyTable := NewHistoryAccountTreeTable(localdb)
//	//todo 多少ID归档一次实现可配置化
//	historyId := (endProofId/10 - 1) * 10
//	for i := historyId + 1; i <= endProofId; i++ {
//		row, err := proofTable.GetData(getProofIdCommitProofKey(i))
//		if err != nil {
//			return localKvs, err
//		}
//		data := row.Data.(*zt.ZkCommitProof)
//		operations := transferPubDatasToOption(data.PubDatas)
//		for _, operation := range operations {
//			fromLeaf, err := getAccountByProofIdAndHistoryId(historyTable, endProofId, historyId, operation.GetAccountId())
//			if err != nil {
//				return localKvs, errors.Wrapf(err, "getAccountByProofIdAndHistoryId")
//			}
//			switch operation.Ty {
//			case zt.TyDepositAction:
//				if fromLeaf == nil {
//					fromLeaf = &zt.HistoryLeaf{
//						AccountId:   operation.GetAccountId(),
//						EthAddress:  operation.GetEthAddress(),
//						Chain33Addr: operation.GetChain33Addr(),
//						ProofId:     endProofId,
//						Tokens: []*zt.TokenBalance{
//							{
//								TokenId: operation.TokenId,
//								Balance: operation.GetAmount(),
//							},
//						},
//					}
//				} else {
//					fromLeaf.ProofId = endProofId
//					var tokenBalance *zt.TokenBalance
//					//找到token
//					for _, token := range fromLeaf.Tokens {
//						if token.TokenId == operation.TokenId {
//							tokenBalance = token
//						}
//					}
//					if tokenBalance == nil {
//						tokenBalance = &zt.TokenBalance{
//							TokenId: operation.TokenId,
//							Balance: operation.Amount,
//						}
//						fromLeaf.Tokens = append(fromLeaf.Tokens, tokenBalance)
//					} else {
//						balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
//						change, _ := new(big.Int).SetString(operation.Amount, 10)
//						tokenBalance.Balance = new(big.Int).Add(balance, change).String()
//					}
//				}
//				err = historyTable.Replace(fromLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//			case zt.TyWithdrawAction:
//				if fromLeaf == nil {
//					return localKvs, errors.New("account not exist")
//				}
//				fromLeaf.ProofId = endProofId
//				var tokenBalance *zt.TokenBalance
//				//找到token
//				for _, token := range fromLeaf.Tokens {
//					if token.TokenId == operation.TokenId {
//						tokenBalance = token
//					}
//				}
//				if tokenBalance == nil {
//					return localKvs, errors.New("token not exist")
//				} else {
//					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
//					change, _ := new(big.Int).SetString(operation.Amount, 10)
//					if change.Cmp(balance) > 0 {
//						return localKvs, errors.New("balance not enough")
//					}
//					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
//				}
//				err = historyTable.Replace(fromLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//			case zt.TyTreeToContractAction:
//				if fromLeaf == nil {
//					return localKvs, errors.New("account not exist")
//				}
//				fromLeaf.ProofId = endProofId
//				var tokenBalance *zt.TokenBalance
//				//找到token
//				for _, token := range fromLeaf.Tokens {
//					if token.TokenId == operation.TokenId {
//						tokenBalance = token
//					}
//				}
//				if tokenBalance == nil {
//					return localKvs, errors.New("token not exist")
//				} else {
//					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
//					change, _ := new(big.Int).SetString(operation.Amount, 10)
//					if change.Cmp(balance) > 0 {
//						return localKvs, errors.New("balance not enough")
//					}
//					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
//				}
//				err = historyTable.Replace(fromLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//			case zt.TyContractToTreeAction:
//				if fromLeaf == nil {
//					return localKvs, errors.New("account not exist")
//				}
//				fromLeaf.ProofId = endProofId
//				var tokenBalance *zt.TokenBalance
//				//找到token
//				for _, token := range fromLeaf.Tokens {
//					if token.TokenId == operation.TokenId {
//						tokenBalance = token
//					}
//				}
//				if tokenBalance == nil {
//					tokenBalance = &zt.TokenBalance{
//						TokenId: operation.TokenId,
//						Balance: operation.Amount,
//					}
//					fromLeaf.Tokens = append(fromLeaf.Tokens, tokenBalance)
//				} else {
//					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
//					change, _ := new(big.Int).SetString(operation.Amount, 10)
//					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
//				}
//				err = historyTable.Replace(fromLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//			case zt.TyTransferAction:
//				toLeaf, err := getAccountByProofIdAndHistoryId(historyTable, endProofId, historyId, operation.GetAccountId())
//				if err != nil {
//					return localKvs, errors.Wrapf(err, "getAccountByProofIdAndHistoryId")
//				}
//				if fromLeaf == nil || toLeaf == nil {
//					return localKvs, errors.New("account not exist")
//				}
//
//				fromLeaf.ProofId = endProofId
//				toLeaf.ProofId = endProofId
//
//				var fromTokenBalance, toTokenBalance *zt.TokenBalance
//				//找到fromToken
//				for _, token := range fromLeaf.Tokens {
//					if token.TokenId == operation.TokenId {
//						fromTokenBalance = token
//					}
//				}
//				if fromTokenBalance == nil {
//					return localKvs, errors.New("token not exist")
//				} else {
//					balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
//					change, _ := new(big.Int).SetString(operation.Amount, 10)
//					if change.Cmp(balance) > 0 {
//						return localKvs, errors.New("balance not enough")
//					}
//					fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
//				}
//
//				//找到toToken
//				for _, token := range toLeaf.Tokens {
//					if token.TokenId == operation.TokenId {
//						toTokenBalance = token
//					}
//				}
//				if toTokenBalance == nil {
//					toTokenBalance = &zt.TokenBalance{
//						TokenId: operation.TokenId,
//						Balance: operation.Amount,
//					}
//					toLeaf.Tokens = append(toLeaf.Tokens, toTokenBalance)
//				} else {
//					balance, _ := new(big.Int).SetString(toTokenBalance.GetBalance(), 10)
//					change, _ := new(big.Int).SetString(operation.Amount, 10)
//					toTokenBalance.Balance = new(big.Int).Add(balance, change).String()
//				}
//
//				err = historyTable.Replace(fromLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//				err = historyTable.Replace(toLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//			case zt.TyTransferToNewAction:
//				if fromLeaf == nil {
//					return localKvs, errors.New("account not exist")
//				}
//
//				fromLeaf.ProofId = endProofId
//
//				var fromTokenBalance *zt.TokenBalance
//				//找到fromToken
//				for _, token := range fromLeaf.Tokens {
//					if token.TokenId == operation.TokenId {
//						fromTokenBalance = token
//					}
//				}
//				if fromTokenBalance == nil {
//					return localKvs, errors.New("token not exist")
//				} else {
//					balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
//					change, _ := new(big.Int).SetString(operation.Amount, 10)
//					if change.Cmp(balance) > 0 {
//						return localKvs, errors.New("balance not enough")
//					}
//					fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
//				}
//
//				toLeaf := &zt.HistoryLeaf{
//					AccountId:   operation.GetToAccountId(),
//					EthAddress:  operation.GetEthAddress(),
//					Chain33Addr: operation.GetChain33Addr(),
//					ProofId:     endProofId,
//					Tokens: []*zt.TokenBalance{
//						{
//							TokenId: operation.TokenId,
//							Balance: operation.GetAmount(),
//						},
//					},
//				}
//
//				err = historyTable.Replace(fromLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//				err = historyTable.Replace(toLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//			case zt.TySetPubKeyAction:
//				if fromLeaf == nil {
//					return localKvs, errors.New("account not exist")
//				}
//				fromLeaf.ProofId = endProofId
//				fromLeaf.PubKey = &zt.ZkPubKey{
//					X: operation.SetPubKey.PubKey.X,
//					Y: operation.SetPubKey.PubKey.Y,
//				}
//				err = historyTable.Replace(fromLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//			case zt.TyProxyExitAction:
//				if fromLeaf == nil {
//					return localKvs, errors.New("account not exist")
//				}
//				fromLeaf.ProofId = endProofId
//				var tokenBalance *zt.TokenBalance
//				//找到token
//				for _, token := range fromLeaf.Tokens {
//					if token.TokenId == operation.TokenId {
//						tokenBalance = token
//					}
//				}
//				if tokenBalance == nil {
//					return localKvs, errors.New("token not exist")
//				} else {
//					tokenBalance.Balance = "0"
//				}
//				err = historyTable.Replace(fromLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//			case zt.TyFullExitAction:
//				if fromLeaf == nil {
//					return localKvs, errors.New("account not exist")
//				}
//				fromLeaf.ProofId = endProofId
//				var tokenBalance *zt.TokenBalance
//				//找到token
//				for _, token := range fromLeaf.Tokens {
//					if token.TokenId == operation.TokenId {
//						tokenBalance = token
//					}
//				}
//				if tokenBalance == nil {
//					return localKvs, errors.New("token not exist")
//				} else {
//					tokenBalance.Balance = "0"
//				}
//				err = historyTable.Replace(fromLeaf)
//				if err != nil {
//					return localKvs, err
//				}
//			}
//		}
//	}
//	localKvs, err := historyTable.Save()
//	if err != nil {
//		return localKvs, err
//	}
//	return localKvs, nil
//}

//首先通过当前proofId去拿，如果没拿到，使用历史id去拿，因为同一个accountId会被多次更新
func getAccountByProofIdAndHistoryId(historyTable *table.Table, currentId, historyId, accountId uint64) (*zt.HistoryLeaf, error) {
	row, err := historyTable.GetData(getHistoryAccountTreeKey(currentId, accountId))
	if err != nil {
		if isNotFound(err) {
			row, err = historyTable.GetData(getHistoryAccountTreeKey(historyId, accountId))
			if err != nil {
				if isNotFound(err) {
					return nil, nil
				} else {
					return nil, err
				}
			}
			data := row.Data.(*zt.HistoryLeaf)
			return data, nil
		} else {
			return nil, err
		}
	}
	data := row.Data.(*zt.HistoryLeaf)
	return data, nil
}

func getDepositOperationByChunk(chunk []byte) *zt.ZkOperation {
	deposit := &zt.ZkDepositWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	deposit.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	deposit.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	deposit.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	deposit.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/8
	deposit.Layer2Addr = zt.Byte2Str(chunk[start:end])

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Deposit{Deposit: deposit}}
	return &zt.ZkOperation{Ty: zt.TyDepositAction, Op: special}
}

func getWithDrawOperationByChunk(chunk []byte) *zt.ZkOperation {
	withdraw := &zt.ZkWithdrawWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	withdraw.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	withdraw.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	withdraw.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	withdraw.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	withdraw.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Withdraw{Withdraw: withdraw}}
	return &zt.ZkOperation{Ty: zt.TyWithdrawAction, Op: special}
}

func getSwapOperationByChunk(chunk []byte) *zt.ZkOperation {
	leftOrder := &zt.ZkSwapOrderInfo{}
	rightOrder := &zt.ZkSwapOrderInfo{}
	operation := &zt.ZkSwapWitnessInfo{Left: leftOrder, Right: rightOrder, Fee: &zt.ZkFee{}}

	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	leftOrder.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	rightOrder.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	//1st token, left asset
	operation.LeftTokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	//2nd token, right asset
	operation.RightTokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	//1st amount, left asset amount
	operation.LeftDealAmount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	//2nd amount right asset amount
	operation.RightDealAmount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	//1st fee, left's fee
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	//2nd fee, right's fee
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Swap{Swap: operation}}
	return &zt.ZkOperation{Ty: zt.TySwapAction, Op: special}
}

func getContract2TreeOptionByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkContractToTreeWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_ContractToTree{ContractToTree: operation}}
	return &zt.ZkOperation{Ty: zt.TyContractToTreeAction, Op: special}
}

func getTree2ContractOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkTreeToContractWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TreeToContract{TreeToContract: operation}}
	return &zt.ZkOperation{Ty: zt.TyTreeToContractAction, Op: special}
}

func getTransferOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkTransferWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.FromAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.ToAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Transfer{Transfer: operation}}
	return &zt.ZkOperation{Ty: zt.TyTransferAction, Op: special}
}

func getTransfer2NewOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkTransferToNewWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.FromAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.ToAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/8
	operation.Layer2Addr = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TransferToNew{TransferToNew: operation}}
	return &zt.ZkOperation{Ty: zt.TyTransferToNewAction, Op: special}
}

func getSetPubKeyOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkSetPubKeyWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TxTypeBitWidth/8
	operation.PubKeyTy = zt.Byte2Uint64(chunk[start:end])
	pubkey := &zt.ZkPubKey{}
	start = end
	end = start + zt.PubKeyBitWidth/8
	pubkey.X = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.PubKeyBitWidth/8
	pubkey.Y = zt.Byte2Str(chunk[start:end])
	operation.PubKey = pubkey

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_SetPubKey{SetPubKey: operation}}
	return &zt.ZkOperation{Ty: zt.TySetPubKeyAction, Op: special}
}

func getProxyExitOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkProxyExitWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	//proxy id
	operation.ProxyID = zt.Byte2Uint64(chunk[start:end])
	start = zt.TxTypeBitWidth / 8
	end = start + zt.AccountBitWidth/8
	//toId
	operation.TargetID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_ProxyExit{ProxyExit: operation}}
	return &zt.ZkOperation{Ty: zt.TyProxyExitAction, Op: special}
}

func getFullExitOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkFullExitWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_FullExit{FullExit: operation}}
	return &zt.ZkOperation{Ty: zt.TyFullExitAction, Op: special}
}

func getFeeOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkFeeWitnessInfo{}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_Fee{Fee: operation}}
	return &zt.ZkOperation{Ty: zt.TyFeeAction, Op: special}
}

func getMintNFTOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkMintNFTWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.MintAcctID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.RecipientID = zt.Byte2Uint64(chunk[start:end])
	//ERC 721/1155 protocol
	start = end
	end = start + zt.TxTypeBitWidth/8
	operation.ErcProtocol = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.ContentHash = append(operation.ContentHash, zt.Byte2Str(chunk[start:end]))
	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.ContentHash = append(operation.ContentHash, zt.Byte2Str(chunk[start:end]))

	start = end
	end = start + zt.TokenBitWidth/8
	operation.Fee.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_MintNFT{MintNFT: operation}}
	return &zt.ZkOperation{Ty: zt.TyMintNFTAction, Op: special}
}

func getWithdrawNFTOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkWithdrawNFTWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	//fromId
	operation.FromAcctID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	//original creator id
	operation.CreatorAcctID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.NFTTokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.CreatorSerialID = zt.Byte2Uint64(chunk[start:end])
	//ERC 721/1155 protocol
	start = end
	end = start + zt.TxTypeBitWidth/8
	operation.ErcProtocol = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.InitMintAmount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.WithdrawAmount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])

	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.ContentHash = append(operation.ContentHash, zt.Byte2Str(chunk[start:end]))
	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.ContentHash = append(operation.ContentHash, zt.Byte2Str(chunk[start:end]))

	start = end
	end = start + zt.TokenBitWidth/8
	operation.Fee.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_WithdrawNFT{WithdrawNFT: operation}}
	return &zt.ZkOperation{Ty: zt.TyWithdrawNFTAction, Op: special}
}

func getTransferNFTOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkTransferNFTWitnessInfo{Fee: &zt.ZkFee{}}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.FromAccountID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.RecipientID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.NFTTokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])

	start = end
	end = start + zt.TokenBitWidth/8
	operation.Fee.TokenID = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacExpBitWidth)/8
	operation.Fee.Fee = zt.DecodePacVal(chunk[start:end], zt.PacExpBitWidth)

	special := &zt.OperationSpecialInfo{Value: &zt.OperationSpecialInfo_TransferNFT{TransferNFT: operation}}
	return &zt.ZkOperation{Ty: zt.TyTransferNFTAction, Op: special}
}
