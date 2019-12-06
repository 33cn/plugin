package executor

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	a:=float32(1.00000212000000000001)
	b:=float32(0.34567)
	c:=float32(1234)
	t.Log(Truncate(a))
	t.Log(Truncate(b))
	t.Log(Truncate(c))
}