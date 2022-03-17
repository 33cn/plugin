package executor

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"math/big"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/pkg/errors"
)

// TreeUpdateInfo 更新信息，用于查询
type TreeUpdateInfo struct {
	updateMap map[string][]byte
}

func NewAccountTree(db dbm.KV) *zt.AccountTree {
	tree := &zt.AccountTree{
		Index:           0,
		TotalIndex:      0,
		MaxCurrentIndex: 1024,
		SubTrees:        make([]*zt.SubTree, 0),
	}
	err := db.Set(GetAccountTreeKey(), types.Encode(tree))
	if err != nil {
		panic(err)
	}
	return tree
}

func AddNewLeaf(statedb dbm.KV, localdb dbm.KV, info *TreeUpdateInfo, ethAddress string, tokenId uint64, amount string, chain33Addr string) ([]*types.KeyValue, []*types.KeyValue, error) {
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue

	if amount == "0" {
		return kvs, localKvs, errors.New("balance is zero")
	}
	tree, err := getAccountTree(statedb, info)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.getAccountTree")
	}

	currentTree := getNewTree()
	subtrees := make([]*zt.SubTree, 0)

	for _, subTree := range tree.GetSubTrees() {
		err := currentTree.PushSubTree(int(subTree.GetHeight()), subTree.GetRootHash())
		if err != nil {
			return kvs, localKvs, errors.Wrapf(err, "pushSubTree")
		}
	}

	tree.Index++
	tree.TotalIndex++

	leaf := &zt.Leaf{
		EthAddress:  ethAddress,
		AccountId:   tree.GetTotalIndex(),
		Chain33Addr: chain33Addr,
		TokenIds:    make([]uint64, 0),
	}

	leaf.TokenIds = append(leaf.TokenIds, tokenId)
	tokenBalance := &zt.TokenBalance{
		TokenId: tokenId,
		Balance: amount,
	}

	kv := &types.KeyValue{
		Key:   GetTokenPrimaryKey(leaf.AccountId, tokenId),
		Value: types.Encode(tokenBalance),
	}
	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()

	leaf.TokenHash, err = getTokenRootHash(statedb, leaf.AccountId, leaf.TokenIds, info)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.getTokenRootHash")
	}

	kv = &types.KeyValue{
		Key:   GetAccountIdPrimaryKey(leaf.AccountId),
		Value: types.Encode(leaf),
	}

	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()

	kv = &types.KeyValue{
		Key:   GetChain33EthPrimaryKey(leaf.Chain33Addr, leaf.EthAddress),
		Value: types.Encode(leaf),
	}

	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()

	currentTree.Push(getLeafHash(leaf))
	for _, subtree := range currentTree.GetAllSubTrees() {
		subtrees = append(subtrees, &zt.SubTree{
			RootHash: subtree.GetSum(),
			Height:   int32(subtree.GetHeight()),
		})
	}

	tree.SubTrees = subtrees

	//到达1024以后，清空
	if tree.Index == tree.MaxCurrentIndex {
		root := &zt.RootInfo{
			Height:     10,
			StartIndex: tree.GetTotalIndex() - tree.GetIndex() + 1,
			RootHash:   zt.Byte2Str(currentTree.Root()),
		}
		tree.Index = 0
		tree.SubTrees = make([]*zt.SubTree, 0)

		kv = &types.KeyValue{
			Key:   GetRootIndexPrimaryKey(root.GetStartIndex()),
			Value: types.Encode(root),
		}
		kvs = append(kvs, kv)
		info.updateMap[string(kv.GetKey())] = kv.GetValue()
	}

	accountTable := NewAccountTreeTable(localdb)
	err = accountTable.Add(leaf)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "accountTable.Add")
	}
	//localdb存入叶子，用于查询
	localKvs, err = accountTable.Save()
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.SaveAccountTreeTable")
	}

	kv = &types.KeyValue{
		Key:   GetAccountTreeKey(),
		Value: types.Encode(tree),
	}

	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()
	return kvs, localKvs, nil
}

func getNewTree() *merkletree.Tree {
	return merkletree.New(mimc.NewMiMC(zt.ZkMimcHashSeed))
}

func getAccountTree(db dbm.KV, info *TreeUpdateInfo) (*zt.AccountTree, error) {
	var tree zt.AccountTree
	if info != nil {
		if val, ok := info.updateMap[string(GetAccountTreeKey())]; ok {
			err := types.Decode(val, &tree)
			if err != nil {
				return nil, err
			}
			return &tree, nil
		}
	}
	val, err := db.Get(GetAccountTreeKey())
	if err != nil {
		return nil, err
	}
	err = types.Decode(val, &tree)
	if err != nil {
		return nil, err
	}
	return &tree, nil
}

func GetLeafByAccountId(db dbm.KV, accountId uint64, info *TreeUpdateInfo) (*zt.Leaf, error) {
	if accountId <= 0 {
		return nil, nil
	}

	var leaf zt.Leaf
	if val, ok := info.updateMap[string(GetAccountIdPrimaryKey(accountId))]; ok {
		err := types.Decode(val, &leaf)
		if err != nil {
			return nil, err
		}
		return &leaf, nil
	}
	val, err := db.Get(GetAccountIdPrimaryKey(accountId))
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, err
		}
	}

	err = types.Decode(val, &leaf)
	if err != nil {
		return nil, err
	}
	return &leaf, nil
}

func GetLeafByEthAddress(db dbm.KV, ethAddress string) ([]*zt.Leaf, error) {
	accountTable := NewAccountTreeTable(db)
	rows, err := accountTable.ListIndex("eth_address", []byte(fmt.Sprintf("%s", ethAddress)), nil, 1, dbm.ListASC)

	datas := make([]*zt.Leaf, 0)
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return datas, nil
		} else {
			return nil, err
		}
	}
	for _, row := range rows {
		data := row.Data.(*zt.Leaf)
		data.EthAddress = zt.DecimalAddr2Hex(data.GetEthAddress())
		data.Chain33Addr = zt.DecimalAddr2Hex(data.GetChain33Addr())
		datas = append(datas, data)
	}
	return datas, nil
}

func GetLeafByChain33Address(db dbm.KV, chain33Addr string) ([]*zt.Leaf, error) {
	accountTable := NewAccountTreeTable(db)
	rows, err := accountTable.ListIndex("chain33_address", []byte(fmt.Sprintf("%s", chain33Addr)), nil, 1, dbm.ListASC)

	datas := make([]*zt.Leaf, 0)
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return datas, nil
		} else {
			return nil, err
		}
	}
	for _, row := range rows {
		data := row.Data.(*zt.Leaf)
		data.EthAddress = zt.DecimalAddr2Hex(data.GetEthAddress())
		data.Chain33Addr = zt.DecimalAddr2Hex(data.GetChain33Addr())
		datas = append(datas, data)
	}
	return datas, nil
}

func GetLeafByChain33AndEthAddress(db dbm.KV, chain33Addr, ethAddress string, info *TreeUpdateInfo) (*zt.Leaf, error) {
	if chain33Addr == "" || ethAddress == "" {
		return nil, types.ErrInvalidParam
	}

	var leaf zt.Leaf
	if val, ok := info.updateMap[string(GetChain33EthPrimaryKey(chain33Addr, ethAddress))]; ok {
		err := types.Decode(val, &leaf)
		if err != nil {
			return nil, err
		}
		return &leaf, nil
	}

	val, err := db.Get(GetChain33EthPrimaryKey(chain33Addr, ethAddress))
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, err
		}
	}

	err = types.Decode(val, &leaf)
	if err != nil {
		return nil, err
	}
	return &leaf, nil
}

func GetLeavesByStartAndEndIndex(db dbm.KV, startIndex uint64, endIndex uint64, info *TreeUpdateInfo) ([]*zt.Leaf, error) {
	leaves := make([]*zt.Leaf, 0)
	for i := startIndex; i <= endIndex; i++ {
		leaf, err := GetLeafByAccountId(db, i, info)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	return leaves, nil
}

func GetAllRoots(db dbm.KV, endIndex uint64, info *TreeUpdateInfo) ([]*zt.RootInfo, error) {
	roots := make([]*zt.RootInfo, 0)
	for i := uint64(1); i <= endIndex; i++ {
		rootInfo, err := GetRootByStartIndex(db, (i-1)*1024+1, info)
		if err != nil {
			return nil, err
		}
		roots = append(roots, rootInfo)
	}
	return roots, nil
}

func GetRootByStartIndex(db dbm.KV, index uint64, info *TreeUpdateInfo) (*zt.RootInfo, error) {
	var rootInfo zt.RootInfo
	if val, ok := info.updateMap[string(GetRootIndexPrimaryKey(index))]; ok {
		err := types.Decode(val, &rootInfo)
		if err != nil {
			return nil, err
		}
		return &rootInfo, nil
	}

	val, err := db.Get(GetRootIndexPrimaryKey(index))
	if err != nil {
		return nil, err
	}

	err = types.Decode(val, &rootInfo)
	if err != nil {
		return nil, err
	}
	return &rootInfo, nil
}

func GetTokenByAccountIdAndTokenId(db dbm.KV, accountId uint64, tokenId uint64, info *TreeUpdateInfo) (*zt.TokenBalance, error) {

	var token zt.TokenBalance
	if val, ok := info.updateMap[string(GetTokenPrimaryKey(accountId, tokenId))]; ok {
		err := types.Decode(val, &token)
		if err != nil {
			return nil, err
		}
		return &token, nil
	}

	val, err := db.Get(GetTokenPrimaryKey(accountId, tokenId))
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, err
		}
	}

	err = types.Decode(val, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func GetTokenByAccountIdAndTokenIdInDB(db dbm.KV, accountId uint64, tokenId uint64) (*zt.TokenBalance, error) {

	var token zt.TokenBalance

	val, err := db.Get(GetTokenPrimaryKey(accountId, tokenId))
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, err
		}
	}

	err = types.Decode(val, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateLeaf 更新叶子结点：1、如果在当前树的叶子中，直接更新  2、如果在归档的树中，需要找到归档的root，重新生成root
func UpdateLeaf(statedb dbm.KV, localdb dbm.KV, info *TreeUpdateInfo, accountId uint64, tokenId uint64, amount string, option int32) ([]*types.KeyValue, []*types.KeyValue, error) {
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue
	leaf, err := GetLeafByAccountId(statedb, accountId, info)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	tree, err := getAccountTree(statedb, info)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.getAccountTree")
	}
	token, err := GetTokenByAccountIdAndTokenId(statedb, accountId, tokenId, info)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.getAccountTree")
	}
	if token == nil {
		if option == zt.Sub {
			return kvs, localKvs, errors.New("token not exist")
		} else {
			token = &zt.TokenBalance{
				TokenId: tokenId,
				Balance: amount,
			}
			leaf.TokenIds = append(leaf.TokenIds, tokenId)
		}
	} else {
		balance, _ := new(big.Int).SetString(token.GetBalance(), 10)
		change, _ := new(big.Int).SetString(amount, 10)
		if option == zt.Add {
			token.Balance = new(big.Int).Add(balance, change).String()
		} else if option == zt.Sub {
			token.Balance = new(big.Int).Sub(balance, change).String()
		} else {
			return kvs, localKvs, types.ErrNotSupport
		}
	}

	kv := &types.KeyValue{
		Key:   GetTokenPrimaryKey(accountId, tokenId),
		Value: types.Encode(token),
	}

	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()

	leaf.TokenHash, err = getTokenRootHash(statedb, accountId, leaf.TokenIds, info)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.getTokenRootHash")
	}

	kv = &types.KeyValue{
		Key:   GetAccountIdPrimaryKey(leaf.AccountId),
		Value: types.Encode(leaf),
	}

	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()

	kv = &types.KeyValue{
		Key:   GetChain33EthPrimaryKey(leaf.Chain33Addr, leaf.EthAddress),
		Value: types.Encode(leaf),
	}

	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()

	//如果还没归档
	if accountId > tree.GetTotalIndex()-tree.GetIndex() {
		currentTree := getNewTree()
		leaves, err := GetLeavesByStartAndEndIndex(statedb, tree.GetTotalIndex()-tree.GetIndex()+1, tree.GetTotalIndex(), info)
		if err != nil {
			return kvs, localKvs, errors.Wrapf(err, "db.GetLeavesByStartAndEndIndex")
		}
		for _, leafVal := range leaves {
			currentTree.Push(getLeafHash(leafVal))
		}

		subtrees := make([]*zt.SubTree, 0)
		for _, subtree := range currentTree.GetAllSubTrees() {
			subtrees = append(subtrees, &zt.SubTree{
				RootHash: subtree.GetSum(),
				Height:   int32(subtree.GetHeight()),
			})
		}

		tree.SubTrees = subtrees
		kv = &types.KeyValue{
			Key:   GetAccountTreeKey(),
			Value: types.Encode(tree),
		}
		kvs = append(kvs, kv)
		info.updateMap[string(kv.GetKey())] = kv.GetValue()
	} else {
		//找到对应的根
		rootInfo, err := GetRootByStartIndex(statedb, (accountId-1)/1024*1024+1, info)
		if err != nil {
			return kvs, localKvs, errors.Wrapf(err, "db.GetRootByStartIndex")
		}
		leaves, err := GetLeavesByStartAndEndIndex(statedb, rootInfo.StartIndex, rootInfo.StartIndex+1023, info)
		if err != nil {
			return kvs, localKvs, errors.Wrapf(err, "db.GetLeavesByStartAndEndIndex")
		}
		currentTree := getNewTree()
		for _, leafVal := range leaves {
			currentTree.Push(getLeafHash(leafVal))
		}

		//生成新root
		rootInfo.RootHash = zt.Byte2Str(currentTree.Root())
		kv = &types.KeyValue{
			Key:   GetRootIndexPrimaryKey(rootInfo.StartIndex),
			Value: types.Encode(rootInfo),
		}
		kvs = append(kvs, kv)
		info.updateMap[string(kv.GetKey())] = kv.GetValue()
	}

	accountTable := NewAccountTreeTable(localdb)
	err = accountTable.Update(GetLocalChain33EthPrimaryKey(leaf.GetChain33Addr(), leaf.GetEthAddress()), leaf)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "accountTable.Update")
	}
	//localdb更新叶子，用于查询
	localKvs, err = accountTable.Save()
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.SaveAccountTreeTable")
	}

	kv = &types.KeyValue{
		Key:   GetAccountTreeKey(),
		Value: types.Encode(tree),
	}

	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()
	return kvs, localKvs, nil
}

func getLeafHash(leaf *zt.Leaf) []byte {
	hash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	accountIdBytes := new(fr.Element).SetUint64(leaf.GetAccountId()).Bytes()
	hash.Write(accountIdBytes[:])
	hash.Write(zt.Str2Byte(leaf.GetEthAddress()))
	hash.Write(zt.Str2Byte(leaf.GetChain33Addr()))
	if leaf.GetPubKey() != nil {
		hash.Write(zt.Str2Byte(leaf.GetPubKey().GetX()))
		hash.Write(zt.Str2Byte(leaf.GetPubKey().GetY()))
	} else {
		hash.Write(zt.Str2Byte("0")) //X
		hash.Write(zt.Str2Byte("0")) //Y
	}
	token := zt.Str2Byte(leaf.GetTokenHash())
	hash.Write(token)
	return hash.Sum(nil)
}

func getTokenRootHash(db dbm.KV, accountId uint64, tokenIds []uint64, info *TreeUpdateInfo) (string, error) {
	tree := getNewTree()
	for _, tokenId := range tokenIds {
		token, err := GetTokenByAccountIdAndTokenId(db, accountId, tokenId, info)
		if err != nil {
			return "", err
		}
		tree.Push(getTokenBalanceHash(token))
	}
	return zt.Byte2Str(tree.Root()), nil
}

func getTokenBalanceHash(token *zt.TokenBalance) []byte {
	hash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	tokenIdBytes := new(fr.Element).SetUint64(token.GetTokenId()).Bytes()
	hash.Write(tokenIdBytes[:])
	hash.Write(zt.Str2Byte(token.Balance))
	return hash.Sum(nil)
}

func getHistoryLeafHash(leaf *zt.HistoryLeaf) []byte {

	hash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	accountIdBytes := new(fr.Element).SetUint64(leaf.GetAccountId()).Bytes()
	hash.Write(accountIdBytes[:])
	hash.Write(zt.Str2Byte(leaf.GetEthAddress()))
	hash.Write(zt.Str2Byte(leaf.GetChain33Addr()))
	if leaf.GetPubKey() != nil {
		hash.Write(zt.Str2Byte(leaf.GetPubKey().GetX()))
		hash.Write(zt.Str2Byte(leaf.GetPubKey().GetY()))
	} else {
		hash.Write(zt.Str2Byte("0")) //X
		hash.Write(zt.Str2Byte("0")) //Y
	}

	tokenTree := getNewTree()
	for _, token := range leaf.Tokens {
		tokenTree.Push(getTokenBalanceHash(token))
	}
	hash.Write(tokenTree.Root())
	return hash.Sum(nil)
}

func CalLeafProof(statedb dbm.KV, leaf *zt.Leaf, info *TreeUpdateInfo) (*zt.MerkleTreeProof, error) {
	tree, err := getAccountTree(statedb, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}

	//leaf不存在的时候，计算子树
	if leaf == nil {
		currentTree := getNewTree()
		roots, err := GetAllRoots(statedb, tree.TotalIndex/1024, info)
		if err != nil {
			return nil, errors.Wrapf(err, "db.GetAllRoots")
		}
		for _, root := range roots {
			rootHash := zt.Str2Byte(root.GetRootHash())
			err = currentTree.PushSubTree(int(root.Height), rootHash)
			if err != nil {
				return nil, errors.Wrapf(err, "db.PushSubTree")
			}
		}
		for _, subTree := range tree.SubTrees {
			err = currentTree.PushSubTree(int(subTree.Height), subTree.RootHash)
			if err != nil {
				return nil, errors.Wrapf(err, "db.PushSubTree")
			}
		}
		subTrees := currentTree.GetAllSubTrees()
		proofSet := make([]string, len(subTrees)+1)
		helpers := make([]string, len(subTrees))
		proofSet[0] = "0"
		for i := 1; i <= len(subTrees); i++ {
			proofSet[i] = zt.Byte2Str(subTrees[len(subTrees)-i].GetSum())
			helpers[i-1] = big.NewInt(0).String()
		}
		proof := &zt.MerkleTreeProof{
			RootHash: zt.Byte2Str(currentTree.Root()),
			ProofSet: proofSet,
			Helpers:  helpers,
		}
		//如果还没有产生第一个叶子，RootHash需要特殊设置
		if tree.TotalIndex == 0 {
			proof.RootHash = "0"
		}
		return proof, nil
	}

	currentTree := getNewTree()
	err = currentTree.SetIndex(leaf.GetAccountId() - 1)
	if err != nil {
		return nil, errors.Wrapf(err, "merkleTree.setIndex")
	}
	roots, err := GetAllRoots(statedb, tree.GetTotalIndex()/1024, info)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetAllRoots")
	}
	if leaf.AccountId > tree.GetTotalIndex()-tree.GetIndex() {
		leaves, err := GetLeavesByStartAndEndIndex(statedb, tree.GetTotalIndex()-tree.GetIndex()+1, tree.GetTotalIndex(), info)
		if err != nil {
			return nil, errors.Wrapf(err, "db.GetLeavesByStartAndEndIndex")
		}
		for _, root := range roots {
			rootHash := zt.Str2Byte(root.GetRootHash())
			err = currentTree.PushSubTree(int(root.Height), rootHash)
			if err != nil {
				return nil, errors.Wrapf(err, "db.PushSubTree")
			}
		}
		for _, v := range leaves {
			currentTree.Push(getLeafHash(v))
		}
	} else {
		startIndex := (leaf.AccountId-1)/1024*1024 + 1
		leaves, err := GetLeavesByStartAndEndIndex(statedb, startIndex, startIndex+1023, info)
		if err != nil {
			return nil, errors.Wrapf(err, "db.GetLeavesByStartAndEndIndex")
		}
		for _, root := range roots {
			//如果需要验证的account在该root节点中，需要对所有root下的leaf的进行push
			if startIndex == root.StartIndex {
				for _, v := range leaves {
					currentTree.Push(getLeafHash(v))
				}
			} else {
				rootHash := zt.Str2Byte(root.GetRootHash())
				err = currentTree.PushSubTree(int(root.Height), rootHash)
				if err != nil {
					return nil, errors.Wrapf(err, "db.PushSubTree")
				}
			}
		}
		for _, subTree := range tree.SubTrees {
			err = currentTree.PushSubTree(int(subTree.Height), subTree.RootHash)
			if err != nil {
				return nil, errors.Wrapf(err, "db.PushSubTree")
			}
		}
	}

	rootHash, proofSet, proofIndex, numLeaves := currentTree.Prove()
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

func CalTokenProof(statedb dbm.KV, leaf *zt.Leaf, token *zt.TokenBalance, info *TreeUpdateInfo) (*zt.MerkleTreeProof, error) {
	if leaf == nil {
		return nil, nil
	}
	tokens := make([]*zt.TokenBalance, 0)
	index := 0
	for i, v := range leaf.TokenIds {
		if token != nil && token.TokenId == v {
			index = i
		}
		tokenVal, err := GetTokenByAccountIdAndTokenId(statedb, leaf.AccountId, v, info)
		if err != nil {
			return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
		}
		tokens = append(tokens, tokenVal)
	}
	//如果存在token
	if token != nil {
		tree := getNewTree()
		err := tree.SetIndex(uint64(index))
		if err != nil {
			return nil, errors.Wrapf(err, "tree.SetIndex")
		}
		for _, balance := range tokens {
			tree.Push(getTokenBalanceHash(balance))
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
	} else {
		//如果不存在token，仅返回子树
		tree := getNewTree()
		for _, balance := range tokens {
			tree.Push(getTokenBalanceHash(balance))
		}
		subTrees := tree.GetAllSubTrees()
		proofSet := make([]string, len(subTrees)+1)
		helpers := make([]string, len(subTrees))
		proofSet[0] = "0"
		for i := 1; i <= len(subTrees); i++ {
			proofSet[i] = zt.Byte2Str(subTrees[len(subTrees)-i].GetSum())
			helpers[i-1] = big.NewInt(0).String()
		}
		proof := &zt.MerkleTreeProof{
			RootHash: zt.Byte2Str(tree.Root()),
			ProofSet: proofSet,
			Helpers:  helpers,
		}
		return proof, nil
	}

}

func UpdatePubKey(statedb dbm.KV, localdb dbm.KV, info *TreeUpdateInfo, pubKey *zt.ZkPubKey, accountId uint64) ([]*types.KeyValue, []*types.KeyValue, error) {
	var kvs []*types.KeyValue
	var localKvs []*types.KeyValue
	tree, err := getAccountTree(statedb, info)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.getAccountTree")
	}
	leaf, err := GetLeafByAccountId(statedb, accountId, info)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	if leaf == nil {
		return kvs, localKvs, errors.New("account not exist")
	}
	leaf.PubKey = pubKey

	kv := &types.KeyValue{
		Key:   GetAccountIdPrimaryKey(leaf.AccountId),
		Value: types.Encode(leaf),
	}

	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()

	kv = &types.KeyValue{
		Key:   GetChain33EthPrimaryKey(leaf.Chain33Addr, leaf.EthAddress),
		Value: types.Encode(leaf),
	}

	kvs = append(kvs, kv)
	info.updateMap[string(kv.GetKey())] = kv.GetValue()

	if accountId > tree.GetTotalIndex()-tree.GetIndex() {
		leaves, err := GetLeavesByStartAndEndIndex(statedb, tree.GetTotalIndex()-tree.GetIndex()+1, tree.GetTotalIndex(), info)
		if err != nil {
			return kvs, localKvs, errors.Wrapf(err, "db.GetLeavesByStartAndEndIndex")
		}
		currentTree := getNewTree()
		for _, v := range leaves {
			currentTree.Push(getLeafHash(v))
		}

		subtrees := make([]*zt.SubTree, 0)
		for _, subtree := range currentTree.GetAllSubTrees() {
			subtrees = append(subtrees, &zt.SubTree{
				RootHash: subtree.GetSum(),
				Height:   int32(subtree.GetHeight()),
			})
		}

		tree.SubTrees = subtrees
		kv = &types.KeyValue{
			Key:   GetAccountTreeKey(),
			Value: types.Encode(tree),
		}
		kvs = append(kvs, kv)
		info.updateMap[string(kv.GetKey())] = kv.GetValue()
	} else {
		//找到对应的根
		rootInfo, err := GetRootByStartIndex(statedb, (accountId-1)/1024*1024+1, info)
		if err != nil {
			return kvs, localKvs, errors.Wrapf(err, "db.GetRootByStartIndex")
		}
		leaves, err := GetLeavesByStartAndEndIndex(statedb, rootInfo.StartIndex, rootInfo.StartIndex+1023, info)
		if err != nil {
			return kvs, localKvs, errors.Wrapf(err, "db.GetLeavesByStartAndEndIndex")
		}
		currentTree := getNewTree()
		for _, leafVal := range leaves {
			currentTree.Push(getLeafHash(leafVal))
		}
		rootInfo.RootHash = zt.Byte2Str(currentTree.Root())
		kv = &types.KeyValue{
			Key:   GetRootIndexPrimaryKey(rootInfo.StartIndex),
			Value: types.Encode(rootInfo),
		}
		kvs = append(kvs, kv)
		info.updateMap[string(kv.GetKey())] = kv.GetValue()
	}
	accountTable := NewAccountTreeTable(localdb)
	err = accountTable.Update(GetLocalChain33EthPrimaryKey(leaf.GetChain33Addr(), leaf.GetEthAddress()), leaf)
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "accountTable.Update")
	}

	//localdb更新叶子，用于查询
	localKvs, err = accountTable.Save()
	if err != nil {
		return kvs, localKvs, errors.Wrapf(err, "db.SaveAccountTreeTable")
	}
	return kvs, localKvs, nil
}
