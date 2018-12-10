package executor

import (
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/f3d/ptypes"
)

func (c *f3d) Exec_Start(payload *pt.F3DStart, tx *types.Transaction, index int) (*types.Receipt, error) {
	// 判断是否为首轮，如果是首轮则需要加载配置信息
	bFirstRound := false
	if bFirstRound {
		pt.SetConfig()
	}

	return &types.Receipt{}, nil
}

func (c *f3d) Exec_Draw(payload *pt.F3DLuckyDraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	return &types.Receipt{}, nil
}

func (c *f3d) Exec_Buy(payload *pt.F3DBuyKey, tx *types.Transaction, index int) (*types.Receipt, error) {
	return &types.Receipt{}, nil
}
