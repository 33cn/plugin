package executor

import (
	"github.com/33cn/chain33/account"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
)

func (r *rgbx) newAccount(symbol string) (*account.DB, error) {
	return account.NewAccountDB(r.GetAPI().GetConfig(), rtypes.RgbxX, formatSymbol(symbol), r.GetStateDB())
}
