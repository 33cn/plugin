package executor

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	ety "github.com/33cn/plugin/plugin/dapp/exchange/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
)

func NewAccountTree(db dbm.KV) *ety.AccountTree {
	tree := &ety.AccountTree{
		Index:           0,
		TotalIndex:      0,
		MaxCurrentIndex: 1024,
		RootMap:         make(map[string]*ety.RootInfo),
		RootIndexMap:    make(map[int32][]byte),
		AddressMap:      make(map[string]int32),
		LeaveMap:        make(map[int32]*ety.Leaf),
		SubTrees:        make([]*ety.SubTree, 0),
	}
	err := db.Set(calcAccountTreeKey(), types.Encode(tree))
	if err != nil {
		panic(err)
	}
	return tree
}

func AddNewLeaf(db dbm.KV, ethAddress string, chainType string, tokenId int32, balance int64) (*ety.Leaf, error) {
	tree, err := getAccountTree(db)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}

	//从localdb和当前tree里面查找有没有存在
	oldLeaf, err := getLeafByEthAddressOnDB(db, ethAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getLeafByEthAddress")
	}
	if oldLeaf != nil {
		return nil, errors.New("accountAlreadyExist")
	}

	if tree.GetAddressMap() == nil {
		tree.AddressMap = make(map[string]int32)
	}

	if tree.GetLeaveMap() == nil {
		tree.LeaveMap = make(map[int32]*ety.Leaf)
	}

	if _, ok := tree.GetAddressMap()[ethAddress]; ok {
		return nil, errors.New("accountAlreadyExist")
	}

	currentTree := getNewTree()
	subtrees := make([]*ety.SubTree, 0)

	for _, subTree := range tree.GetSubTrees() {
		err := currentTree.PushSubTree(int(subTree.GetHeight()), subTree.GetRootHash())
		if err != nil {
			return nil, errors.Wrapf(err, "pushSubTree")
		}
	}

	tree.Index++
	tree.TotalIndex++

	leaf := &ety.Leaf{
		EthAddress: ethAddress,
		RootHash:   []byte("current"),
		AccountId:  tree.GetTotalIndex(),
	}

	//如果有初始balance，设置
	if chainType != "" && tokenId != 0 {
		leaf.ChainBalanceMap = make(map[string]int32)
		leaf.ChainBalances = make([]*ety.ChainBalance, 0)
		tokenBalance := &ety.TokenBalance{
			TokenId: tokenId,
			Balance: balance,
		}
		chainBalance := &ety.ChainBalance{
			ChainType:       chainType,
			TokenBalanceMap: make(map[int32]int32),
			TokenBalances: make([]*ety.TokenBalance, 0),
		}
		chainBalance.TokenBalanceMap[tokenId] = 0
		chainBalance.TokenBalances = append(chainBalance.TokenBalances, tokenBalance)
		setChainBalanceRootHash(chainBalance)
		leaf.ChainBalanceMap[chainType] = 0
		leaf.ChainBalances = append(leaf.ChainBalances, chainBalance)
	}

	currentTree.Push(getLeafHash(leaf))
	for _, subtree := range currentTree.GetAllSubTrees() {
		subtrees = append(subtrees, &ety.SubTree{
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
			tree.RootMap = make(map[string]*ety.RootInfo)
		}
		if tree.RootIndexMap == nil {
			tree.RootIndexMap = make(map[int32][]byte)
		}
		rootInfo := &ety.RootInfo{
			Height:     10,
			StartIndex: tree.GetTotalIndex() - tree.MaxCurrentIndex + 1,
		}
		rootHash := currentTree.Root()
		tree.RootMap[hex.EncodeToString(rootHash)] = rootInfo
		tree.RootIndexMap[rootInfo.StartIndex/tree.MaxCurrentIndex] = rootHash
		accountTable := NewAccountTreeTable(db)
		leaves := make([]*ety.Leaf, 0)
		for i:= int32(1); i<= tree.MaxCurrentIndex;i++ {
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
		tree.LeaveMap = make(map[int32]*ety.Leaf)
		tree.SubTrees = make([]*ety.SubTree, 0)
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

func calcAccountTreeKey() []byte {
	return []byte("accountTree:")
}

func getAccountTree(db dbm.KV) (*ety.AccountTree, error) {
	val, err := db.Get(calcAccountTreeKey())
	if err != nil {
		return nil, err
	}
	var tree ety.AccountTree
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

func setAccountTree(db dbm.KV, tree *ety.AccountTree) error {
	err := db.Set(calcAccountTreeKey(), types.Encode(tree))
	if err != nil {
		return err
	}
	return nil
}

func addLeaves(table *table.Table, leaves []*ety.Leaf) error {
	for _, leaf := range leaves {
		err := table.Add(leaf)
		if err != nil {
			return err
		}

	}
	return nil
}

func updateLeaves(table *table.Table, leaves []*ety.Leaf) error {
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

func GetLeafByAccountId(db dbm.KV, accountId int32) (*ety.Leaf, error) {
	tree, err := getAccountTree(db)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}
	if accountId > tree.GetTotalIndex() {
		return nil, errors.New("account not exist")
	}
	var leaf *ety.Leaf
	//如果accountId在当前tree中
	if accountId > tree.GetTotalIndex() - tree.GetIndex() {
		if tree.GetLeaveMap() != nil {
			leaf = tree.GetLeaveMap()[accountId]
		}
	} else {
		accountTable := NewAccountTreeTable(db)
		row, err := accountTable.GetData(GetAccountIdPrimaryKey(accountId))
		if err != nil {
			if err.Error() == types.ErrNotFound.Error() {
				return nil, nil
			} else {
				return nil, err
			}
		}
		leaf = row.Data.(*ety.Leaf)
	}

	return leaf, nil
}

func getLeafByEthAddressOnDB(db dbm.KV, ethAddress string) (*ety.Leaf, error) {
	accountTable := NewAccountTreeTable(db)
	rows, err := accountTable.ListIndex("eth_address", []byte(fmt.Sprintf("%s", ethAddress)), nil, 1, dbm.ListASC)

	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, err
		}
	}

	data := rows[0].Data.(*ety.Leaf)
	return data, nil
}

func GetLeafByEthAddress(db dbm.KV, ethAddress string) (*ety.Leaf, error) {
	tree, err := getAccountTree(db)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}
	if tree.GetAddressMap() != nil {
		accountId,ok:= tree.GetAddressMap()[ethAddress]
		if ok {
			return tree.GetLeaveMap()[accountId], nil
		}
	}
	return getLeafByEthAddressOnDB(db, ethAddress)
}

func getLeavesByRoot(table *table.Table, root []byte) ([]*ety.Leaf, error) {
	rows, err := table.ListIndex("root_hash", root, nil, 2000, dbm.ListASC)
	if err != nil {
		return nil, err
	}
	leaves := make([]*ety.Leaf, 0)
	for _, row := range rows {
		data := row.Data.(*ety.Leaf)
		leaves = append(leaves, data)
	}
	return leaves, nil
}

// UpdateLeaf 更新叶子结点：1、如果在当前树的叶子中，直接更新  2、如果在归档的树中，需要找到归档的root，重新生成root
func UpdateLeaf(db dbm.KV, accountId int32, chainType string, tokenId int32, balance int64) error {
	tree, err := getAccountTree(db)
	if err != nil {
		return errors.Wrapf(err, "db.getAccountTree")
	}

	//从当前的树里面找
	if leaf, ok := tree.GetLeaveMap()[accountId]; ok {
		updateBalance(leaf, chainType, tokenId, balance)
		currentTree := getNewTree()
		for _, leafVal := range tree.GetLeaveMap() {
			currentTree.Push(getLeafByte(leafVal))
		}

		subtrees := make([]*ety.SubTree, 0)
		for _, subtree := range currentTree.GetAllSubTrees() {
			subtrees = append(subtrees, &ety.SubTree{
				RootHash: subtree.GetSum(),
				Height:   int32(subtree.GetHeight()),
			})
		}

		tree.SubTrees = subtrees
		err = setAccountTree(db, tree)
		if err != nil {
			return errors.Wrapf(err, "db.setAccountTree")
		}
		return nil
	}

	//从归档里面找
	leaf, err := getLeafByEthAddress(db, ethAddress)
	if err != nil {
		return errors.Wrapf(err, "db.getLeafByEthAddress")
	}
	if leaf == nil {
		return errors.New("account not exist")
	}
	accountTable := NewAccountTreeTable(db)

	//找到对应的根
	if rootInfo, ok := tree.GetRootMap()[hex.EncodeToString(leaf.GetRootHash())]; ok {
		leaves, err := getLeavesByRoot(accountTable, leaf.GetRootHash())
		if err != nil {
			return errors.Wrapf(err, "db.getLeavesByRoot")
		}
		oldRootHash := leaf.GetRootHash()
		currentTree := getNewTree()
		for _, leafVal := range leaves {
			if leafVal.GetAccountId() == leaf.GetAccountId() {
				updateBalance(leafVal, chainType, tokenId, balance)
			}
			currentTree.Push(getLeafByte(leafVal))
		}
		rootHash := currentTree.Root()
		for _, leafVal := range leaves {
			leafVal.RootHash = rootHash
		}
		err = updateLeaves(accountTable, leaves)
		if err != nil {
			return errors.Wrapf(err, "db.updateLeaves")
		}
		//落盘保存
		err = SaveAccountTreeTable(db, accountTable)
		if err != nil {
			return errors.Wrapf(err, "db.SaveAccountTreeTable")
		}
		//更新树
		delete(tree.RootMap, hex.EncodeToString(oldRootHash))
		tree.RootMap[hex.EncodeToString(rootHash)] = rootInfo
		tree.RootIndexMap[rootInfo.StartIndex/tree.MaxCurrentIndex] = rootHash
		err = setAccountTree(db, tree)
		if err != nil {
			return errors.Wrapf(err, "db.setAccountTree")
		}
	} else {
		return errors.Wrapf(err, "rootMap.notFindRoot")
	}
	return nil
}

func getLeafHash(leaf *ety.Leaf) []byte {
	hash := mimc.NewMiMC(mixTy.MimcHashSeed)
	hash.Write([]byte(string(leaf.GetAccountId())))
	hash.Write([]byte(leaf.GetEthAddress()))
	hash.Write([]byte(leaf.GetPublicKey().GetX() + leaf.GetPublicKey().GetY()))
	for _, balance := range leaf.GetChainBalances() {
		hash.Write(balance.GetRootHash())
	}
	return hash.Sum(nil)
}

func setChainBalanceRootHash(balance *ety.ChainBalance) {
	tree := getNewTree()
	for _, tokenBalance := range balance.GetTokenBalances() {
		hash := mimc.NewMiMC(mixTy.MimcHashSeed)
		hash.Write([]byte(string(tokenBalance.GetTokenId())))
		hash.Write(big.NewInt(tokenBalance.GetBalance()).Bytes())
		tree.Push(hash.Sum(nil))
	}
	balance.RootHash = tree.Root()
}

func updateBalance(leaf *ety.Leaf, chainType string, tokenId int32, balance int64) {
	if chain, ok := leaf.GetChainBalanceMap()[chainType]; ok {
		if token, ok := chain.GetTokenBalanceMap()[tokenId]; ok {
			token.Balance += balance
		} else {
			tokenBalance := &ety.TokenBalance{TokenId: tokenId, Balance: balance}
			chain.TokenBalanceMap[tokenId] = tokenBalance
		}
	} else {
		tokenBalance := &ety.TokenBalance{TokenId: tokenId, Balance: balance}
		tokenBalanceMap := make(map[int32]*ety.TokenBalance)
		tokenBalanceMap[tokenId] = tokenBalance
		chainBalance := &ety.ChainBalance{ChainType: chainType, TokenBalanceMap: tokenBalanceMap}
		leaf.ChainBalanceMap[chainType] = chainBalance
	}
}

func CalProof(db dbm.KV, ethAddress string) (*ety.AccountTreeProof, error) {
	tree, err := getAccountTree(db)
	if err != nil {
		return nil, errors.Wrapf(err, "db.getAccountTree")
	}
	var leaf *ety.Leaf
	//如果accountId在未归档中，从当前找，否则从localdb找
	if v, ok := tree.GetLeaveMap()[ethAddress]; ok {
		leaf = v
	} else {
		leaf , err = getLeafByEthAddress(db, ethAddress)
		if err != nil {
			return nil, errors.Wrapf(err, "db.getLeafByEthAddress")
		}
	}
	if leaf == nil {
		return nil, errors.New("account not exist")
	}

	currentTree := getNewTree()
	err = currentTree.SetIndex(uint64(leaf.GetAccountId() - 1))
	if err != nil {
		return nil, errors.Wrapf(err, "merkleTree.setIndex")
	}
	accountTable := NewAccountTreeTable(db)
	for i := int32(0); i < tree.TotalIndex/tree.MaxCurrentIndex; i++ {
		//如果需要验证的account在该归档节点中，需要捞出来所有root下的leaf的进行push
		if i == leaf.GetAccountId() / tree.MaxCurrentIndex {
			leaves, err := getLeavesByRoot(accountTable, leaf.GetRootHash())
			if err != nil {
				return nil, errors.Wrapf(err, "db.getLeavesByRoot")
			}
			for _,v:=range leaves {
				fmt.Print("account Id ", v.GetAccountId())
				currentTree.Push(getLeafByte(v))
			}
		} else {
			err = currentTree.PushSubTree(10, tree.RootIndexMap[i])
		}
	}

	for _, v := range tree.GetLeaveMap() {
		currentTree.Push(getLeafByte(v))
	}

	rootHash, proofSet, proofIndex, numLeaves := currentTree.Prove()
	return &ety.AccountTreeProof{RootHash: rootHash, ProofSet: proofSet, ProofIndex: proofIndex, NumLeaves: numLeaves}, nil
}
