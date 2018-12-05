package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

func TestCalcFrozen(t *testing.T) {
	m, err := newMeans("LeftProportion")
	assert.Nil(t, err)
	assert.NotNil(t, m)

	cases := []struct {
		start         int64
		now           int64
		period        int64
		total         int64
		tenThousandth int64
		expect        int64
	}{
		{10000, 10001, 10, 10000, 2, 9998},
		{10000, 10011, 10, 10000, 2, 9996},
		{10000, 10001, 10, 1e17, 2, 9998 * 1e13},
		{10000, 10011, 10, 1e17, 2, 9998 * 9998 * 1e9},
	}

	for _, c := range cases {
		c := c
		t.Run("test LeftProportion", func(t *testing.T) {
			create := pty.UnfreezeCreate{
				StartTime:   c.start,
				AssetExec:   "coins",
				AssetSymbol: "bty",
				TotalCount:  c.total,
				Beneficiary: "x",
				Means:       "LeftProportion",
				MeansOpt: &pty.UnfreezeCreate_LeftProportion{
					LeftProportion: &pty.LeftProportion{
						Period:        c.period,
						TenThousandth: c.tenThousandth,
					},
				},
			}
			u := &pty.Unfreeze{
				TotalCount: c.total,
				Means:      "LeftProportion",
				StartTime:  c.start,
				MeansOpt: &pty.Unfreeze_LeftProportion{
					LeftProportion: &pty.LeftProportion{
						Period:        c.period,
						TenThousandth: c.tenThousandth,
					},
				},
			}
			u, err = m.setOpt(u, &create)
			assert.Nil(t, err)

			f, err := m.calcFrozen(u, c.now)
			assert.Nil(t, err)

			assert.Equal(t, c.expect, f)

		})
	}
}
