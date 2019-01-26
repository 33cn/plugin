package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

func TestCalcFrozen(t *testing.T) {
	types.SetTitleOnlyForTest("chain33")
	m, err := newMeans("LeftProportion", 15000000)
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

func TestLeftV1(t *testing.T) {
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
		t.Run("test LeftProportionV1", func(t *testing.T) {
			create := pty.UnfreezeCreate{
				StartTime:   c.start,
				AssetExec:   "coins",
				AssetSymbol: "bty",
				TotalCount:  c.total,
				Beneficiary: "x",
				Means:       pty.LeftProportionX,
				MeansOpt: &pty.UnfreezeCreate_LeftProportion{
					LeftProportion: &pty.LeftProportion{
						Period:        c.period,
						TenThousandth: c.tenThousandth,
					},
				},
			}
			u := &pty.Unfreeze{
				TotalCount: c.total,
				Means:      pty.LeftProportionX,
				StartTime:  c.start,
				MeansOpt: &pty.Unfreeze_LeftProportion{
					LeftProportion: &pty.LeftProportion{
						Period:        c.period,
						TenThousandth: c.tenThousandth,
					},
				},
			}
			m := leftProportion{}
			u, err := m.setOpt(u, &create)
			assert.Nil(t, err)

			f, err := m.calcFrozen(u, c.now)
			assert.Nil(t, err)

			assert.Equal(t, c.expect, f)

		})
	}
}

func TestFixV1(t *testing.T) {
	cases := []struct {
		start  int64
		now    int64
		period int64
		total  int64
		amount int64
		expect int64
	}{
		{10000, 10001, 10, 10000, 2, 9998},
		{10000, 10011, 10, 10000, 2, 9996},
		{10000, 10001, 10, 1e17, 2, 1e17 - 2},
		{10000, 10011, 10, 1e17, 2, 1e17 - 4},
	}

	for _, c := range cases {
		c := c
		t.Run("test FixAmountV1", func(t *testing.T) {
			create := pty.UnfreezeCreate{
				StartTime:   c.start,
				AssetExec:   "coins",
				AssetSymbol: "bty",
				TotalCount:  c.total,
				Beneficiary: "x",
				Means:       pty.FixAmountX,
				MeansOpt: &pty.UnfreezeCreate_FixAmount{
					FixAmount: &pty.FixAmount{
						Period: c.period,
						Amount: c.amount,
					},
				},
			}
			u := &pty.Unfreeze{
				TotalCount: c.total,
				Means:      pty.FixAmountX,
				StartTime:  c.start,
				MeansOpt: &pty.Unfreeze_FixAmount{
					FixAmount: &pty.FixAmount{
						Period: c.period,
						Amount: c.amount,
					},
				},
			}
			m := fixAmount{}
			u, err := m.setOpt(u, &create)
			assert.Nil(t, err)

			f, err := m.calcFrozen(u, c.now)
			assert.Nil(t, err)

			assert.Equal(t, c.expect, f)

		})
	}
}
