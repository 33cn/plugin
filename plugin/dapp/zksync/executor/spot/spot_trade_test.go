package spot

import (
	"testing"

	"gotest.tools/assert"
)

type Idummy interface {
	Far() int
}

type dummy struct {
	i int
}

func (d *dummy) Far() int {
	return d.i
}

func Test_initTrade(t *testing.T) {
	var x, y Idummy
	x = &dummy{i: 2}
	y = x
	assert.Equal(t, x, y, "test dummy equal")
}
