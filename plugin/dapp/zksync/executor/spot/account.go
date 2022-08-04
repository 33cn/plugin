package spot

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

// account repos -> asset_ty -> account repo
// load user
type accountRepos struct {
	zkRepo    *accountRepo
	tokenRepo *TokenAccountRepo
	evmxgo    *EvmxgoNftAccountRepo
}

func newAccountRepos(dexName string, statedb dbm.KV, p et.DBprefix, cfg *types.Chain33Config, execAddr string) (*accountRepos, error) {
	var repos accountRepos
	var err error
	repos.zkRepo = newAccountRepo(dexName, statedb, p)
	repos.tokenRepo, err = newTokenAccountRepo(statedb, cfg, execAddr)
	if err != nil {
		return nil, err
	}
	repos.evmxgo, err = newEvmxgoNftAccountRepo(statedb, cfg)
	if err != nil {
		return nil, err
	}

	return &repos, nil
}

type AssetAccounts struct {
	buyAcc  AssetAccount
	sellAcc AssetAccount
	buy     *et.ZkAsset
	sell    *et.ZkAsset
	same    bool
}

func (repos *accountRepos) LoadAccount(addr string, zkAccID uint64, asset *et.ZkAsset) (AssetAccount, error) {
	switch asset.Ty {
	case et.AssetType_L1Erc20:
		acc1, err := repos.zkRepo.LoadAccount(addr, zkAccID)
		if err != nil {
			return nil, err
		}
		info := AccountInfo{address: addr, accid: zkAccID, asset: asset}
		return &ZkAccount{acc: acc1, AccountInfo: info}, nil
	case et.AssetType_Token:
		acc, err := repos.tokenRepo.NewAccount(addr, zkAccID, asset)
		if err != nil {
			return nil, err
		}
		return acc, nil
	case et.AssetType_ZkNft:
		acc1, err := repos.zkRepo.LoadAccount(addr, asset.GetZkAssetid())
		if err != nil {
			return nil, err
		}
		info := AccountInfo{address: addr, accid: zkAccID, asset: asset}
		return &ZkAccount{acc: acc1, AccountInfo: info}, nil
	case et.AssetType_EvmNft:
		acc1, err := repos.evmxgo.NewAccount(addr, zkAccID, asset)
		if err != nil {
			return nil, err
		}
		return acc1, nil
	}
	panic("not support")

}

// account 是一个对象代表一个人的一个资产 (go/evm Asset)
// dexAccount 是一个对象代表一个人的所有资产 (L1 Asset or Asset in zkspot)
// 统一成一个对象 多种账号
// L1 资产是同账号管理的

func isZkAsset(ty1 et.AssetType) bool {
	return ty1 == et.AssetType_L1Erc20 || ty1 == et.AssetType_ZkNft
}

func sameAccountType(ty1, ty2 et.AssetType) bool {
	return isZkAsset(ty1) && isZkAsset(ty2)
}

func (repos *accountRepos) LoadAccounts(addr string, zkAccID uint64, buy, sell *et.ZkAsset) (*AssetAccounts, error) {
	acc1, err := repos.LoadAccount(addr, zkAccID, buy)
	if err != nil {
		return nil, err
	}
	accs := AssetAccounts{
		buyAcc: acc1,
		buy:    buy,
		sell:   sell,
	}
	if sameAccountType(buy.Ty, sell.Ty) {
		accs.sell = accs.buy
		accs.same = true
		return &accs, nil
	}
	acc2, err := repos.LoadAccount(addr, zkAccID, sell)
	if err != nil {
		return nil, err
	}
	accs.same = false
	accs.sellAcc = acc2

	return &accs, nil
}
