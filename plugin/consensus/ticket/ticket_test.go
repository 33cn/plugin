package ticket

import (
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/util/testnode"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
)

// 执行： go test -cover
func TestTicket(t *testing.T) {
	cfg, sub := testnode.GetDefaultConfig()
	cfg.Consensus.Name = "ticket"
	mock33 := testnode.NewWithConfig(cfg, sub, nil)
	defer mock33.Close()
	err := mock33.WaitHeight(100)
	assert.Nil(t, err)
}
