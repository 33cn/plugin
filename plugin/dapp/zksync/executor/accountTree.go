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
	ethAddr := confManager.GStr(zt.ZkCfgEthFeeAddr)
	chain33Addr := confManager.GStr(zt.ZkCfgLayer2FeeAddr)
	if len(ethAddr) <= 0 || len(chain33Addr) <= 0 {
		panic(fmt.Sprintf("zksync not cfg init fee addr, ethAddr=%s,33Addr=%s", ethAddr, chain33Addr))
	}
	ethAddrDecimal, _ := zt.HexAddr2Decimal(ethAddr)
	chain33AddrDecimal, _ := zt.HexAddr2Decimal(chain33Addr)
	return ethAddrDecimal, chain33AddrDecimal
}

// 由于ethAddr+chain33Addr 唯一确定一个accountId,所以设置初始账户的chain33Addr不相同
func getInitAccountLeaf(ethFeeAddr, chain33FeeAddr string) []*zt.Leaf {
	zeroHash := zt.Str2Byte("0")
	defaultAccount := &zt.Leaf{
		EthAddress:  "0",
		AccountId:   zt.SystemDefaultAcctId,
		Chain33Addr: "3",
		TokenHash:   zeroHash,
	}

	//default system FeeAccount
	//feeAcct需要预设置缺省tokenId=0,balance=0,为了和deposit流程保持一致,deposit都会先有token更新再有pubkey更新
	//不然在token=null场景下设置pubkey,电路会计算出错，因为rhs部分会计算token tree part
	//缺省电路token tree都是0,会把token=0作为一个新node计算,而预设tokenId就可以解决这个问题
	feeAccount := &zt.Leaf{
		EthAddress:  ethFeeAddr,
		AccountId:   zt.SystemFeeAccountId,
		Chain33Addr: chain33FeeAddr,
		TokenHash:   zeroHash,
		TokenIds:    []uint64{0},
	}
	//default NFT system account
	NFTAccount := &zt.Leaf{
		EthAddress:  "0",
		AccountId:   zt.SystemNFTAccountId,
		Chain33Addr: "1",
		TokenHash:   zeroHash,
	}

	treeToContractAccount := &zt.Leaf{
		EthAddress:  ethFeeAddr,
		AccountId:   zt.SystemTree2ContractAcctId,
		Chain33Addr: "2",
		TokenHash:   zeroHash,
	}
	return []*zt.Leaf{defaultAccount, feeAccount, NFTAccount, treeToContractAccount}
}

// 获取系统初始root，如果未设置fee账户，缺省采用配置文件，
func getInitTreeRoot(cfg *types.Chain33Config, ethAddrDecimal, layer2AddrDecimal string) string {
	var feeEth, fee33 string
	if len(ethAddrDecimal) > 0 && len(layer2AddrDecimal) > 0 {
		feeEth, fee33 = ethAddrDecimal, layer2AddrDecimal
	} else {
		feeEth, fee33 = getCfgFeeAddr(cfg)
	}
	h := mimc.NewMiMC(zt.ZkMimcHashSeed)

	leafs := getInitAccountLeaf(feeEth, fee33)
	getInitLeafTokenHash(h, leafs)

	merkleTree := getNewTreeWithHash(h)
	for _, l := range leafs {
		merkleTree.Push(getLeafHash(h, l))
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

func getInitLeafTokenHash(h hash.Hash, leafs []*zt.Leaf) {
	for _, leaf := range leafs {
		if len(leaf.TokenIds) > 0 {
			if len(leaf.TokenIds) > 1 {
				panic("init token list should only one")
			}
			token := &zt.TokenBalance{TokenId: leaf.TokenIds[0], Balance: "0"}
			tokenHash := getTokenBalanceHash(h, token)
			//leaf.tokenHash 是对所有tokenTree做的一个hash，如果只有一个节点，也是对这一个节点做hash
			tree := getNewTreeWithHash(h)
			tree.Push(tokenHash)
			leaf.TokenHash = tree.Root()
		}
	}
	h.Reset()
}

func NewInitAccount(ethFeeAddr, chain33FeeAddr string) ([]*types.KeyValue, error) {
	if len(ethFeeAddr) <= 0 || len(chain33FeeAddr) <= 0 {
		return nil, errors.New("zksync default fee addr(ethFeeAddr,zkChain33FeeAddr) is nil")
	}
	var kvs []*types.KeyValue
	initLeafAccounts := getInitAccountLeaf(ethFeeAddr, chain33FeeAddr)

	for _, leaf := range initLeafAccounts {
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

		//对有tokenIds的leaf增加到db
		if len(leaf.TokenIds) > 0 {
			token := &zt.TokenBalance{
				TokenId: leaf.TokenIds[0],
				Balance: "0",
			}

			kv = &types.KeyValue{
				Key:   GetTokenPrimaryKey(leaf.AccountId, token.TokenId),
				Value: types.Encode(token),
			}
			kvs = append(kvs, kv)
		}
	}

	return kvs, nil
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

func AddNewLeafOpt(ethAddress string, tokenId, accountId uint64, amount string, chain33Addr string) []*types.KeyValue {
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

	return kvs
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
	balancekv, balanceHistory, err := updateTokenBalance(accountID, tokenID, amount, option, statedb)
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
	l2Log.BalanceBefore = balanceHistory.before
	l2Log.BalanceAfter = balanceHistory.after

	if makeEncode {
		log = &types.ReceiptLog{
			Log: types.Encode(l2Log),
		}
	}

	return kvs, log, l2Log, nil
}

func applyL2AccountCreate(accountID, tokenID uint64, amount, ethAddress, chain33Addr string, statedb dbm.KV, makeEncode bool) ([]*types.KeyValue, *types.ReceiptLog, *zt.AccountTokenBalanceReceipt, error) {
	var kvs []*types.KeyValue
	var log *types.ReceiptLog
	//如果NFTAccountId第一次初始化token，因为缺省初始balance为（SystemNFTTokenId+1),这里add时候默认为+2
	if accountID == zt.SystemNFTAccountId && tokenID == zt.SystemNFTTokenId {
		amount = new(big.Int).SetUint64(zt.SystemNFTTokenId + 2).String()
	}

	kvs = append(kvs, AddNewLeafOpt(ethAddress, tokenID, accountID, amount, chain33Addr)...)

	//设置新账户的ID.
	newAccountKV := CalcNewAccountIDkv(int64(accountID))
	kvs = append(kvs, newAccountKV)

	l2Log := &zt.AccountTokenBalanceReceipt{}
	l2Log.EthAddress = ethAddress
	l2Log.Chain33Addr = chain33Addr
	l2Log.TokenId = tokenID
	l2Log.AccountId = accountID
	l2Log.BalanceBefore = "0"
	l2Log.BalanceAfter = amount

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

func getNewTreeWithHash(h hash.Hash) *merkletree.Tree {
	h.Reset()
	return merkletree.New(h)
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
	rows, err := accountTable.ListIndex("eth_address", []byte(fmt.Sprintf("%s", ethAddress)), nil, 1000, dbm.ListASC)

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
		data.EthAddress, _ = zt.DecimalAddr2Hex(data.GetEthAddress(), zt.EthAddrLen)
		data.Chain33Addr, _ = zt.DecimalAddr2Hex(data.GetChain33Addr(), zt.BTYAddrLen)
		datas = append(datas, data)
	}
	return datas, nil
}

func GetLeafByChain33Address(db dbm.KV, chain33Addr string) ([]*zt.Leaf, error) {
	accountTable := NewAccountTreeTable(db)
	rows, err := accountTable.ListIndex("chain33_address", []byte(fmt.Sprintf("%s", chain33Addr)), nil, 1000, dbm.ListASC)

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
		data.EthAddress, _ = zt.DecimalAddr2Hex(data.GetEthAddress(), zt.EthAddrLen)
		data.Chain33Addr, _ = zt.DecimalAddr2Hex(data.GetChain33Addr(), zt.BTYAddrLen)
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

func getLeafHash(h hash.Hash, leaf *zt.Leaf) []byte {
	//h := mimc.NewMiMC(zt.ZkMimcHashSeed)
	h.Reset()
	accountIdBytes := new(fr.Element).SetUint64(leaf.GetAccountId()).Bytes()
	h.Write(accountIdBytes[:])
	h.Write(zt.Str2Byte(leaf.GetEthAddress()))
	h.Write(zt.Str2Byte(leaf.GetChain33Addr()))

	getLeafPubKeyHash(h, leaf.GetPubKey())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetNormal())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSystem())
	getLeafPubKeyHash(h, leaf.GetProxyPubKeys().GetSuper())

	h.Write(leaf.GetTokenHash())
	sum := h.Sum(nil)
	h.Reset()
	return sum
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
