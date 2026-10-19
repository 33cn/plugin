package executor

import (
	"testing"

	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
	"github.com/stretchr/testify/assert"
)

// TestIsConfirmedUsesCurrentWeight 确认修复分叉后按 owner 当前权重累计，不再使用确认时的过期快照
func TestIsConfirmedUsesCurrentWeight(t *testing.T) {
	acc := &mty.MultiSig{
		RequiredWeight: 10,
		Owners: []*mty.Owner{
			{OwnerAddr: "ownerA", Weight: 5},
			{OwnerAddr: "ownerB", Weight: 5},
		},
	}
	tx := &mty.MultiSigTx{
		ConfirmedOwner: []*mty.Owner{
			{OwnerAddr: "ownerA", Weight: 5},
			{OwnerAddr: "ownerB", Weight: 5},
		},
	}
	assert.True(t, isConfirmed(10, tx, acc, false))
	assert.True(t, isConfirmed(10, tx, acc, true))

	// ownerB 被删除后：旧快照仍能凑足阈值，按当前 owner 权重则不能
	acc.Owners = acc.Owners[:1]
	assert.True(t, isConfirmed(10, tx, acc, false), "pre-fork keeps the stale snapshot behavior")
	assert.False(t, isConfirmed(10, tx, acc, true), "post-fork must ignore removed owner")

	// ownerA 被降权后，按当前权重重新累计也不能凑足阈值
	acc2 := &mty.MultiSig{RequiredWeight: 10, Owners: []*mty.Owner{{OwnerAddr: "ownerA", Weight: 2}}}
	assert.True(t, isConfirmed(10, tx, acc2, false))
	assert.False(t, isConfirmed(10, tx, acc2, true))
}
