package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

//func TestSplitNFTContent(t *testing.T) {
//	hash := "7b8c47ff0f29187c4fd7b9404d6d8671c3a05d041a2126753722fe940f30e2d3"
//	fmt.Println("len", len(hash))
//	a, b, err := SplitNFTContent(hash)
//	assert.Nil(t, err)
//	t.Log("a", a.Text(16), "b", b.Text(16))
//	t.Log("a", a.BitLen(), "b", b.BitLen())
//}

func TestFindExponent(t *testing.T) {
	s := "12304"
	r := ZkFindExponentPart(s)
	assert.True(t, r == 0)

	s = "123040"
	r = ZkFindExponentPart(s)
	assert.True(t, r == 1)

	s = "0"
	r = ZkFindExponentPart(s)
	assert.True(t, r == 0)

	s = "12"
	for i := 0; i < 33; i++ {
		s += "0"
	}
	r = ZkFindExponentPart(s)
	assert.True(t, r == 31)
	//fmt.Println("s",s)
	//fmt.Println("s.len",len(s),"exp",r,"s",s[0:len(s)-r])

}

func TestFindManExpPart(t *testing.T) {
	s := "12304"
	m, e := ZkTransferManExpPart(s)
	assert.True(t, m == s)
	assert.True(t, e == 0)

	s = "123040"
	m, e = ZkTransferManExpPart(s)
	assert.True(t, m == "12304")
	assert.True(t, e == 1)

	s = "0"
	m, e = ZkTransferManExpPart(s)
	assert.True(t, m == "0")
	assert.True(t, e == 0)

	s = "12"
	for i := 0; i < 31; i++ {
		s += "0"
	}
	m, e = ZkTransferManExpPart(s)
	assert.True(t, m == "12")
	assert.True(t, e == 31)

	s = "12"
	for i := 0; i < 30; i++ {
		s += "0"
	}
	m, e = ZkTransferManExpPart(s)
	assert.True(t, m == "12")
	assert.True(t, e == 30)

	s = "12"
	for i := 0; i < 32; i++ {
		s += "0"
	}
	m, e = ZkTransferManExpPart(s)
	assert.True(t, m == "120")
	assert.True(t, e == 31)
}

func TestDecodePacVal(t *testing.T) {
	val := "0d02"
	bVal, err := hex.DecodeString(val)
	assert.Nil(t, err)
	rst := DecodePacVal(bVal, PacExpBitWidth)
	assert.Equal(t, "10400", rst)

	val = "002e"
	bVal, err = hex.DecodeString(val)
	assert.Nil(t, err)
	rst = DecodePacVal(bVal, PacExpBitWidth)
	assert.Equal(t, "100000000000000", rst)
}

func TestDecimalAddr2Hex(t *testing.T) {
	type args struct {
		addr string
	}
	var tests = []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{"1",
			args{
				addr: "4670991539099926443578259456495483705028406529",
			},
			"00d1745a6ad93272201680cd18f2eb4cd6366d01",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := DecimalAddr2Hex(tt.args.addr)
			assert.Equalf(t, tt.want, got, "DecimalAddr2Hex(%v)", tt.args.addr)
			assert.Equalf(t, tt.want1, got1, "DecimalAddr2Hex(%v)", tt.args.addr)
		})
	}
}

func TestHexAddr2Decimal(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{"1",
			args{
				addr: "0x00d1745a6ad93272201680cd18f2eb4cd6366d01",
			},
			"4670991539099926443578259456495483705028406529",
			true,
		},
		{"2",
			args{
				addr: "d1745a6ad93272201680cd18f2eb4cd6366d01",
			},
			"4670991539099926443578259456495483705028406529",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := HexAddr2Decimal(tt.args.addr)
			assert.Equalf(t, tt.want, got, "HexAddr2Decimal(%v)", tt.args.addr)
			assert.Equalf(t, tt.want1, got1, "HexAddr2Decimal(%v)", tt.args.addr)
		})
	}
}
