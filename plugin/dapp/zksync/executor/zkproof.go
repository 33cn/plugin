package executor

import (
	"bytes"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/zksnark"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/frontend"
	"github.com/pkg/errors"
)

func makeSetVerifyKeyReceipt(oldKey, newKey *zt.ZkVerifyKey) *types.Receipt {
	key := getVerifyKey()
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
	key := getLastProofIdKey()
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
		onChainIdKey := getLastOnChainProofIdKey()
		r.KV = append(r.KV, &types.KeyValue{Key: onChainIdKey,
			Value: types.Encode(&zt.LastOnChainProof{ProofId: newState.ProofId, OnChainProofId: newState.OnChainProofId})})
	}
	return r
}

func makeCommitProofRecordReceipt(proof *zt.CommitProofState, maxRecordId uint64) *types.Receipt {
	//对record里面已有的proofId，直接更新
	key := getProofIdKey(proof.ProofId)
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

	//如果此proofId 比maxRecordId更大，更新maxRecordId，方便遍历
	if proof.ProofId > maxRecordId {
		r.KV = append(r.KV, &types.KeyValue{Key: getMaxRecordProofIdKey(),
			Value: types.Encode(&types.Int64{Data: int64(proof.ProofId)})})
	}

	return r
}

func makeProofId2QueueIdReceipt(proofId uint64, firstQueueId, lastQueueId int64) *types.Receipt {
	key := getProofId2QueueIdKey(proofId)
	data := &zt.ProofId2QueueIdData{
		ProofId:      proofId,
		FirstQueueId: firstQueueId,
		LastQueueId:  lastQueueId,
	}
	log := &zt.ReceiptProofId2QueueIDData{
		Data: data,
	}
	r := &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(data)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: zt.TySetProofId2QueueIdLog, Log: types.Encode(log)},
		},
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

func isVerifier(statedb dbm.KV, addr string) bool {
	verifier, err := getVerifierData(statedb)
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
		key := &zt.ZkVerifyKey{
			Key: payload.Key,
		}
		return makeSetVerifyKeyReceipt(nil, key), nil
	}

	if err != nil {
		return nil, errors.Wrap(err, "setVerifyKey.getVerifyKeyData")
	}
	newKey := &zt.ZkVerifyKey{
		Key: payload.Key,
	}
	return makeSetVerifyKeyReceipt(oldKey, newKey), nil
}

func getLastCommitProofData(db dbm.KV) (*zt.CommitProofState, error) {
	key := getLastProofIdKey()
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

func getLastOnChainProofData(db dbm.KV) (*zt.LastOnChainProof, error) {
	key := getLastOnChainProofIdKey()
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

func getMaxRecordProofIdData(db dbm.KV) (*types.Int64, error) {
	key := getMaxRecordProofIdKey()
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
	//默认系统是在平行链下，获取缺省配置的eth和layer2Addr. 如果是在主链上只能配置一个eth/layer2Addr，不允许修改
	initRoot := getInitTreeRoot(a.api.GetConfig(), "", "")
	if initRoot != payload.OldTreeRoot {
		return errors.Wrapf(types.ErrInvalidParam, "calcInitRoot=%s, proof's oldRoot=%s", initRoot, payload.OldTreeRoot)
	}
	return nil
}

func (a *Action) commitProof(payload *zt.ZkCommitProof) (*types.Receipt, error) {
	cfg := a.api.GetConfig()

	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, a.fromaddr) {
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
		return nil, err
	}

	//1. 先验证proof是否ok
	//get verify key
	verifyKey, err := getVerifyKeyData(a.statedb)
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
		//pubdatas上链，主要是考虑提前提交的record proof的验证
		PubDatas: payload.PubDatas,
	}

	//2. 验证proof是否连续，不连续则暂时保存(考虑交易顺序被打散的场景)
	lastProof, err := getLastCommitProofData(a.statedb)
	if err != nil {
		return nil, errors.Wrap(err, "get last commit Proof")
	}
	//小于当前proofId的proof直接reject
	if payload.ProofId < lastProof.ProofId+1 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "commitedId=%d less or equal  lastProofId=%d", payload.ProofId, lastProof.ProofId)
	}
	//get未处理的证明的最大id
	maxRecordId, err := getMaxRecordProofIdData(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "getMaxRecordProofId for id=%d", payload.ProofId)
	}
	//不连续，先保存数据库,连续时候再验证
	if payload.ProofId > lastProof.ProofId+1 {
		return makeCommitProofRecordReceipt(newProof, uint64(maxRecordId.Data)), nil
	}
	//onChainProof是提交到L1的proof，可能proofId不连续(有些proof都是L2 op)，但onChainId需要连续
	lastOnChainProof, err := getLastOnChainProofData(a.statedb)
	if err != nil {
		return nil, errors.Wrap(err, "getLastOnChainProof")
	}
	lastOnChainProofId, err := checkNewProof(lastProof, newProof, lastOnChainProof.OnChainProofId)
	if err != nil {
		return nil, errors.Wrapf(err, "checkNewProof id=%d", newProof.ProofId)
	}
	receipt := makeCommitProofReceipt(lastProof, newProof)

	//检查证明的pubdata是否和queue一致
	firstQueueId, err := GetL2FirstQueueId(a.statedb)
	if err != nil {
		return nil, errors.Wrapf(err, "GetL2FirstQueueId")
	}
	oldFirstQueId := firstQueueId
	firstQueueId, err = checkNewProofPubData(a.statedb, oldFirstQueId, payload.PubDatas)
	if err != nil {
		return nil, errors.Wrapf(err, "checkNewProofPubData")
	}
	//记录proofId对应的开始结束queueId，可以根据proof定位其结束的queueId，以此对余下的queueId的某些操作比如withdraw做回滚(逃生舱场景)
	mergeReceipt(receipt, makeProofId2QueueIdReceipt(payload.ProofId, oldFirstQueId+1, firstQueueId))

	//循环检查可能未处理的recordProof
	lastProof = newProof
	for i := lastProof.ProofId + 1; i < uint64(maxRecordId.Data); i++ {
		recordProof, _ := getRecordProof(a.statedb, i)
		if recordProof == nil {
			break
		}
		lastOnChainProofId, err = checkNewProof(lastProof, recordProof, lastOnChainProofId)
		if err != nil {
			zklog.Error("commitProof.checkRecordProof", "lastProofId", lastProof.ProofId, "recordProofId", recordProof.ProofId, "err", err)
			//record检查出错，不作为本交易的错误，待下次更新错误的proofId
			break
		}
		//检查证明的pubdata是否和queue一致
		newFirstQueueId, err := checkNewProofPubData(a.statedb, firstQueueId, recordProof.PubDatas)
		if err != nil {
			zklog.Error("checkRecordProof.checkNewProofPubData", "firstQueueId", firstQueueId, "recordProofId", recordProof.ProofId, "err", err)
			break
		}
		//整个证明验证成功才更新
		mergeReceipt(receipt, makeCommitProofReceipt(lastProof, recordProof))
		lastProof = recordProof
		mergeReceipt(receipt, makeProofId2QueueIdReceipt(recordProof.ProofId, firstQueueId+1, newFirstQueueId))
		firstQueueId = newFirstQueueId
	}
	//更新firstQueueId 到成功的proof最后一个queueId
	mergeReceipt(receipt, makeSetL2FirstQueueIdReceipt(oldFirstQueId, firstQueueId))
	return receipt, nil
}

func getRecordProof(db dbm.KV, id uint64) (*zt.CommitProofState, error) {
	key := getProofIdKey(id)
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
		lastOnChainProofId = newProof.OnChainProofId
	}

	return lastOnChainProofId, nil
}

//检查来自proof的pubdata和queue里的operation一致
//链上每个op都会把数据压入queue中，包括fee，链下提交的证明的pubdatas要和压入的queue op顺序和数值严格一致，好处是抗回滚
//first queue op从id=0开始,一旦被proof验证过后firstOpId会移到最后一个验证了的id
//  1,2,3|,4,5,6,7|-----
//       3=first queue op, 7=last queue op
func checkNewProofPubData(db dbm.KV, lastQueueId int64, pubData []string) (int64, error) {
	ops := transferPubDataToOps(pubData)
	for _, o := range ops {
		lastQueueId += 1
		queueOp, err := GetL2QueueIdOp(db, lastQueueId)
		if err != nil {
			return 0, errors.Wrapf(err, "GetL2QueueIdOp id=%d", lastQueueId)
		}
		err = checkOpSame(queueOp, o)
		if err != nil {
			return 0, errors.Wrapf(err, "checkOpSame queueId=%d", lastQueueId)
		}
	}
	return lastQueueId, nil

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
	calcPubDataHash := calcPubDataCommitHash(mimcHash, proof.BlockStart, proof.BlockEnd, proof.OldTreeRoot, proof.NewTreeRoot, proof.PubDatas)
	if commitPubDataHash.String() != calcPubDataHash {
		return errors.Wrapf(types.ErrInvalidParam, "pubData hash not match, PI=%s,calc=%s", commitPubDataHash.String(), calcPubDataHash)
	}

	//计算onChain pubData hash 需要和commit的一致
	commitOnChainPubDataHash := proofCircuit.OnChainPubDataCommitment.GetWitnessValue(ecc.BN254)
	calcOnChainPubDataHash := calcOnChainPubDataCommitHash(mimcHash, proof.NewTreeRoot, proof.OnChainPubDatas)
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

//合约管理员或管理员设置在链上的管理员才可设置
func (a *Action) setVerifier(payload *zt.ZkVerifier) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
	}
	if len(payload.Verifiers) == 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "verifier nil")
	}

	oldKey, err := getVerifierData(a.statedb)
	if isNotFound(errors.Cause(err)) {
		key := &zt.ZkVerifier{Verifiers: payload.Verifiers}
		return makeSetVerifierReceipt(nil, key), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "setVerifyKey.getVerifyKeyData")
	}
	newKey := &zt.ZkVerifier{Verifiers: payload.Verifiers}
	return makeSetVerifierReceipt(oldKey, newKey), nil
}

func getVerifierData(db dbm.KV) (*zt.ZkVerifier, error) {
	key := getVerifier()
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
	key := getVerifier()
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
