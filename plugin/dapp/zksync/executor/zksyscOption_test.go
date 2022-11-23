package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckSystemFeeAcctGap(t *testing.T) {
	sysFeeTokens := make(map[uint64]string)
	tokensGap := make(map[uint64]string)

	sysFeeTokens[0] = "100"
	sysFeeTokens[1] = "200"
	tokensGap[0] = "190"
	tokensGap[1] = "200"
	tokensGap[2] = "300"

	_, rst := checkSystemFeeAcctGap(sysFeeTokens, tokensGap)
	assert.Equal(t, "tokenId=0,gap=90,tokenId=2,gap=300", rst)

	tokensGap[0] = "90"
	sysFeeTokens[2] = "360"
	gapData, rst := checkSystemFeeAcctGap(sysFeeTokens, tokensGap)
	t.Log(gapData)
	//t.Log(rst)
	assert.Equal(t, 0, len(rst))
	assert.Equal(t, "90", gapData[0].NeedRollback)
	assert.Equal(t, "200", gapData[1].NeedRollback)
	assert.Equal(t, 0, len(gapData[1].Gap))
	assert.Equal(t, "300", gapData[2].NeedRollback)

}
