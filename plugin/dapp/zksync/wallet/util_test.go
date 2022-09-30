package wallet

import (
	"testing"

	"github.com/33cn/chain33/common"
	"github.com/stretchr/testify/assert"
)

func TestGetEddsaPriKeySeed(t *testing.T) {
	key := "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	seed, err := GetLayer2PrivateKeySeed(key, "", "")
	assert.Nil(t, err)
	t.Log("seed", common.ToHex(seed))
}
