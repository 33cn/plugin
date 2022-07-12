package executor

import (
	"bytes"
	"fmt"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
	"hash"
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
	if payload.GetChainTitleId() <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "chain title not set")
	}
	if !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
	}

	oldKey, err := getVerifyKeyData(a.statedb, new(big.Int).SetUint64(payload.GetChainTitleId()).String())
	if isNotFound(errors.Cause(err)) {
		key := &zt.ZkVerifyKey{ChainTitleId: payload.GetChainTitleId(), Key: payload.Key}

		return makeSetVerifyKeyReceipt(nil, key), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "setVerifyKey.getVerifyKeyData")
	}
	newKey := &zt.ZkVerifyKey{ChainTitleId: payload.ChainTitleId, Key: payload.Key}
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
		return errors.Wrapf(types.ErrInvalidParam, "calcInitRoot=%s, proof's oldRoot=%s", initRoot, payload.OldTreeRoot)
	}
	return nil
}

//
func (a *Action) commitProof(payload *zt.ZkCommitProof) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if payload.GetChainTitleId() <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "chainTitle is null")
	}
	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, new(big.Int).SetUint64(payload.ChainTitleId).String(), a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not validator")
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
		zklog.Error("commitProof.verifyInitRoot", "chainId", payload.ChainTitleId, "err", err)
		return nil, err
	}

	//1. 先验证proof是否ok
	//get verify key
	verifyKey, err := getVerifyKeyData(a.statedb, new(big.Int).SetUint64(payload.ChainTitleId).String())
	if err != nil {
		return nil, errors.Wrapf(err, "get verify key")
	}
	err = verifyProof(verifyKey.Key, payload)
	if err != nil {
		return nil, errors.Wrapf(err, "verify proof")
	}

	//更新数据库, public and proof, pubdata 不上链，存localdb
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
		ChainTitleId:      payload.ChainTitleId,
	}

	//2. 验证proof是否连续，不连续则暂时保存(考虑交易顺序被打散的场景)
	lastProof, err := getLastCommitProofData(a.statedb, new(big.Int).SetUint64(payload.ChainTitleId).String())
	if err != nil {
		return nil, errors.Wrap(err, "get last commit Proof")
	}
	if payload.ProofId < lastProof.ProofId+1 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "commitedId=%d <= lastProofId=%d", payload.ProofId, lastProof.ProofId)
	}
	//get未处理的证明的最大id
	maxRecordId, err := getMaxRecordProofIdData(a.statedb, new(big.Int).SetUint64(payload.ChainTitleId).String())
	if err != nil {
		return nil, errors.Wrapf(err, "getMaxRecordProofId for id=%d", payload.ProofId)
	}
	//不连续，先保存数据库,连续时候再验证
	if payload.ProofId > lastProof.ProofId+1 {
		return makeCommitProofRecordReceipt(newProof, uint64(maxRecordId.Data)), nil
	}
	lastOnChainProof, err := getLastOnChainProofData(a.statedb, new(big.Int).SetUint64(payload.ChainTitleId).String())
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
		recordProof, _ := getRecordProof(a.statedb, new(big.Int).SetUint64(payload.ChainTitleId).String(), i)
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
	if payload.GetChainTitleId() <= 0 || len(payload.Verifiers) == 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "chainTitle or verifier nil")
	}
	chainTitleId := new(big.Int).SetUint64(payload.ChainTitleId).String()
	oldKey, err := getVerifierData(a.statedb, chainTitleId)
	if isNotFound(errors.Cause(err)) {
		key := &zt.ZkVerifier{ChainTitleId: payload.ChainTitleId, Verifiers: payload.Verifiers}
		return makeSetVerifierReceipt(nil, key), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "setVerifyKey.getVerifyKeyData")
	}
	newKey := &zt.ZkVerifier{ChainTitleId: payload.ChainTitleId, Verifiers: payload.Verifiers}
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
			TokenHash:   l.TokenHash,
		}
		historyLeaf = append(historyLeaf, history)
	}
	return historyLeaf
}

func getHistoryAccountByRoot(localdb dbm.KV, chainTitleId uint64, targetRootHash string) (*zt.HistoryAccountProofInfo, error) {
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
		for _, operation := range operations {
			switch operation.Ty {
			case zt.TyDepositAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					fromLeaf = &zt.HistoryLeaf{
						AccountId:   operation.GetAccountId(),
						EthAddress:  operation.GetEthAddress(),
						Chain33Addr: operation.GetChain33Addr(),
						Tokens: []*zt.TokenBalance{
							{
								TokenId: operation.TokenId,
								Balance: operation.GetAmount(),
							},
						},
					}
				} else {
					var tokenBalance *zt.TokenBalance
					//找到token
					for _, token := range fromLeaf.Tokens {
						if token.TokenId == operation.TokenId {
							tokenBalance = token
						}
					}
					if tokenBalance == nil {
						tokenBalance = &zt.TokenBalance{
							TokenId: operation.TokenId,
							Balance: operation.Amount,
						}
						fromLeaf.Tokens = append(fromLeaf.Tokens, tokenBalance)
					} else {
						balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
						change, _ := new(big.Int).SetString(operation.Amount, 10)
						tokenBalance.Balance = new(big.Int).Add(balance, change).String()
					}
				}
				accountMap[operation.AccountId] = fromLeaf
				if operation.AccountId > maxAccountId {
					maxAccountId = operation.AccountId
				}
			case zt.TyWithdrawAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					fee, _ := new(big.Int).SetString(operation.FeeData.Fee, 10)
					//add fee
					change = new(big.Int).Add(change, fee)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				accountMap[operation.AccountId] = fromLeaf

			case zt.TyTransferAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				toLeaf, ok := accountMap[operation.ToAccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}

				var fromTokenBalance, toTokenBalance *zt.TokenBalance
				//找到fromToken
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						fromTokenBalance = token
					}
				}
				if fromTokenBalance == nil {
					return nil, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					fee, _ := new(big.Int).SetString(operation.FeeData.Fee, 10)
					//add fee
					change = new(big.Int).Add(change, fee)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("balance not enough")
					}
					fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}

				//找到toToken
				for _, token := range toLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						toTokenBalance = token
					}
				}
				if toTokenBalance == nil {
					toTokenBalance = &zt.TokenBalance{
						TokenId: operation.TokenId,
						Balance: operation.Amount,
					}
					toLeaf.Tokens = append(toLeaf.Tokens, toTokenBalance)
				} else {
					balance, _ := new(big.Int).SetString(toTokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					toTokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				accountMap[operation.AccountId] = fromLeaf
				accountMap[operation.ToAccountId] = toLeaf
			case zt.TyTransferToNewAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}

				var fromTokenBalance *zt.TokenBalance
				//找到fromToken
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						fromTokenBalance = token
					}
				}
				if fromTokenBalance == nil {
					return nil, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					fee, _ := new(big.Int).SetString(operation.FeeData.Fee, 10)
					//add fee
					change = new(big.Int).Add(change, fee)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("balance not enough")
					}
					fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}

				toLeaf := &zt.HistoryLeaf{
					AccountId:   operation.GetToAccountId(),
					EthAddress:  operation.GetEthAddress(),
					Chain33Addr: operation.GetChain33Addr(),
					Tokens: []*zt.TokenBalance{
						{
							TokenId: operation.TokenId,
							Balance: operation.GetAmount(),
						},
					},
				}

				accountMap[operation.AccountId] = fromLeaf
				accountMap[operation.ToAccountId] = toLeaf
				if operation.ToAccountId > maxAccountId {
					maxAccountId = operation.ToAccountId
				}
			case zt.TyForceExitAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New(fmt.Sprintf("account=%d not exist", operation.AccountId))
				}

				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New(fmt.Sprintf("token=%d not exist", operation.TokenId))
				} else {
					tokenBalance.Balance = "0"
				}
				accountMap[operation.AccountId] = fromLeaf
			case zt.TySetPubKeyAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}

				pubKey := &zt.ZkPubKey{
					X: operation.SetPubKey.PubKey.X,
					Y: operation.SetPubKey.PubKey.Y,
				}
				if fromLeaf.ProxyPubKeys == nil {
					fromLeaf.ProxyPubKeys = new(zt.AccountProxyPubKeys)
				}
				switch operation.SetPubKey.Ty {
				case 0:
					fromLeaf.PubKey = pubKey
				case zt.NormalProxyPubKey:
					fromLeaf.ProxyPubKeys.Normal = pubKey
				case zt.SystemProxyPubKey:
					fromLeaf.ProxyPubKeys.System = pubKey
				case zt.SuperProxyPubKey:
					fromLeaf.ProxyPubKeys.Super = pubKey
				default:
					return nil, errors.New(fmt.Sprintf("setPubKey ty=%d not support", operation.SetPubKey.Ty))
				}
				accountMap[operation.AccountId] = fromLeaf

			case zt.TyFullExitAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New(fmt.Sprintf("account=%d not exist", operation.AccountId))
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New(fmt.Sprintf("token=%d not exist", operation.TokenId))
				} else {
					tokenBalance.Balance = "0"
				}
				accountMap[operation.AccountId] = fromLeaf

			case zt.TySwapAction:
				//token: left asset, right asset 顺序
				//电路顺序为：sell-leftAsset, buy+leftAsset, sell-rightAsset-fee, buy+rightAsset-2ndFee
				//这里考虑leaf获取方便，顺序调整为 sell-leftAsset buy+rightAsset-2ndfee, sell-rightAsset-fee,buy+leftAsset
				takerLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//找到sell token
				for _, token := range takerLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("taker token not exist")
				}
				balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
				change, _ := new(big.Int).SetString(operation.Amount, 10)
				if change.Cmp(balance) > 0 {
					return nil, errors.New("balance not enough")
				}
				tokenBalance.Balance = new(big.Int).Sub(balance, change).String()

				//buy token
				tokenBalance = nil
				for _, token := range takerLeaf.Tokens {
					if token.TokenId == operation.SwapData.TokenId {
						tokenBalance = token
					}
				}
				//taker 2nd fee
				change, _ = new(big.Int).SetString(operation.SwapData.Amount, 10)
				fee, _ := new(big.Int).SetString(operation.SwapData.Fee, 10)
				if fee.Cmp(change) > 0 {
					return nil, errors.New("change not enough to fee to taker")
				}
				//sub fee
				change = new(big.Int).Sub(change, fee)
				if tokenBalance == nil {
					newToken := &zt.TokenBalance{
						TokenId: operation.SwapData.TokenId,
						Balance: change.String(),
					}
					takerLeaf.Tokens = append(takerLeaf.Tokens, newToken)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				accountMap[operation.AccountId] = takerLeaf

				//toAccount leaf
				makerLeaf, ok := accountMap[operation.ToAccountId]
				if !ok {
					return nil, errors.New(fmt.Sprintf("maker account=%d not exist", operation.ToAccountId))
				}

				//找到right asset
				tokenBalance = nil
				for _, token := range makerLeaf.Tokens {
					if token.TokenId == operation.SwapData.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New(fmt.Sprintf("maker sell token=%d not exist", operation.SwapData.TokenId))
				}
				balance, _ = new(big.Int).SetString(tokenBalance.GetBalance(), 10)
				change, _ = new(big.Int).SetString(operation.SwapData.Amount, 10)
				if change.Cmp(balance) > 0 {
					return nil, errors.New("maker token balance not enough")
				}
				newBalance := new(big.Int).Sub(balance, change)
				//1st fee
				fee, _ = new(big.Int).SetString(operation.FeeData.Fee, 10)
				if fee.Cmp(newBalance) > 0 {
					return nil, errors.New("change not enough to fee to taker")
				}
				tokenBalance.Balance = new(big.Int).Sub(newBalance, fee).String()

				//buy token
				tokenBalance = nil
				for _, token := range makerLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					newToken := &zt.TokenBalance{
						TokenId: operation.TokenId,
						Balance: operation.Amount,
					}
					makerLeaf.Tokens = append(makerLeaf.Tokens, newToken)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				accountMap[operation.ToAccountId] = makerLeaf

			case zt.TyContractToTreeAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					tokenBalance = &zt.TokenBalance{
						TokenId: operation.TokenId,
						Balance: operation.Amount,
					}
					fromLeaf.Tokens = append(fromLeaf.Tokens, tokenBalance)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				accountMap[operation.AccountId] = fromLeaf
			case zt.TyTreeToContractAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
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
				accountMap[operation.AccountId] = fromLeaf

			case zt.TyFeeAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New(fmt.Sprintf("fee accountId=%d not exist", operation.AccountId))
				}
				if operation.AccountId != zt.SystemFeeAccountId {
					return nil, errors.New(fmt.Sprintf("fee accountId=%d not systmeId=%d", operation.AccountId, zt.SystemFeeAccountId))
				}
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.FeeData.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					newToken := &zt.TokenBalance{
						TokenId: operation.FeeData.TokenId,
						Balance: operation.FeeData.Fee,
					}
					fromLeaf.Tokens = append(fromLeaf.Tokens, newToken)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.FeeData.Fee, 10)

					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				accountMap[operation.AccountId] = fromLeaf

			case zt.TyMintNFTAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//1. fee
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.FeeData.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("mint nft token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.FeeData.Fee, 10)
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
				newSysNFTTokenBalance, err := getNewNFTTokenBalance(operation.AccountId, serialId, operation.NFTData.ErcProtocol, mintAmount.Uint64(),
					operation.NFTData.Content.Part1, operation.NFTData.Content.Part2)
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
				toLeaf, ok := accountMap[operation.ToAccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				tokenBalance = nil
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.FeeData.TokenId {
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
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//1. fee
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.FeeData.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("withdraw nft fee token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.FeeData.Fee, 10)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("withdraw nft fee balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				//2. NFT token balance-amount
				tokenBalance = nil
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.NFTData.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("withdraw nft token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("withdraw nft  balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
			case zt.TyTransferNFTAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				var tokenBalance *zt.TokenBalance
				//1. fee
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.FeeData.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return nil, errors.New("transfer nft fee token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.FeeData.Fee, 10)
					if change.Cmp(balance) > 0 {
						return nil, errors.New("withdraw nft fee balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				//2. NFT token balance-amount
				tokenBalance = nil
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.NFTData.TokenId {
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

				toLeaf, ok := accountMap[operation.ToAccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}

				// NFT token balance+amount
				tokenBalance = nil
				for _, token := range toLeaf.Tokens {
					if token.TokenId == operation.NFTData.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					newToken := &zt.TokenBalance{
						TokenId: operation.NFTData.TokenId,
						Balance: operation.Amount,
					}
					toLeaf.Tokens = append(toLeaf.Tokens, newToken)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
			default:
				return nil, errors.New(fmt.Sprintf("not support op ty=%d", operation.Ty))
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

	return historyAccounts, nil

}

func GetHistoryAccountProof(historyAccountInfo *zt.HistoryAccountProofInfo, targetAccountId, targetTokenId uint64) (*zt.OperationMetaBranch, error) {
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
	var witness zt.OperationMetaBranch
	witness.AccountWitness = accountW
	witness.TokenWitness = tokenW

	return &witness, nil
}

//根据rootHash获取account在该root下的证明
func getAccountProofInHistory(localdb dbm.KV, req *zt.ZkReqExistenceProof) (*zt.OperationMetaBranch, error) {
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
	case zt.TyForceExitAction:
		return zt.ForceExitChunks
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
	case zt.TyForceExitAction:
		return getForceExitOperationByChunk(totalChunk)
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
func saveHistoryAccountTree(localdb dbm.KV, endProofId uint64, chainTitle string) ([]*types.KeyValue, error) {
	var localKvs []*types.KeyValue
	proofTable := NewCommitProofTable(localdb)
	historyTable := NewHistoryAccountTreeTable(localdb)
	//todo 多少ID归档一次实现可配置化
	historyId := (endProofId/10 - 1) * 10
	for i := historyId + 1; i <= endProofId; i++ {
		row, err := proofTable.GetData(getProofIdCommitProofKey(chainTitle, i))
		if err != nil {
			return localKvs, err
		}
		data := row.Data.(*zt.ZkCommitProof)
		operations := transferPubDatasToOption(data.PubDatas)
		for _, operation := range operations {
			fromLeaf, err := getAccountByProofIdAndHistoryId(historyTable, endProofId, historyId, operation.GetAccountId())
			if err != nil {
				return localKvs, errors.Wrapf(err, "getAccountByProofIdAndHistoryId")
			}
			switch operation.Ty {
			case zt.TyDepositAction:
				if fromLeaf == nil {
					fromLeaf = &zt.HistoryLeaf{
						AccountId:   operation.GetAccountId(),
						EthAddress:  operation.GetEthAddress(),
						Chain33Addr: operation.GetChain33Addr(),
						ProofId:     endProofId,
						Tokens: []*zt.TokenBalance{
							{
								TokenId: operation.TokenId,
								Balance: operation.GetAmount(),
							},
						},
					}
				} else {
					fromLeaf.ProofId = endProofId
					var tokenBalance *zt.TokenBalance
					//找到token
					for _, token := range fromLeaf.Tokens {
						if token.TokenId == operation.TokenId {
							tokenBalance = token
						}
					}
					if tokenBalance == nil {
						tokenBalance = &zt.TokenBalance{
							TokenId: operation.TokenId,
							Balance: operation.Amount,
						}
						fromLeaf.Tokens = append(fromLeaf.Tokens, tokenBalance)
					} else {
						balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
						change, _ := new(big.Int).SetString(operation.Amount, 10)
						tokenBalance.Balance = new(big.Int).Add(balance, change).String()
					}
				}
				err = historyTable.Replace(fromLeaf)
				if err != nil {
					return localKvs, err
				}
			case zt.TyWithdrawAction:
				if fromLeaf == nil {
					return localKvs, errors.New("account not exist")
				}
				fromLeaf.ProofId = endProofId
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return localKvs, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					if change.Cmp(balance) > 0 {
						return localKvs, errors.New("balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				err = historyTable.Replace(fromLeaf)
				if err != nil {
					return localKvs, err
				}
			case zt.TyTreeToContractAction:
				if fromLeaf == nil {
					return localKvs, errors.New("account not exist")
				}
				fromLeaf.ProofId = endProofId
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return localKvs, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					if change.Cmp(balance) > 0 {
						return localKvs, errors.New("balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}
				err = historyTable.Replace(fromLeaf)
				if err != nil {
					return localKvs, err
				}
			case zt.TyContractToTreeAction:
				if fromLeaf == nil {
					return localKvs, errors.New("account not exist")
				}
				fromLeaf.ProofId = endProofId
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					tokenBalance = &zt.TokenBalance{
						TokenId: operation.TokenId,
						Balance: operation.Amount,
					}
					fromLeaf.Tokens = append(fromLeaf.Tokens, tokenBalance)
				} else {
					balance, _ := new(big.Int).SetString(tokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					tokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}
				err = historyTable.Replace(fromLeaf)
				if err != nil {
					return localKvs, err
				}
			case zt.TyTransferAction:
				toLeaf, err := getAccountByProofIdAndHistoryId(historyTable, endProofId, historyId, operation.GetAccountId())
				if err != nil {
					return localKvs, errors.Wrapf(err, "getAccountByProofIdAndHistoryId")
				}
				if fromLeaf == nil || toLeaf == nil {
					return localKvs, errors.New("account not exist")
				}

				fromLeaf.ProofId = endProofId
				toLeaf.ProofId = endProofId

				var fromTokenBalance, toTokenBalance *zt.TokenBalance
				//找到fromToken
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						fromTokenBalance = token
					}
				}
				if fromTokenBalance == nil {
					return localKvs, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					if change.Cmp(balance) > 0 {
						return localKvs, errors.New("balance not enough")
					}
					fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}

				//找到toToken
				for _, token := range toLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						toTokenBalance = token
					}
				}
				if toTokenBalance == nil {
					toTokenBalance = &zt.TokenBalance{
						TokenId: operation.TokenId,
						Balance: operation.Amount,
					}
					toLeaf.Tokens = append(toLeaf.Tokens, toTokenBalance)
				} else {
					balance, _ := new(big.Int).SetString(toTokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					toTokenBalance.Balance = new(big.Int).Add(balance, change).String()
				}

				err = historyTable.Replace(fromLeaf)
				if err != nil {
					return localKvs, err
				}
				err = historyTable.Replace(toLeaf)
				if err != nil {
					return localKvs, err
				}
			case zt.TyTransferToNewAction:
				if fromLeaf == nil {
					return localKvs, errors.New("account not exist")
				}

				fromLeaf.ProofId = endProofId

				var fromTokenBalance *zt.TokenBalance
				//找到fromToken
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						fromTokenBalance = token
					}
				}
				if fromTokenBalance == nil {
					return localKvs, errors.New("token not exist")
				} else {
					balance, _ := new(big.Int).SetString(fromTokenBalance.GetBalance(), 10)
					change, _ := new(big.Int).SetString(operation.Amount, 10)
					if change.Cmp(balance) > 0 {
						return localKvs, errors.New("balance not enough")
					}
					fromTokenBalance.Balance = new(big.Int).Sub(balance, change).String()
				}

				toLeaf := &zt.HistoryLeaf{
					AccountId:   operation.GetToAccountId(),
					EthAddress:  operation.GetEthAddress(),
					Chain33Addr: operation.GetChain33Addr(),
					ProofId:     endProofId,
					Tokens: []*zt.TokenBalance{
						{
							TokenId: operation.TokenId,
							Balance: operation.GetAmount(),
						},
					},
				}

				err = historyTable.Replace(fromLeaf)
				if err != nil {
					return localKvs, err
				}
				err = historyTable.Replace(toLeaf)
				if err != nil {
					return localKvs, err
				}
			case zt.TySetPubKeyAction:
				if fromLeaf == nil {
					return localKvs, errors.New("account not exist")
				}
				fromLeaf.ProofId = endProofId
				fromLeaf.PubKey = &zt.ZkPubKey{
					X: operation.SetPubKey.PubKey.X,
					Y: operation.SetPubKey.PubKey.Y,
				}
				err = historyTable.Replace(fromLeaf)
				if err != nil {
					return localKvs, err
				}
			case zt.TyForceExitAction:
				if fromLeaf == nil {
					return localKvs, errors.New("account not exist")
				}
				fromLeaf.ProofId = endProofId
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return localKvs, errors.New("token not exist")
				} else {
					tokenBalance.Balance = "0"
				}
				err = historyTable.Replace(fromLeaf)
				if err != nil {
					return localKvs, err
				}
			case zt.TyFullExitAction:
				if fromLeaf == nil {
					return localKvs, errors.New("account not exist")
				}
				fromLeaf.ProofId = endProofId
				var tokenBalance *zt.TokenBalance
				//找到token
				for _, token := range fromLeaf.Tokens {
					if token.TokenId == operation.TokenId {
						tokenBalance = token
					}
				}
				if tokenBalance == nil {
					return localKvs, errors.New("token not exist")
				} else {
					tokenBalance.Balance = "0"
				}
				err = historyTable.Replace(fromLeaf)
				if err != nil {
					return localKvs, err
				}
			}
		}
	}
	localKvs, err := historyTable.Save()
	if err != nil {
		return localKvs, err
	}
	return localKvs, nil
}

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
	operation := &zt.ZkOperation{Ty: zt.TyDepositAction}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/8
	operation.Chain33Addr = zt.Byte2Str(chunk[start:end])
	return operation
}

func getWithDrawOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyWithdrawAction, FeeData: new(zt.ZkOpFeeData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}

func getSwapOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TySwapAction, FeeData: new(zt.ZkOpFeeData), SwapData: new(zt.ZkOpSwapData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.ToAccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	//1st token, left asset
	operation.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	//2nd token, right asset
	operation.SwapData.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacAmountExpBitWidth)/8
	//1st amount, left asset amount
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacAmountExpBitWidth)
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacAmountExpBitWidth)/8
	//2nd amount right asset amount
	operation.SwapData.Amount = zt.DecodePacVal(chunk[start:end], zt.PacAmountExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	//1st fee, maker's fee
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	//2nd fee, taker's fee
	operation.SwapData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}

func getContract2TreeOptionByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyContractToTreeAction}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacAmountExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacAmountExpBitWidth)
	return operation
}

func getTree2ContractOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyTreeToContractAction}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacAmountExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacAmountExpBitWidth)
	return operation
}

func getTransferOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyTransferAction, FeeData: new(zt.ZkOpFeeData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.ToAccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacAmountExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacAmountExpBitWidth)
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}

func getTransfer2NewOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyTransferToNewAction, FeeData: new(zt.ZkOpFeeData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.ToAccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacAmountManBitWidth+zt.PacAmountExpBitWidth)/8
	operation.Amount = zt.DecodePacVal(chunk[start:end], zt.PacAmountExpBitWidth)
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/8
	operation.Chain33Addr = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}

func getSetPubKeyOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TySetPubKeyAction, SetPubKey: new(zt.ZkSetPubKeyData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TxTypeBitWidth/8
	operation.SetPubKey.Ty = zt.Byte2Uint64(chunk[start:end])
	pubkey := &zt.ZkPubKey{}
	start = end
	end = start + zt.PubKeyBitWidth/8
	pubkey.X = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.PubKeyBitWidth/8
	pubkey.Y = zt.Byte2Str(chunk[start:end])
	operation.SetPubKey.PubKey = pubkey
	return operation
}

func getForceExitOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyForceExitAction, FeeData: new(zt.ZkOpFeeData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}

func getFullExitOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyFullExitAction, FeeData: new(zt.ZkOpFeeData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}

func getFeeOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyFeeAction, FeeData: new(zt.ZkOpFeeData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.FeeData.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}

func getMintNFTOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyMintNFTAction, FeeData: new(zt.ZkOpFeeData), NFTData: new(zt.ZkOpNFTData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.ToAccountId = zt.Byte2Uint64(chunk[start:end])
	//ERC 721/1155 protocol
	start = end
	end = start + zt.TxTypeBitWidth/8
	operation.NFTData.ErcProtocol = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.NFTData.Content.Part1 = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.NFTData.Content.Part2 = zt.Byte2Str(chunk[start:end])

	start = end
	end = start + zt.TokenBitWidth/8
	operation.FeeData.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}

func getWithdrawNFTOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyWithdrawNFTAction, FeeData: new(zt.ZkOpFeeData), NFTData: new(zt.ZkOpNFTData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	//fromId
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	//original creator id
	operation.NFTData.CreatorId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.NFTData.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.NFTData.CreatorSerialId = zt.Byte2Uint64(chunk[start:end])
	//ERC 721/1155 protocol
	start = end
	end = start + zt.TxTypeBitWidth/8
	operation.NFTData.ErcProtocol = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.NFTData.MintAmount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.AddrBitWidth/8
	operation.EthAddress = zt.Byte2Str(chunk[start:end])

	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.NFTData.Content.Part1 = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/(2*8)
	operation.NFTData.Content.Part2 = zt.Byte2Str(chunk[start:end])

	start = end
	end = start + zt.TokenBitWidth/8
	operation.FeeData.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}

func getTransferNFTOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyMintNFTAction, FeeData: new(zt.ZkOpFeeData), NFTData: new(zt.ZkOpNFTData)}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.AccountBitWidth/8
	operation.ToAccountId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.TokenBitWidth/8
	operation.NFTData.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + zt.NFTAmountBitWidth/8
	operation.NFTData.Amount = zt.Byte2Str(chunk[start:end])

	start = end
	end = start + zt.TokenBitWidth/8
	operation.FeeData.TokenId = zt.Byte2Uint64(chunk[start:end])
	start = end
	end = start + (zt.PacFeeManBitWidth+zt.PacFeeExpBitWidth)/8
	operation.FeeData.Fee = zt.DecodePacVal(chunk[start:end], zt.PacFeeExpBitWidth)
	return operation
}
