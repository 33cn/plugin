package spot

import (
	"fmt"

	"github.com/33cn/chain33/account"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func NewZkAsset(left uint64) *et.ZkAsset {
	return &et.ZkAsset{
		Ty:    et.AssetType_L1Erc20,
		Value: &et.ZkAsset_ZkAssetid{ZkAssetid: left},
	}
}

func NewZkNftAsset(left uint64) *et.ZkAsset {
	return &et.ZkAsset{
		Ty:    et.AssetType_ZkNft,
		Value: &et.ZkAsset_ZkAssetid{ZkAssetid: left},
	}
}

func NewEvmNftAsset(left uint64) *et.ZkAsset {
	return &et.ZkAsset{
		Ty:    et.AssetType_EvmNft,
		Value: &et.ZkAsset_EvmNftID{EvmNftID: left},
	}
}

//  support kinds of asset
type AssetAccount interface {
	Transfer(to AssetAccount, amountt int64) (*types.Receipt, error)
	TransferFrozen(to AssetAccount, amountt int64) (*types.Receipt, error)

	Frozen(amount int64) (*types.Receipt, error)
	UnFrozen(amount int64) (*types.Receipt, error)
	CheckBalance(amount int64) error

	GetCoinPrecision() int64
	GetAccountInfo() AccountInfo
}

type AccountInfo struct {
	address string
	accid   uint64
	asset   *et.ZkAsset
}

// support nft asset from evm contract
type NftAccount struct {
	accdb *EvmxgoNftAccountRepo
	AccountInfo

	nftid  uint64
	symbol string
	acc    *account.DB
}

type EvmxgoNftAccountRepo struct {
	cfg     *types.Chain33Config
	statedb dbm.KV
	//symbol  string
	accdb    *account.DB
	execAddr string // TODO evmxgo contract address
}

func (acc *NftAccount) GetCoinPrecision() int64 {
	return 1
}

func (acc *NftAccount) TransferFrozen(to AssetAccount, amountt int64) (*types.Receipt, error) {
	toAddr := to.GetAccountInfo().address
	return acc.acc.ExecTransferFrozen(acc.address, toAddr, acc.accdb.execAddr, amountt)
}
func (acc *NftAccount) Frozen(amount int64) (*types.Receipt, error) {
	return acc.acc.ExecFrozen(acc.address, acc.accdb.execAddr, amount)
}
func (acc *NftAccount) Transfer(to AssetAccount, amount int64) (*types.Receipt, error) {
	toAddr := to.GetAccountInfo().address
	return acc.acc.ExecTransfer(acc.address, toAddr, acc.accdb.execAddr, amount)
}
func (acc *NftAccount) UnFrozen(amount int64) (*types.Receipt, error) {
	return acc.acc.ExecActive(acc.address, acc.accdb.execAddr, amount)
}

func (acc *NftAccount) CheckBalance(amount int64) error {
	balance := acc.acc.LoadExecAccount(acc.address, acc.accdb.execAddr)
	if balance.Balance < amount {
		elog.Error("TokenAccount balance", "balance", balance.Balance, "need", amount)
		return et.ErrAssetBalance
	}
	return nil
}

func (acc *NftAccount) GetAccountInfo() AccountInfo {
	return acc.AccountInfo
}

func newEvmxgoNftAccountRepo(db dbm.KV, cfg *types.Chain33Config) (*EvmxgoNftAccountRepo, error) {
	return &EvmxgoNftAccountRepo{
		statedb: db,
		cfg:     cfg}, nil
}

func (accdb *EvmxgoNftAccountRepo) NewAccount(addr string, accid uint64, asset *et.ZkAsset) (*NftAccount, error) {
	nftid := asset.GetEvmNftID()
	var err error
	symbol := fmt.Sprintf("%d", nftid)
	if accdb.accdb == nil {
		accdb.accdb, err = account.NewAccountDB(accdb.cfg, "evmxgo", symbol, accdb.statedb)
		if err != nil {
			return nil, err
		}
	}
	accInfo := AccountInfo{
		address: addr,
		accid:   accid,
		asset:   asset,
	}

	return &NftAccount{accdb: accdb, AccountInfo: accInfo, nftid: nftid, symbol: symbol}, nil
}

// support go token from go contract
type TokenAccount struct {
	accdb *TokenAccountRepo
	AccountInfo
	execer string
	symbol string

	acc *account.DB
}

func GetCoinPrecision(ty int32) int64 {
	if ty == int32(et.AssetType_EvmNft) || ty == int32(et.AssetType_ZkNft) {
		return 1
	}
	// TODO
	return 1e8
}
func (acc *TokenAccount) GetCoinPrecision() int64 {
	return 1e8
}

func (acc *TokenAccount) TransferFrozen(to AssetAccount, amountt int64) (*types.Receipt, error) {
	toAddr := to.GetAccountInfo().address
	return acc.acc.ExecTransferFrozen(acc.address, toAddr, acc.accdb.execAddr, amountt)
}
func (acc *TokenAccount) Frozen(amount int64) (*types.Receipt, error) {
	return acc.acc.ExecFrozen(acc.address, acc.accdb.execAddr, amount)
}
func (acc *TokenAccount) Transfer(to AssetAccount, amount int64) (*types.Receipt, error) {
	toAddr := to.GetAccountInfo().address
	return acc.acc.ExecTransfer(acc.address, toAddr, acc.accdb.execAddr, amount)
}
func (acc *TokenAccount) UnFrozen(amount int64) (*types.Receipt, error) {
	return acc.acc.ExecActive(acc.address, acc.accdb.execAddr, amount)
}

func (acc *TokenAccount) CheckBalance(amount int64) error {
	balance := acc.acc.LoadExecAccount(acc.address, acc.accdb.execAddr)
	if balance.Balance < amount {
		elog.Error("TokenAccount balance", "balance", balance.Balance, "need", amount)
		return et.ErrAssetBalance
	}
	return nil
}

func (acc *TokenAccount) GetAccountInfo() AccountInfo {
	return acc.AccountInfo
}

type TokenAccountRepo struct {
	cfg      *types.Chain33Config
	statedb  dbm.KV
	execAddr string
}

func newTokenAccountRepo(db dbm.KV, cfg *types.Chain33Config, execAddr string) (*TokenAccountRepo, error) {
	return &TokenAccountRepo{
		statedb:  db,
		cfg:      cfg,
		execAddr: execAddr}, nil
}

func (accdb *TokenAccountRepo) NewAccount(addr string, accid uint64, asset *et.ZkAsset) (*TokenAccount, error) {
	accInfo := AccountInfo{
		address: addr,
		accid:   accid,
		asset:   asset,
	}
	acc := &TokenAccount{accdb: accdb, AccountInfo: accInfo, execer: asset.GetTokenAsset().Execer, symbol: asset.GetTokenAsset().Symbol}
	var err error
	acc.acc, err = account.NewAccountDB(accdb.cfg, asset.GetTokenAsset().Execer, asset.GetTokenAsset().Symbol, accdb.statedb)
	if err != nil {
		return nil, err
	}

	return acc, nil
}

type ZkAccount struct {
	acc *DexAccount
	AccountInfo
}

func (acc *ZkAccount) GetCoinPrecision() int64 {
	return 1e8
}

func (acc *ZkAccount) TransferFrozen(to AssetAccount, amount int64) (*types.Receipt, error) {
	panic("not support")
	//return acc.acc.ExecTransferFrozen(acc.address, to, acc.accdb.execAddr, amount)
}
func (acc *ZkAccount) Frozen(amount int64) (*types.Receipt, error) {
	return acc.acc.Frozen(acc.asset.GetZkAssetid(), uint64(amount))
}
func (acc *ZkAccount) Transfer(to AssetAccount, amount int64) (*types.Receipt, error) {
	panic("not support")
	//return acc.acc.ExecTransfer(acc.address, to, acc.accdb.execAddr, amount)
}
func (acc *ZkAccount) UnFrozen(amount int64) (*types.Receipt, error) {
	return acc.acc.Active(acc.asset.GetZkAssetid(), uint64(amount))
}

func (acc *ZkAccount) CheckBalance(amount int64) error {
	balance := acc.acc.GetBalance(acc.asset.GetZkAssetid())
	if balance < uint64(amount) {
		elog.Error("ZkAccount balance", "balance", balance, "need", amount)
		return et.ErrAssetBalance
	}
	return nil
}

func (acc *ZkAccount) GetAccountInfo() AccountInfo {
	return acc.AccountInfo
}
