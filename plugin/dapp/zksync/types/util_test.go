package types

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitNFTContent(t *testing.T) {
	hash := "7b8c47ff0f29187c4fd7b9404d6d8671c3a05d041a2126753722fe940f30e2d3"
	fmt.Println("len", len(hash))
	a, b, err := SplitNFTContent(hash)
	assert.Nil(t, err)
	t.Log("a", a.Text(16), "b", b.Text(16))
	t.Log("a", a.BitLen(), "b", b.BitLen())
}
