package executor

import (
	"fmt"
	"github.com/33cn/chain33/util"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccountTree(t *testing.T) {
	dir, statedb, localdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, statedb)
	NewAccountTree(localdb)
	tree, err := getAccountTree(localdb)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tree)
	for i := 0; i < 10000; i++ {
		leaf := &et.Leaf{
			Hash: append([]byte("0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"), byte(i)),
			AccountId: int32(i + 1),
		}
		err = AddNewLeaf(localdb, leaf)
		assert.Equal(t, nil, err)
	}
	tree, err = getAccountTree(localdb)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, 1024, tree.GetIndex())
	for i := 0; i < 10000; i++ {
		leaf := &et.Leaf{
			Hash:      append([]byte("0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81c"), byte(i)),
			AccountId: int32(i + 1),
		}
		err = UpdateLeaf(localdb, leaf)
		assert.Equal(t, nil, err)
	}
	proof, err:= CalProof(localdb, 5)
	fmt.Print(proof)
}


