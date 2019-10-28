// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/trade/types"
)

/*
   在以前版本中只有token 合约发行的币在trade 里面交易， 订单中 symbol 为 token 的symbol，
   现在 symbol 扩展成 exec.sybol@title, @title 先忽略， (因为不必要, 只支持主链到平行链)。
   在订单中增加 exec， 表示币从那个合约中来的。

   在主链
     原来的订单  exec = "" symbol = "TEST"
     新的订单    exec =  "token"  symbol = "token.TEST"

   在平行链, 主链资产和本链资产的表示区别如下
     exec = "paracross"  symbol = "token.TEST"
     exec = "token"      symbol = "token.TEST"

*/

//GetExecSymbol : return exec, symbol
func GetExecSymbol(order *pt.SellOrder) (string, string) {
	if order.AssetExec == "" {
		return defaultAssetExec, defaultAssetExec + "." + order.TokenSymbol
	}
	return order.AssetExec, order.TokenSymbol
}

func checkAsset(cfg *types.Chain33Config, height int64, exec, symbol string) bool {
	if cfg.IsDappFork(height, pt.TradeX, pt.ForkTradeAssetX) {
		if exec == "" || symbol == "" {
			return false
		}
	} else {
		if exec != "" {
			return false
		}
	}
	return true
}

func checkPrice(cfg *types.Chain33Config, height int64, exec, symbol string) bool {
	if cfg.IsDappFork(height, pt.TradeX, pt.ForkTradePriceX) {
		if exec == "" && symbol != "" || exec != "" && symbol == "" {
			return false
		}
	} else {
		if exec != "" || symbol != "" {
			return false
		}
	}
	return true
}

func notSameAsset(cfg *types.Chain33Config, height int64, assetExec, assetSymbol, priceExec, priceSymbol string) bool {
	if cfg.IsDappFork(height, pt.TradeX, pt.ForkTradePriceX) {
		if assetExec == priceExec && assetSymbol == priceSymbol {
			return false
		}
	}
	return true
}

func createAccountDB(cfg *types.Chain33Config, height int64, db db.KV, exec, symbol string) (*account.DB, error) {
	if cfg.IsDappFork(height, pt.TradeX, pt.ForkTradeFixAssetDBX) {
		if exec == "" {
			exec = defaultAssetExec
		}
		return account.NewAccountDB(cfg, exec, symbol, db)
	} else if cfg.IsDappFork(height, pt.TradeX, pt.ForkTradeAssetX) {
		return account.NewAccountDB(cfg, exec, symbol, db)
	}

	return account.NewAccountDB(cfg, defaultAssetExec, symbol, db)
}

func createPriceDB(cfg *types.Chain33Config, height int64, db db.KV, exec, symbol string) (*account.DB, error) {
	if cfg.IsDappFork(height, pt.TradeX, pt.ForkTradePriceX) {
		// 为空默认使用 coins
		if exec == "" {
			acc := account.NewCoinsAccount(cfg)
			acc.SetDB(db)
			return acc, nil
		}
		return account.NewAccountDB(cfg, exec, symbol, db)
	}
	acc := account.NewCoinsAccount(cfg)
	acc.SetDB(db)
	return acc, nil
}
