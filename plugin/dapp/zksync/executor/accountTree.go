package executor

import (
	"fmt"
	"hash"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/pkg/errors"
)

// TreeUpdateInfo 更新信息，用于查询
type TreeUpdateInfo struct {
	updateMap    map[string][]byte
	kvs          []*types.KeyValue
	localKvs     []*types.KeyValue
	accountTable *table.Table
}

func getCfgFeeAddr(cfg *types.Chain33Config) (string, string) {
	confManager := types.ConfSub(cfg, zt.Zksync)
	ethAddr := confManager.GStr("ethFeeAddr")
	chain33Addr := confManager.GStr("zkChain33FeeAddr")
	if len(ethAddr) <= 0 || len(chain33Addr) <= 0 {
		panic(fmt.Sprintf("zksync not cfg init fee addr, ethAddr=%s,33Addr=%s", ethAddr, chain33Addr))
	}
	return zt.HexAddr2Decimal(ethAddr), zt.HexAddr2Decimal(chain33Addr)
}

func getInitAccountLeaf(ethFeeAddr, chain33FeeAddr string) []*zt.Leaf {
	//default system FeeAccount
	feeAccount := &zt.Leaf{
		EthAddress:  ethFeeAddr,
		AccountId:   zt.SystemFeeAccountId,
		Chain33Addr: chain33FeeAddr,
		TokenHash:   "0",
	}
	//default NFT system account
	NFTAccount := &zt.Leaf{
		EthAddress:  "0",
		AccountId:   zt.SystemNFTAccountId,
		Chain33Addr: "0",
		TokenHash:   "0",
	}
	return []*zt.Leaf{feeAccount, NFTAccount}
}

//获取系统初始root，如果未设置fee账户，缺省采用配置文件，
func getInitTreeRoot(cfg *types.Chain33Config, ethAddr, chain33Addr string) string {
	var feeEth, fee33 string
	if len(ethAddr) > 0 && len(chain33Addr) > 0 {
		feeEth, fee33 = zt.HexAddr2Decimal(ethAddr), zt.HexAddr2Decimal(chain33Addr)
	} else {
		feeEth, fee33 = getCfgFeeAddr(cfg)
	}

	leafs := getInitAccountLeaf(feeEth, fee33)
	merkleTree := getNewTree()

	for _, l := range leafs {
		merkleTree.Push(getLeafHash(l))
	}
	tree := &zt.AccountTree{
		SubTrees: make([]*zt.SubTree, 0),
	}

	//叶子会按2^n合并，如果是三个leaf，就会产生2个subTree,这里当前初始只有2个leaf
	for _, subtree := range merkleTree.GetAllSubTrees() {
		tree.SubTrees = append(tree.SubTrees, &zt.SubTree{
			RootHash: subtree.GetSum(),
			Height:   int32(subtree.GetHeight()),
		})
	}

	return zt.Byte2Str(tree.SubTrees[len(tree.SubTrees)-1].RootHash)
}

// NewAccountTree 生成账户树，同时生成1号账户
func NewAccountTree(localDb dbm.KVDB, ethFeeAddr, chain33FeeAddr string) ([]*types.KeyValue, *table.Table) {
	if len(ethFeeAddr) <= 0 || len(chain33FeeAddr) <= 0 {
		panic(fmt.Sprintf("zksync default fee addr(ethFeeAddr,zkChain33FeeAddr) is nil"))
	}
	var kvs []*types.KeyValue
	initLeafAccounts := getInitAccountLeaf(ethFeeAddr, chain33FeeAddr)
	leafFeeAccount := initLeafAccounts[0]
	kv := &types.KeyValue{
		Key:   GetAccountIdPrimaryKey(leafFeeAccount.AccountId),
		Value: types.Encode(leafFeeAccount),
	}
	kvs = append(kvs, kv)

	kv = &types.KeyValue{
		Key:   GetChain33EthPrimaryKey(leafFeeAccount.Chain33Addr, leafFeeAccount.EthAddress),
		Value: types.Encode(leafFeeAccount),
	}
	kvs = append(kvs, kv)

	//NFT account
	leafNFTAccount := initLeafAccounts[1]
	kv = &types.KeyValue{
		Key:   GetAccountIdPrimaryKey(leafNFTAccount.AccountId),
		Value: types.Encode(leafNFTAccount),
	}
	kvs = append(kvs, kv)

	kv = &types.KeyValue{
		Key:   GetChain33EthPrimaryKey(leafNFTAccount.Chain33Addr, leafNFTAccount.EthAddress),
		Value: types.Encode(leafNFTAccount),
	}
	kvs = append(kvs, kv)

	accountTable := NewAccountTreeTable(localDb)
	err := accountTable.Add(leafFeeAccount)
	if err != nil {
		panic(err)
	}
	err = accountTable.Add(leafNFTAccount)
	if err != nil {
		panic(err)
	}

	merkleTree := getNewTree()
	merkleTree.Push(getLeafHash(leafFeeAccount))
	merkleTree.Push(getLeafHash(leafNFTAccount))

	tree := &zt.AccountTree{
		Index:           2,
		TotalIndex:      2,
		MaxCurrentIndex: 1024,
		SubTrees:        make([]*zt.SubTree, 0),
	}

	for _, subtree := range merkleTree.GetAllSubTrees() {
		tree.SubTrees = append(tree.SubTrees, &zt.SubTree{
			RootHash: subtree.GetSum(),
			Height:   int32(subtree.GetHeight()),
		})
	}

	kv = &types.KeyValue{
		Key:   GetAccountTreeKey(),
		Value: types.Encode(tree),
	}
	kvs = append(kvs, kv)

	return kvs, accountTable
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
		EthAddress:   ethAddress,
		AccountId:    tree.GetTotalIndex(),
		Chain33Addr:  chain33Addr,
		TokenIds:     make([]uint64, 0),
		ProxyPubKeys: new(zt.AccountProxyPubKeys),
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
	if info.accountTable != nil {
		accountTable = info.accountTable
	}
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
			return kvs, localKvs, errors.New(fmt.Sprintf("token not exist,tokenId=%d", tokenId))
		} else {
			token = &zt.TokenBalance{
				TokenId: tokenId,
				Balance: amount,
			}
			//如果NFTAccountId第一次初始化token，需要设置特殊balance作为新NFT token ID
			if accountId == zt.SystemNFTAccountId && tokenId == zt.SystemNFTTokenId {
				token.Balance = new(big.Int).SetUint64(zt.SystemNFTTokenId + 2).String()
			}
			leaf.TokenIds = append(leaf.TokenIds, tokenId)
		}
	} else {
		balance, _ := new(big.Int).SetString(token.GetBalance(), 10)
		change, _ := new(big.Int).SetString(amount, 10)
		if option == zt.Add {
			token.Balance = new(big.Int).Add(balance, change).String()
		} else if option == zt.Sub {
			if balance.Cmp(change) < 0 {
				return nil, nil, errors.Wrapf(types.ErrNotAllow, "amount=%d,tokenId=%d,balance=%s less sub amoumt=%s",
					accountId, tokenId, balance, amount)
			}
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
	if info.accountTable != nil {
		accountTable = info.accountTable
	}
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
	h := mimc.NewMiMC(zt.ZkMimcHashSeed)
	accountIdBytes := new(fr.Element).SetUint64(leaf.GetAccountId()).Bytes()
	h.Write(accountIdBytes[:])
	h.Write(zt.Str2Byte(leaf.GetEthAddress()))
	h.Write(zt.Str2Byte(leaf.GetChain33Addr()))

	getLeafPubKeyHash(h, leaf.GetPubKey())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetNormal())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSystem())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSuper())

	token := zt.Str2Byte(leaf.GetTokenHash())
	h.Write(token)
	return h.Sum(nil)
}

func getLeafPubKeyHash(h hash.Hash, pubKey *zt.ZkPubKey) {
	if pubKey != nil {
		h.Write(zt.Str2Byte(pubKey.GetX()))
		h.Write(zt.Str2Byte(pubKey.GetY()))
		return
	}

	h.Write(zt.Str2Byte("0")) //X
	h.Write(zt.Str2Byte("0")) //Y
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
	h := mimc.NewMiMC(zt.ZkMimcHashSeed)
	tokenIdBytes := new(fr.Element).SetUint64(token.GetTokenId()).Bytes()
	h.Write(tokenIdBytes[:])
	h.Write(zt.Str2Byte(token.Balance))
	return h.Sum(nil)
}

func getHistoryLeafHash(leaf *zt.HistoryLeaf) []byte {

	h := mimc.NewMiMC(zt.ZkMimcHashSeed)
	accountIdBytes := new(fr.Element).SetUint64(leaf.GetAccountId()).Bytes()
	h.Write(accountIdBytes[:])
	h.Write(zt.Str2Byte(leaf.GetEthAddress()))
	h.Write(zt.Str2Byte(leaf.GetChain33Addr()))
	if leaf.GetPubKey() != nil {
		h.Write(zt.Str2Byte(leaf.GetPubKey().GetX()))
		h.Write(zt.Str2Byte(leaf.GetPubKey().GetY()))
	} else {
		h.Write(zt.Str2Byte("0")) //X
		h.Write(zt.Str2Byte("0")) //Y
	}
	getLeafPubKeyHash(h, leaf.GetPubKey())
	if leaf.GetProxyPubKeys() != nil {
		getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetNormal())
		getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSystem())
		getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSuper())
	}

	tokenTree := getNewTree()
	for _, token := range leaf.Tokens {
		tokenTree.Push(getTokenBalanceHash(token))
	}
	h.Write(tokenTree.Root())
	return h.Sum(nil)
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

func UpdatePubKey(statedb dbm.KV, localdb dbm.KV, info *TreeUpdateInfo, pubKeyTy uint64, pubKey *zt.ZkPubKey, accountId uint64) ([]*types.KeyValue, []*types.KeyValue, error) {
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
	if nil == leaf.ProxyPubKeys {
		leaf.ProxyPubKeys = &zt.AccountProxyPubKeys{}
	}
	switch pubKeyTy {
	case 0:
		leaf.PubKey = pubKey
	case zt.NormalProxyPubKey:
		leaf.ProxyPubKeys.Normal = pubKey
	case zt.SystemProxyPubKey:
		leaf.ProxyPubKeys.System = pubKey
	case zt.SuperProxyPubKey:
		leaf.ProxyPubKeys.Super = pubKey
	default:
		return nil, nil, errors.Wrapf(types.ErrInvalidParam, "wrong pubkey ty=%d", pubKeyTy)
	}

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
