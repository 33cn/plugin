package state

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewContractAccount(t *testing.T) {
	contractAccount := NewContractAccount("", nil)
	if nil != contractAccount {
		assert.Equal(t, nil, contractAccount)
	}
}
