package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_decimals(t *testing.T) {
	value2comp := 10
	for i := 1; i <= 10; i++ {
		value, ok := Decimal2value[i]
		require.Equal(t, value, int64(value2comp))
		require.Equal(t, ok, true)
		fmt.Println("value=", value)
		value2comp = value2comp * 10
	}
}
