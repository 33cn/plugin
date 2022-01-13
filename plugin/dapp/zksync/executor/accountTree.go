package executor

import (
	"encoding/hex"
	"fmt"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/pkg/errors"
	"math/big"
)

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

func AddNewLeaf(db dbm.KV, tree *zt.AccountTree, ethAddress string, tokenId uint64, amount string, chain33Addr string) (*zt.Leaf, []*types.KeyValue, error) {
	var kvs []*types.KeyValue

	if tokenId == 0 || amount == "0" {
		return nil, kvs, errors.New("balance is zero")
	}

	//查找叶子是否存在
	oldLeaf, err := GetLeafByChain33AndEthAddress(db, ethAddress, chain33Addr)
	if err != nil {
		return nil, kvs, errors.Wrapf(err, "db.getLeafByEthAddress")
	}
	if oldLeaf != nil {
		return nil, kvs, errors.New("accountAlreadyExist")
	}

	currentTree := getNewTree()
	subtrees := make([]*zt.SubTree, 0)

	for _, subTree := range tree.GetSubTrees() {
		err := currentTree.PushSubTree(int(subTree.GetHeight()), subTree.GetRootHash())
		if err != nil {
			return nil, kvs, errors.Wrapf(err, "pushSubTree")
		}
	}

	tree.Index++
	tree.TotalIndex++

	leaf := &zt.Leaf{
		EthAddress:  ethAddress,
		AccountId:   tree.GetTotalIndex(),
		Chain33Addr: chain33Addr,
	}

	tokenBalance := &zt.TokenBalance{
		AccountId: leaf.AccountId,
		TokenId:   tokenId,
		Balance:   amount,
	}

	chainBalance.TokenBalanceMap[tokenId] = 0
	chainBalance.TokenBalances = append(chainBalance.TokenBalances, tokenBalance)
	setChainBalanceRootHash(chainBalance)
	leaf.ChainBalanceMap[chainType] = 0
	leaf.ChainBalances = append(leaf.ChainBalances, chainBalance)

	currentTree.Push(getLeafHash(leaf))
	for _, subtree := range currentTree.GetAllSubTrees() {
		subtrees = append(subtrees, &zt.SubTree{
			RootHash: subtree.GetSum(),
			Height:   int32(subtree.GetHeight()),
		})
	}
	tree.AddressMap[ethAddress] = tree.Index
	tree.LeaveMap[tree.Index] = leaf

	tree.SubTrees = subtrees

	//到达1024以后，归档
	if tree.Index == tree.MaxCurrentIndex {
		tree.Index = 0
		if tree.RootMap == nil {
			tree.RootMap = make(map[string]*zt.RootInfo)
		}
		if tree.RootIndexMap == nil {
			tree.RootIndexMap = make(map[int32][]byte)
		}
		rootInfo := &zt.RootInfo{
			Height:     10,
			StartIndex: tree.GetTotalIndex() - tree.MaxCurrentIndex + 1,
		}
		rootHash := currentTree.Root()
		tree.RootMap[hex.EncodeToString(rootHash)] = rootInfo
		tree.RootIndexMap[rootInfo.StartIndex/tree.MaxCurrentIndex] = rootHash
		accountTable := NewAccountTreeTable(db)
		leaves := make([]*zt.Leaf, 0)
		for i := int32(1); i <= tree.MaxCurrentIndex; i++ {
			v := tree.GetLeaveMap()[i]
			v.RootHash = rootHash
			leaves = append(leaves, v)
		}
		err = addLeaves(accountTable, leaves)
		if err != nil {
			return nil, errors.Wrapf(err, "db.addLeaves")
		}
		//落盘归档
		err = SaveAccountTreeTable(db, accountTable)
		if err != nil {
			return nil, errors.Wrapf(err, "db.SaveAccountTreeTable")
		}
		//清空当前的叶子和子树
		tree.AddressMap = make(map[string]int32)
		tree.LeaveMap = make(map[int32]*zt.Leaf)
		tree.SubTrees = make([]*zt.SubTree, 0)
	}

	err = setAccountTree(db, tree)
	if err != nil {
		return nil, errors.Wrapf(err, "db.setAccountTree")
	}
	return leaf, nil
}

func getNewTree() *merkletree.Tree {
	return merkletree.New(mimc.NewMiMC(mixTy.MimcHashSeed))
}

func getAccountTree(db dbm.KV) (*zt.AccountTree, error) {
	val, err := db.Get(GetAccountTreeKey())
	if err != nil {
		return nil, err
	}
	var tree zt.AccountTree
	err = types.Decode(val, &tree)
	if err != nil {
		return nil, err
	}
	return &tree, nil
}

func GetNowTotalIndex(db dbm.KV) (int32, error) {
	tree, err := getAccountTree(db)
	if err != nil {
		return 0, errors.Wrapf(err, "db.getAccountTree")
	}

	return tree.TotalIndex + 1, nil
}

func setAccountTree(db dbm.KV, tree *zt.AccountTree) error {
	err := db.Set(calcAccountTreeKey(), types.Encode(tree))
	if err != nil {
		return err
	}
	return nil
}

func addLeaves(table *table.Table, leaves []*zt.Leaf) error {
	for _, leaf := range leaves {
		err := table.Add(leaf)
		if err != nil {
			return err
		}

	}
	return nil
}

func updateLeaves(table *table.Table, leaves []*zt.Leaf) error {
	for _, leaf := range leaves {
		err := table.Update(GetAccountIdPrimaryKey(leaf.AccountId), leaf)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveAccountTreeTable(db dbm.KV, table *table.Table) error {
	kvs, err := table.Save()
	if err != nil {
		return err
	}
	for _, kv := range kvs {
		err = db.Set(kv.GetKey(), kv.GetValue())
		if err != nil {
			return err
		}
	}
	return nil
}

func GetLeafByAccountId(db dbm.KV, accountId uint64) (*zt.Leaf, error) {
	if accountId <= 0 {
		return nil, nil
	}

	var leaf zt.Leaf
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

	data := make([]*zt.Leaf, 0)
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return data, nil
		} else {
			return nil, err
		}
	}
	for _, row := range rows {
		data = append(data, row.Data.(*zt.Leaf))
	}
	return data, nil
}

func GetLeafByChain33Address(db dbm.KV, chain33Addr string) ([]*zt.Leaf, error) {
	accountTable := NewAccountTreeTable(db)
	rows, err := accountTable.ListIndex("chain33_address", []byte(fmt.Sprintf("%s", chain33Addr)), nil, 1, dbm.ListASC)

	data := make([]*zt.Leaf, 0)
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return data, nil
		} else {
			return nil, err
		}
	}
	for _, row := range rows {
		data = append(data, row.Data.(*zt.Leaf))
	}
	return data, nil
}

func GetLeafByChain33AndEthAddress(db dbm.KV, chain33Addr, ethAddress string) (*zt.Leaf, error) {
	if chain33Addr == "" || ethAddress == "" {
		return nil, nil
	}

	var leaf zt.Leaf
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

func GetLeavesByRootInfo(db dbm.KV, rootInfo *zt.RootInfo) ([]*zt.Leaf, error) {
	leaves := make([]*zt.Leaf, 0)
	for i := rootInfo.StartIndex; i < rootInfo.StartIndex+1024; i++ {
		leaf, err := GetLeafByAccountId(db, i)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, leaf)
	}
	return leaves, nil
}

func GetAllRoots(db dbm.KV, endIndex uint64) ([]*zt.RootInfo, error) {
	roots := make([]*zt.RootInfo, 0)
	for i := uint64(1); i <= endIndex; i++ {
		rootInfo, err := GetRootByIndex(db, i)
		if err != nil {
			return nil, err
		}
		roots = append(roots, rootInfo)
	}
	return roots, nil
}

func GetRootByIndex(db dbm.KV, index uint64) (*zt.RootInfo, error) {
	val, err := db.Get(GetRootIndexPrimaryKey(index))
	if err != nil {
		return nil, err
	}
	var rootInfo zt.RootInfo
	err = types.Decode(val, &rootInfo)
	if err != nil {
		return nil, err
	}
	return &rootInfo, nil
}

func GetTokenByAccountIdAndTokenId(db dbm.KV, accountId uint64, tokenId uint64) (*zt.TokenBalance, error) {
	val, err := db.Get(GetTokenPrimaryKey(accountId, tokenId))
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, err
		}
	}
	var token zt.TokenBalance
	err = types.Decode(val, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// UpdateLeaf 更新叶子结点：1、如果在当前树的叶子中，直接更新  2、如果在归档的树中，需要找到归档的root，重新生成root
func UpdateLeaf(db dbm.KV, accountId int32, chainType string, tokenId int32, balance int64) (*zt.Leaf, error) {
	tree, err := getAccountTree(db)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}

	//找到叶子
	leaf, err := GetLeafByAccountId(db, accountId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}
	if leaf == nil {
		return nil, errors.New("account not exist")
	}

	//从当前的树里面更新
	if accountId > tree.GetTotalIndex()-tree.GetIndex() {
		updateBalance(leaf, chainType, tokenId, balance)
		tree.GetLeaveMap()[accountId] = leaf
		currentTree := getNewTree()
		for i := int32(1); i <= tree.GetIndex(); i++ {
			currentTree.Push(getLeafHash(tree.GetLeaveMap()[i]))
		}

		subtrees := make([]*zt.SubTree, 0)
		for _, subtree := range currentTree.GetAllSubTrees() {
			subtrees = append(subtrees, &zt.SubTree{
				RootHash: subtree.GetSum(),
				Height:   int32(subtree.GetHeight()),
			})
		}

		tree.SubTrees = subtrees
		err = setAccountTree(db, tree)
		if err != nil {
			return nil, errors.Wrapf(err, "db.setAccountTree")
		}
		return leaf, nil
	}

	accountTable := NewAccountTreeTable(db)

	//找到对应的根
	if rootInfo, ok := tree.GetRootMap()[hex.EncodeToString(leaf.GetRootHash())]; ok {
		leaves, err := getLeavesByRoot(accountTable, leaf.GetRootHash())
		if err != nil {
			return nil, errors.Wrapf(err, "db.getLeavesByRoot")
		}
		oldRootHash := leaf.GetRootHash()
		currentTree := getNewTree()
		for _, leafVal := range leaves {
			if leafVal.GetAccountId() == leaf.GetAccountId() {
				updateBalance(leafVal, chainType, tokenId, balance)
			}
			currentTree.Push(getLeafHash(leafVal))
		}
		rootHash := currentTree.Root()
		for _, leafVal := range leaves {
			leafVal.RootHash = rootHash
		}
		err = updateLeaves(accountTable, leaves)
		if err != nil {
			return nil, errors.Wrapf(err, "db.updateLeaves")
		}
		//落盘保存
		err = SaveAccountTreeTable(db, accountTable)
		if err != nil {
			return nil, errors.Wrapf(err, "db.SaveAccountTreeTable")
		}
		//更新树
		delete(tree.RootMap, hex.EncodeToString(oldRootHash))
		tree.RootMap[hex.EncodeToString(rootHash)] = rootInfo
		tree.RootIndexMap[rootInfo.StartIndex/tree.MaxCurrentIndex] = rootHash
		err = setAccountTree(db, tree)
		if err != nil {
			return nil, errors.Wrapf(err, "db.setAccountTree")
		}
	} else {
		return nil, errors.Wrapf(err, "rootMap.notFindRoot")
	}
	return leaf, nil
}

func getLeafHash(leaf *zt.Leaf) []byte {
	hash := mimc.NewMiMC(mixTy.MimcHashSeed)
	hash.Write(new(big.Int).SetUint64(leaf.GetAccountId()).Bytes())
	hash.Write([]byte(leaf.GetEthAddress()))
	hash.Write([]byte(leaf.GetChain33Addr()))
	if leaf.GetPubKey() != nil {
		hash.Write([]byte(leaf.GetPubKey().X))
		hash.Write([]byte(leaf.GetPubKey().Y))
	}
	hash.Write(leaf.GetTokenHash())
	return hash.Sum(nil)
}

func getTokenRootHash(db dbm.KV, accountId uint64, tokenIds []uint64) ([]byte, error) {
	tree := getNewTree()
	for _, tokenId := range tokenIds {
		token, err := GetTokenByAccountIdAndTokenId(db, accountId, tokenId)
		if err != nil {
			return nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
		}
		tree.Push(getTokenBalanceHash(token))
	}
	return tree.Root(), nil
}

func getTokenBalanceHash(token *zt.TokenBalance) []byte {
	hash := mimc.NewMiMC(mixTy.MimcHashSeed)
	hash.Write(new(big.Int).SetUint64(token.GetTokenId()).Bytes())
	hash.Write([]byte(token.Balance))
	return hash.Sum(nil)
}

func updateBalance(leaf *zt.Leaf, chainType string, tokenId int32, balance int64) {
	if chain, ok := leaf.GetChainBalanceMap()[chainType]; ok {
		chainBalance := leaf.GetChainBalances()[chain]
		if index, ok := chainBalance.GetTokenBalanceMap()[tokenId]; ok {
			tokenBalance := chainBalance.GetTokenBalances()[index]
			tokenBalance.Balance += balance
		} else {
			tokenBalance := &zt.TokenBalance{TokenId: tokenId, Balance: balance}
			chainBalance.TokenBalanceMap[tokenId] = int32(len(chainBalance.TokenBalances))
			chainBalance.TokenBalances = append(chainBalance.TokenBalances, tokenBalance)
		}
		setChainBalanceRootHash(chainBalance)
	} else {
		tokenBalance := &zt.TokenBalance{TokenId: tokenId, Balance: balance}
		tokenBalanceMap := make(map[int32]int32)
		tokenBalances := make([]*zt.TokenBalance, 0)
		tokenBalances = append(tokenBalances, tokenBalance)
		tokenBalanceMap[tokenId] = 0
		chainBalance := &zt.ChainBalance{ChainType: chainType, TokenBalanceMap: tokenBalanceMap, TokenBalances: tokenBalances}
		setChainBalanceRootHash(chainBalance)
		leaf.ChainBalanceMap[chainType] = int32(len(leaf.ChainBalances))
		leaf.ChainBalances = append(leaf.ChainBalances, chainBalance)
	}
}

func CalLeafProof(db dbm.KV, accountId int32) (*zt.MerkleTreeProof, error) {
	tree, err := getAccountTree(db)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}

	leaf, err := GetLeafByAccountId(db, accountId)
	if err != nil {
		return nil, errors.Wrapf(err, "db.GetLeafByAccountId")
	}

	//leaf不存在的时候，计算子树
	if leaf == nil {
		currentTree := getNewTree()
		for i := int32(0); i < tree.TotalIndex/tree.MaxCurrentIndex; i++ {
			err = currentTree.PushSubTree(10, tree.RootIndexMap[i])
			if err != nil {
				return nil, errors.Wrapf(err, "db.PushSubTree")
			}
		}
		for i := int32(1); i <= tree.GetIndex(); i++ {
			currentTree.Push(getLeafHash(tree.GetLeaveMap()[i]))
		}
		subTrees := currentTree.GetAllSubTrees()
		proofSet := make([][]byte, len(subTrees)+1)
		helpers := make([]uint32, len(subTrees)+1)
		for i := len(subTrees); i > 0; i-- {
			proofSet[i] = subTrees[i-1].GetSum()
			helpers[i] = 1
		}
		proof := &zt.MerkleTreeProof{
			RootHash: currentTree.Root(),
			ProofSet: proofSet,
			Helpers:  helpers,
		}
		return proof, nil
	}

	currentTree := getNewTree()
	err = currentTree.SetIndex(uint64(leaf.GetAccountId() - 1))
	if err != nil {
		return nil, errors.Wrapf(err, "merkleTree.setIndex")
	}
	accountTable := NewAccountTreeTable(db)
	for i := int32(0); i < tree.TotalIndex/tree.MaxCurrentIndex; i++ {
		//如果需要验证的account在该归档节点中，需要捞出来所有root下的leaf的进行push
		if i == leaf.GetAccountId()/tree.MaxCurrentIndex {
			leaves, err := getLeavesByRoot(accountTable, leaf.GetRootHash())
			if err != nil {
				return nil, errors.Wrapf(err, "db.getLeavesByRoot")
			}
			for _, v := range leaves {
				fmt.Print("account Id ", v.GetAccountId())
				currentTree.Push(getLeafHash(v))
			}
		} else {
			err = currentTree.PushSubTree(10, tree.RootIndexMap[i])
			if err != nil {
				return nil, errors.Wrapf(err, "db.PushSubTree")
			}
		}
	}
	for i := int32(1); i <= tree.GetIndex(); i++ {
		currentTree.Push(getLeafHash(tree.GetLeaveMap()[i]))
	}

	rootHash, proofSet, proofIndex, numLeaves := currentTree.Prove()
	helpers := make([]uint32, 0)
	for _, v := range merkletree.GenerateProofHelper(proofSet, proofIndex, numLeaves) {
		helpers = append(helpers, uint32(v))
	}

	return &zt.MerkleTreeProof{RootHash: rootHash, ProofSet: proofSet, ProofIndex: proofIndex, NumLeaves: numLeaves, Helpers: helpers}, nil
}

func CalTokenProof(db dbm.KV, chainBalance *zt.ChainBalance, tokenId int32) (*zt.MerkleTreeProof, error) {
	//之前没有这条chain上的token，返回nil
	if chainBalance == nil {
		return nil, nil
	} else {
		//如果存在token
		if index, ok := chainBalance.GetTokenBalanceMap()[tokenId]; ok {
			tree := getNewTree()
			err := tree.SetIndex(uint64(index))
			if err != nil {
				return nil, errors.Wrapf(err, "tree.SetIndex")
			}
			for _, balance := range chainBalance.GetTokenBalances() {
				tree.Push(getTokenBalanceHash(balance))
			}
			rootHash, proofSet, proofIndex, numLeaves := tree.Prove()
			helpers := make([]uint32, 0)
			for _, v := range merkletree.GenerateProofHelper(proofSet, proofIndex, numLeaves) {
				helpers = append(helpers, uint32(v))
			}
			return &zt.MerkleTreeProof{RootHash: rootHash, ProofSet: proofSet, ProofIndex: proofIndex, NumLeaves: numLeaves, Helpers: helpers}, nil
		} else {
			//如果不存在token，仅返回子树
			tree := getNewTree()
			for _, balance := range chainBalance.GetTokenBalances() {
				tree.Push(getTokenBalanceHash(balance))
			}
			subTrees := tree.GetAllSubTrees()
			proofSet := make([][]byte, len(subTrees)+1)
			helpers := make([]uint32, len(subTrees)+1)
			for i := len(subTrees); i > 0; i-- {
				proofSet[i] = subTrees[i].GetSum()
				helpers[i] = 1
			}
			proof := &zt.MerkleTreeProof{
				RootHash: tree.Root(),
				ProofSet: proofSet,
				Helpers:  helpers,
			}
			return proof, nil
		}
	}
}

func UpdatePubKey(db dbm.KV, leaf *zt.Leaf, pubKey []byte) (*zt.Leaf, error) {
	tree, err := getAccountTree(db)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}
	accountId := leaf.GetAccountId()
	leaf.PubKey = pubKey
	if accountId > tree.GetTotalIndex()-tree.GetIndex() {
		tree.GetLeaveMap()[accountId] = leaf
		currentTree := getNewTree()
		for i := int32(1); i <= tree.GetIndex(); i++ {
			currentTree.Push(getLeafHash(tree.GetLeaveMap()[i]))
		}

		subtrees := make([]*zt.SubTree, 0)
		for _, subtree := range currentTree.GetAllSubTrees() {
			subtrees = append(subtrees, &zt.SubTree{
				RootHash: subtree.GetSum(),
				Height:   int32(subtree.GetHeight()),
			})
		}

		tree.SubTrees = subtrees
		err = setAccountTree(db, tree)
		if err != nil {
			return nil, errors.Wrapf(err, "db.setAccountTree")
		}
		return leaf, nil
	}

	accountTable := NewAccountTreeTable(db)

	//找到对应的根
	if rootInfo, ok := tree.GetRootMap()[hex.EncodeToString(leaf.GetRootHash())]; ok {
		leaves, err := getLeavesByRoot(accountTable, leaf.GetRootHash())
		if err != nil {
			return nil, errors.Wrapf(err, "db.getLeavesByRoot")
		}
		oldRootHash := leaf.GetRootHash()
		currentTree := getNewTree()
		for _, leafVal := range leaves {
			if leafVal.GetAccountId() == leaf.GetAccountId() {
				leafVal.PubKey = pubKey
			}
			currentTree.Push(getLeafHash(leafVal))
		}
		rootHash := currentTree.Root()
		for _, leafVal := range leaves {
			leafVal.RootHash = rootHash
		}
		err = updateLeaves(accountTable, leaves)
		if err != nil {
			return nil, errors.Wrapf(err, "db.updateLeaves")
		}
		//落盘保存
		err = SaveAccountTreeTable(db, accountTable)
		if err != nil {
			return nil, errors.Wrapf(err, "db.SaveAccountTreeTable")
		}
		//更新树
		delete(tree.RootMap, hex.EncodeToString(oldRootHash))
		tree.RootMap[hex.EncodeToString(rootHash)] = rootInfo
		tree.RootIndexMap[rootInfo.StartIndex/tree.MaxCurrentIndex] = rootHash
		err = setAccountTree(db, tree)
		if err != nil {
			return nil, errors.Wrapf(err, "db.setAccountTree")
		}
	} else {
		return nil, errors.Wrapf(err, "rootMap.notFindRoot")
	}
	return leaf, nil
}
