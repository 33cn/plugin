package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMulCurvePointG(t *testing.T) {
	p := MulCurvePointG(3300000)
	assert.Equal(t, "16340309671848023141603674621476146712179749929747125480153351030768864391631", p.X.String())
	assert.Equal(t, "7282133334630698770430559968655675427988723121976895210725923327624387185615", p.Y.String())
}
