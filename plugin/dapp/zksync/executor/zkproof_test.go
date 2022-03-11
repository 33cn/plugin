package executor

import (
	"github.com/33cn/chain33/util"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestHistoryProof(t *testing.T) {
	dir, statedb, localdb := util.CreateTestDB()

	defer util.CloseTestDB(dir, statedb)

	proof, err := getAccountProofInHistory(localdb, 1, "")
	assert.Equal(t, nil, err)
	t.Log(proof)
}
