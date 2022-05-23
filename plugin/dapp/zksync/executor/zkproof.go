package executor

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"

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
			{Ty: zt.TySetVerifyKeyLog, Log: types.Encode(log)},
		},
	}

}

func makeCommitProofReceipt(old, new *zt.CommitProofState) *types.Receipt {
	key := getLastProofKey()
	onChainIdKey := getLastOnChainProofIdKey()
	log := &zt.ReceiptCommitProof{
		Prev:    old,
		Current: new,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(new)},
			{Key: onChainIdKey, Value: types.Encode(&zt.LastOnChainProof{ProofId: new.ProofId, OnChainProofId: new.OnChainProofId})},
		},
		Logs: []*types.ReceiptLog{
			{Ty: zt.TyCommitProofLog, Log: types.Encode(log)},
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
		key := &zt.ZkVerifyKey{Key: payload.Key}

		return makeSetVerifyKeyReceipt(nil, key), nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "setVerifyKey.getVerifyKeyData")
	}
	newKey := &zt.ZkVerifyKey{Key: payload.Key}
	return makeSetVerifyKeyReceipt(oldKey, newKey), nil
}

func getLastCommitProofData(db dbm.KV, cfg *types.Chain33Config) (*zt.CommitProofState, error) {
	key := getLastProofKey()
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
				NewTreeRoot: getInitTreeRoot(cfg, "", ""),
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

//
func (a *Action) commitProof(payload *zt.ZkCommitProof) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) && !isVerifier(a.statedb, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not validator")
	}

	lastProof, err := getLastCommitProofData(a.statedb, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "get last commit Proof")
	}

	//proofId需要连续,高度需要衔接
	if lastProof.ProofId+1 != payload.ProofId || (lastProof.ProofId > 0 && lastProof.BlockEnd != payload.BlockStart) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "last proof id+1=%d, newId=%d, lastBlockEnd=%d,newStart=%d",
			lastProof.ProofId+1, payload.ProofId, lastProof.BlockEnd, payload.BlockStart)
	}

	//if lastProof.ProofId+1 != payload.ProofId {
	//	return nil, errors.Wrapf(types.ErrInvalidParam, "last proof id+1=%d, newId=%d, lastBlockEnd=%d,newStart=%d",
	//		lastProof.ProofId+1, payload.ProofId,lastProof.BlockEnd,payload.BlockStart)
	//}

	//tree root 需要衔接
	if lastProof.NewTreeRoot != payload.OldTreeRoot {
		return nil, errors.Wrapf(types.ErrInvalidParam, "last proof treeRoot=%s, commit oldTreeRoot=%s",
			lastProof.NewTreeRoot, payload.OldTreeRoot)
	}

	//如果包含OnChainPubData, ProofSubId需要连续
	if len(payload.OnChainPubDatas) > 0 {
		lastOnChainProof, err := getLastOnChainProofData(a.statedb)
		if err != nil {
			return nil, errors.Wrap(err, "getProofSubId")
		}
		if lastOnChainProof.OnChainProofId+1 != payload.OnChainProofId {
			return nil, errors.Wrapf(types.ErrInvalidParam, "lastSubId not match, lastSubId+1=%d, commit=%d", lastOnChainProof.GetOnChainProofId()+1, payload.OnChainProofId)
		}
	} else if payload.GetOnChainProofId() != 0 { //非onChain proof subId需要填0
		return nil, errors.Wrapf(types.ErrInvalidParam, "not onChain proof subId should be 0")
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
		BlockStart:        payload.BlockStart,
		BlockEnd:          payload.BlockEnd,
		IndexStart:        payload.IndexStart,
		IndexEnd:          payload.IndexEnd,
		OpIndex:           payload.OpIndex,
		ProofId:           payload.ProofId,
		OldTreeRoot:       payload.OldTreeRoot,
		NewTreeRoot:       payload.NewTreeRoot,
		PublicInput:       payload.PublicInput,
		Proof:             payload.Proof,
		OnChainProofId:    payload.OnChainProofId,
		CommitBlockHeight: a.height,
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

	//计算pubData hash 需要和commit的一致
	commitPubDataHash := proofCircuit.PubDataCommitment.GetWitnessValue(ecc.BN254)
	calcPubDataHash := calcPubDataCommitHash(proof.BlockStart, proof.BlockEnd, proof.OldTreeRoot, proof.NewTreeRoot, proof.PubDatas)
	if commitPubDataHash.String() != calcPubDataHash {
		return errors.Wrapf(types.ErrInvalidParam, "pubData hash not match, PI=%s,calc=%s", commitPubDataHash.String(), calcPubDataHash)
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

	for _, r := range pubDatas {
		t = f.SetString(r).Bytes()
		mimcHash.Write(t[:])
	}
	ret := mimcHash.Sum(nil)

	return f.SetBytes(ret).String()
}

//合约管理员或管理员设置在链上的管理员才可设置
func (a *Action) setVerifier(payload *zt.ZkVerifier) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "from addr is not manager")
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

func makeSetVerifierReceipt(old, new *zt.ZkVerifier) *types.Receipt {
	key := getVerifier()
	log := &zt.ReceiptSetVerifier{
		Prev:    old,
		Current: new,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(new)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: zt.TySetVerifierLog, Log: types.Encode(log)},
		},
	}

}

//根据rootHash获取account在该root下的证明
func getAccountProofInHistory(localdb dbm.KV, accountId uint64, rootHash string) (*zt.MerkleTreeProof, error) {
	proofTable := NewCommitProofTable(localdb)
	accountMap := make(map[uint64]*zt.HistoryLeaf)
	maxAccountId := uint64(0)
	rows, err := proofTable.ListIndex("root", getRootCommitProofKey(rootHash), nil, 1, zt.ListASC)
	if err != nil {
		return nil, errors.Wrapf(err, "proofTable.ListIndex")
	}
	proof := rows[0].Data.(*zt.ZkCommitProof)
	if proof == nil {
		return nil, errors.New("proof not exist")
	}

	for i := uint64(1); i <= proof.ProofId; i++ {
		row, err := proofTable.GetData(getProofIdCommitProofKey(i))
		if err != nil {
			return nil, err
		}
		data := row.Data.(*zt.ZkCommitProof)
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
					if change.Cmp(balance) > 0 {
						return nil, errors.New("balance not enough")
					}
					tokenBalance.Balance = new(big.Int).Sub(balance, change).String()
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
			case zt.TySetPubKeyAction:
				fromLeaf, ok := accountMap[operation.AccountId]
				if !ok {
					return nil, errors.New("account not exist")
				}
				fromLeaf.PubKey = &zt.ZkPubKey{
					X: operation.PubKey.X,
					Y: operation.PubKey.Y,
				}
				accountMap[operation.AccountId] = fromLeaf
			case zt.TyForceExitAction:
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
					tokenBalance.Balance = "0"
				}
				accountMap[operation.AccountId] = fromLeaf
			case zt.TyFullExitAction:
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
					tokenBalance.Balance = "0"
				}
				accountMap[operation.AccountId] = fromLeaf
			}
		}
	}

	tree := getNewTree()
	err = tree.SetIndex(accountId - 1)
	if err != nil {
		return nil, errors.Wrapf(err, "tree.SetIndex")
	}
	for i := uint64(1); i <= maxAccountId; i++ {
		tree.Push(getHistoryLeafHash(accountMap[i]))
	}

	merkleRoot, proofSet, proofIndex, numLeaves := tree.Prove()
	helpers := make([]string, 0)
	proofStringSet := make([]string, 0)
	for _, v := range merkletree.GenerateProofHelper(proofSet, proofIndex, numLeaves) {
		helpers = append(helpers, big.NewInt(int64(v)).String())
	}
	for _, v := range proofSet {
		proofStringSet = append(proofStringSet, zt.Byte2Str(v))
	}
	merkleProof := &zt.MerkleTreeProof{
		RootHash: zt.Byte2Str(merkleRoot),
		ProofSet: proofStringSet,
		Helpers:  helpers,
	}
	return merkleProof, nil
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
			fmt.Print(operation, "\n")
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
	case zt.TyDepositAction:
		return zt.DepositChunks
	case zt.TyWithdrawAction:
		return zt.WithdrawChunks
	case zt.TyContractToTreeAction:
		return zt.Contract2TreeChunks
	case zt.TyTreeToContractAction:
		return zt.Tree2ContractChunks
	case zt.TyTransferAction:
		return zt.TransferChunks
	case zt.TyTransferToNewAction:
		return zt.Transfer2NewChunks
	case zt.TyForceExitAction:
		return zt.ForceExitChunks
	case zt.TyFullExitAction:
		return zt.FullExitChunks
	case zt.TySetPubKeyAction:
		return zt.SetPubKeyChunks
	case zt.TyNoopAction:
		return zt.NoopChunks
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
	case zt.TyTreeToContractAction:
		return getTree2ContractOperationByChunk(totalChunk)
	case zt.TyContractToTreeAction:
		return getContract2TreeOptionByChunk(totalChunk)
	case zt.TyTransferAction:
		return getTransferOperationByChunk(totalChunk)
	case zt.TyTransferToNewAction:
		return getTransfer2NewOperationByChunk(totalChunk)
	case zt.TySetPubKeyAction:
		return getSetPubKeyOperationByChunk(totalChunk)
	case zt.TyForceExitAction:
		return getForceExitOperationByChunk(totalChunk)
	case zt.TyFullExitAction:
		return getFullExitOperationByChunk(totalChunk)
	default:
		panic("operationTy not support")
	}
}

//根据proofId重建merkleTree
func saveHistoryAccountTree(localdb dbm.KV, endProofId uint64) ([]*types.KeyValue, error) {
	var localKvs []*types.KeyValue
	proofTable := NewCommitProofTable(localdb)
	historyTable := NewHistoryAccountTreeTable(localdb)
	//todo 多少ID归档一次实现可配置化
	historyId := (endProofId/10 - 1) * 10
	for i := historyId + 1; i <= endProofId; i++ {
		row, err := proofTable.GetData(getProofIdCommitProofKey(i))
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
					X: operation.PubKey.X,
					Y: operation.PubKey.Y,
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
	operation := &zt.ZkOperation{Ty: zt.TyWithdrawAction}
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
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
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
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	return operation
}

func getTransferOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyTransferAction}
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
	end = start + zt.AmountBitWidth/8
	operation.Amount = zt.Byte2Str(chunk[start:end])
	return operation
}

func getTransfer2NewOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyTransferToNewAction}
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

func getSetPubKeyOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TySetPubKeyAction}
	start := zt.TxTypeBitWidth / 8
	end := start + zt.AccountBitWidth/8
	operation.AccountId = zt.Byte2Uint64(chunk[start:end])
	pubkey := &zt.ZkPubKey{}
	start = end
	end = start + zt.HashBitWidth/8
	pubkey.X = zt.Byte2Str(chunk[start:end])
	start = end
	end = start + zt.HashBitWidth/8
	pubkey.Y = zt.Byte2Str(chunk[start:end])
	operation.PubKey = pubkey
	return operation
}

func getForceExitOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyForceExitAction}
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
	return operation
}

func getFullExitOperationByChunk(chunk []byte) *zt.ZkOperation {
	operation := &zt.ZkOperation{Ty: zt.TyFullExitAction}
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
	return operation
}
