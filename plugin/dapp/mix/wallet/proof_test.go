package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCommitValue(t *testing.T) {
	var note, transfer, minFee uint64
	note = 100
	transfer = 60
	minFee = 1
	_, err := getCommitValue(note, transfer, minFee)
	assert.Nil(t, err)

	//transfer > note
	note = 100
	transfer = 100
	minFee = 1
	_, err = getCommitValue(note, transfer, minFee)
	t.Log(err)
	assert.NotNil(t, err)

	note = 100
	transfer = 101
	minFee = 0
	_, err = getCommitValue(note, transfer, minFee)
	t.Log(err)
	assert.NotNil(t, err)

	//change=0
	note = 100
	transfer = 99
	minFee = 1
	_, err = getCommitValue(note, transfer, minFee)
	assert.Nil(t, err)
}
