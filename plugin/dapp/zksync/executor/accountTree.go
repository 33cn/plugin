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

type balancehistory struct {
	before string
	after  string
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
		TokenHash:   []byte("0"),
	}
	//default NFT system account
	NFTAccount := &zt.Leaf{
		EthAddress:  "0",
		AccountId:   zt.SystemNFTAccountId,
		Chain33Addr: "0",
		TokenHash:   []byte("0"),
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

func NewInitAccount(ethFeeAddr, chain33FeeAddr string) ([]*types.KeyValue, error) {
	if len(ethFeeAddr) <= 0 || len(chain33FeeAddr) <= 0 {
		return nil, errors.New("zksync default fee addr(ethFeeAddr,zkChain33FeeAddr) is nil")
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

	return kvs, nil
}

func deposit2NewAccount(accountId uint64, tokenId uint64, amount string) (*types.KeyValue, *balancehistory, error) {
	token := &zt.TokenBalance{
		TokenId: tokenId,
		Balance: amount,
	}
	//如果NFTAccountId第一次初始化token，因为缺省初始balance为（SystemNFTTokenId+1),这里add时候默认为+2
	if accountId == zt.SystemNFTAccountId && tokenId == zt.SystemNFTTokenId {
		token.Balance = new(big.Int).SetUint64(zt.SystemNFTTokenId + 2).String()
	}
	balanceInfoPtr := &balancehistory{}
	balanceInfoPtr.before = "0"
	balanceInfoPtr.after = token.Balance

	kv := &types.KeyValue{
		Key:   GetTokenPrimaryKey(accountId, tokenId),
		Value: types.Encode(token),
	}

	return kv, balanceInfoPtr, nil
}

func updateTokenBalance(accountId uint64, tokenId uint64, amount string, option int32, statedb dbm.KV) (*types.KeyValue, *balancehistory, error) {
	token, err := GetTokenByAccountIdAndTokenId(statedb, accountId, tokenId)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "updateTokenBalance.GetTokenByAccountIdAndTokenId")
	}

	balanceInfoPtr := &balancehistory{}

	if token == nil {
		if option == zt.Sub {
			return nil, nil, errors.Errorf("token not exist with Id=%d", tokenId)
		} else {
			token = &zt.TokenBalance{
				TokenId: tokenId,
				Balance: amount,
			}
			//如果NFTAccountId第一次初始化token，因为缺省初始balance为（SystemNFTTokenId+1),这里add时候默认为+2
			if accountId == zt.SystemNFTAccountId && tokenId == zt.SystemNFTTokenId {
				token.Balance = new(big.Int).SetUint64(zt.SystemNFTTokenId + 2).String()
			}
			balanceInfoPtr.before = "0"
			balanceInfoPtr.after = token.Balance
		}
	} else {
		balance, _ := new(big.Int).SetString(token.GetBalance(), 10)
		delta, _ := new(big.Int).SetString(amount, 10)
		balanceInfoPtr.before = token.GetBalance()
		if option == zt.Add {
			token.Balance = new(big.Int).Add(balance, delta).String()
		} else {
			if balance.Cmp(delta) < 0 {
				return nil, nil, errors.Wrapf(types.ErrNotAllow, "amount=%d,tokenId=%d,balance=%s less sub amoumt=%s",
					accountId, tokenId, balance, amount)
			}
			token.Balance = new(big.Int).Sub(balance, delta).String()
		}

		balanceInfoPtr.after = token.Balance
	}

	kv := &types.KeyValue{
		Key:   GetTokenPrimaryKey(accountId, tokenId),
		Value: types.Encode(token),
	}

	return kv, balanceInfoPtr, nil
}

func AddNewLeafOpt(statedb dbm.KV, ethAddress string, tokenId, accountId uint64, amount string, chain33Addr string) ([]*types.KeyValue, error) {
	var kvs []*types.KeyValue

	leaf := &zt.Leaf{
		EthAddress:  ethAddress,
		AccountId:   accountId,
		Chain33Addr: chain33Addr,
		TokenIds:    make([]uint64, 0),
		//ProxyPubKeys: new(zt.AccountProxyPubKeys),
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

	kv = &types.KeyValue{
		Key:   GetAccountIdPrimaryKey(leaf.AccountId),
		Value: types.Encode(leaf),
	}
	kvs = append(kvs, kv)

	kv = &types.KeyValue{
		Key:   GetChain33EthPrimaryKey(leaf.Chain33Addr, leaf.EthAddress),
		Value: types.Encode(leaf),
	}
	kvs = append(kvs, kv)

	return kvs, nil
}

func updateLeafOpt(statedb dbm.KV, leaf *zt.Leaf, tokenId uint64, option int32) ([]*types.KeyValue, error) {
	var kvs []*types.KeyValue

	token, err := GetTokenByAccountIdAndTokenId(statedb, leaf.AccountId, tokenId)
	if err != nil {
		return kvs, errors.Wrapf(err, "db.getAccountTree")
	}
	if token == nil && option == zt.Add {
		leaf.TokenIds = append(leaf.TokenIds, tokenId)
	}

	kv := &types.KeyValue{
		Key:   GetAccountIdPrimaryKey(leaf.AccountId),
		Value: types.Encode(leaf),
	}
	kvs = append(kvs, kv)

	kv = &types.KeyValue{
		Key:   GetChain33EthPrimaryKey(leaf.Chain33Addr, leaf.EthAddress),
		Value: types.Encode(leaf),
	}
	kvs = append(kvs, kv)

	return kvs, nil
}

func applyL2AccountUpdate(accountID, tokenID uint64, amount string, option int32, statedb dbm.KV, leaf *zt.Leaf, makeEncode bool) ([]*types.KeyValue, *types.ReceiptLog, *zt.AccountTokenBalanceReceipt, error) {
	var kvs []*types.KeyValue
	var log *types.ReceiptLog
	balancekv, balancehistory, err := updateTokenBalance(accountID, tokenID, amount, option, statedb)
	if nil != err {
		return nil, nil, nil, err
	}
	kvs = append(kvs, balancekv)

	updateLeafKvs, err := updateLeafOpt(statedb, leaf, tokenID, zt.Add)
	if nil != err {
		return nil, nil, nil, err
	}

	kvs = append(kvs, updateLeafKvs...)

	l2Log := &zt.AccountTokenBalanceReceipt{}
	l2Log.EthAddress = leaf.EthAddress
	l2Log.Chain33Addr = leaf.Chain33Addr
	l2Log.TokenId = tokenID
	l2Log.AccountId = accountID
	l2Log.BalanceBefore = balancehistory.before
	l2Log.BalanceAfter = balancehistory.after

	if makeEncode {
		log = &types.ReceiptLog{
			Log: types.Encode(l2Log),
		}
	}

	return kvs, log, l2Log, nil
}

func applyL2AccountCreate(accountID, tokenID uint64, amount, ethAddress,  chain33Addr string, statedb dbm.KV, makeEncode bool) ([]*types.KeyValue, *types.ReceiptLog, *zt.AccountTokenBalanceReceipt, error) {
	var kvs []*types.KeyValue
	var log *types.ReceiptLog
	balancekv, balancehistory, err := deposit2NewAccount(accountID, tokenID, amount)
	if nil != err {
		return nil, nil, nil, err
	}
	kvs = append(kvs, balancekv)

	addLeafKvs, err := AddNewLeafOpt(statedb, ethAddress, tokenID, accountID, amount, chain33Addr)
	if nil != err {
		return nil, nil, nil, err
	}
	kvs = append(kvs, addLeafKvs...)

	//设置新账户的ID.
	newAccountKV := CalcNewAccountIDkv(int64(accountID))
	kvs = append(kvs, newAccountKV)

	l2Log := &zt.AccountTokenBalanceReceipt{}
	l2Log.EthAddress = ethAddress
	l2Log.Chain33Addr = chain33Addr
	l2Log.TokenId = tokenID
	l2Log.AccountId = accountID
	l2Log.BalanceBefore = balancehistory.before
	l2Log.BalanceAfter = balancehistory.after

	//为了避免不必要的计算，在transfer等场景中，该操作在函数外部进行
	if makeEncode {
		log = &types.ReceiptLog{
			Log: types.Encode(l2Log),
		}
	}

	return kvs, log, l2Log, nil
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

func GetLeafByAccountId(db dbm.KV, accountId uint64) (*zt.Leaf, error) {
	if accountId <= 0 {
		return nil, nil
	}

	val, err := db.Get(GetAccountIdPrimaryKey(accountId))
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var leaf zt.Leaf
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

func GetLeafByChain33AndEthAddress(db dbm.KV, chain33Addr, ethAddress string) (*zt.Leaf, error) {
	if chain33Addr == "" || ethAddress == "" {
		return nil, types.ErrInvalidParam
	}

	val, err := db.Get(GetChain33EthPrimaryKey(chain33Addr, ethAddress))
	if err != nil {
		if err.Error() == types.ErrNotFound.Error() {
			return nil, nil
		} else {
			return nil, err
		}
	}

	var leaf zt.Leaf
	err = types.Decode(val, &leaf)
	if err != nil {
		return nil, err
	}
	return &leaf, nil
}

func GetLeavesByStartAndEndIndex(db dbm.KV, startIndex uint64, endIndex uint64) ([]*zt.Leaf, error) {
	leaves := make([]*zt.Leaf, 0)
	for i := startIndex; i <= endIndex; i++ {
		leaf, err := GetLeafByAccountId(db, i)
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

	h.Write(leaf.GetTokenHash())
	return h.Sum(nil)
}

func getHistoryLeafHash(leaf *zt.HistoryLeaf) []byte {
	h := mimc.NewMiMC(zt.ZkMimcHashSeed)
	accountIdBytes := new(fr.Element).SetUint64(leaf.GetAccountId()).Bytes()
	h.Write(accountIdBytes[:])
	h.Write(zt.Str2Byte(leaf.GetEthAddress()))
	h.Write(zt.Str2Byte(leaf.GetChain33Addr()))

	getLeafPubKeyHash(h, leaf.GetPubKey())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetNormal())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSystem())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSuper())

	h.Write(zt.Str2Byte(getHistoryTokenHash(leaf.AccountId, leaf.Tokens)))
	return h.Sum(nil)
}

func getHistoryTokenHash(accountId uint64, tokens []*zt.TokenBalance) string {
	if (accountId == zt.SystemFeeAccountId || accountId == zt.SystemNFTAccountId) && len(tokens) <= 0 {
		return "0"
	}

	tokenTree := getNewTree()
	for _, token := range tokens {
		tokenTree.Push(getTokenBalanceHash(token))
	}
	return zt.Byte2Str(tokenTree.Root())
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

func getTokenRootHash(db dbm.KV, accountId uint64, tokenIds []uint64) (string, error) {
	tree := getNewTree()
	for _, tokenId := range tokenIds {
		token, err := GetTokenByAccountIdAndTokenId(db, accountId, tokenId)
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

func UpdatePubKey(statedb dbm.KV, localdb dbm.KV, pubKeyTy uint64, pubKey *zt.ZkPubKey, accountId uint64) ([]*types.KeyValue, []*types.KeyValue, error) {
	var kvs []*types.KeyValue
	leaf, err := GetLeafByAccountId(statedb, accountId)
	if err != nil {
		return kvs, nil, errors.Wrapf(err, "db.GetTokenByAccountIdAndTokenId")
	}
	if leaf == nil {
		return kvs, nil, errors.New("account not exist")
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


	kv = &types.KeyValue{
		Key:   GetChain33EthPrimaryKey(leaf.Chain33Addr, leaf.EthAddress),
		Value: types.Encode(leaf),
	}

	kvs = append(kvs, kv)

	return kvs, nil, nil
}
