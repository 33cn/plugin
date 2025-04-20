package types

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"strconv"
	"strings"
)

// ToString encode to string
func (o *OutPoint) ToString() string {
	if o == nil {
		return "nil-OutPoint"
	}
	return FormatUtxo(o.Hash, o.Index)
}

// FormatUtxo format utxo as string
func FormatUtxo(hash string, index uint32) string {

	buf := make([]byte, 2*chainhash.HashSize+1, 2*chainhash.HashSize+1+10)
	copy(buf, hash)
	buf[2*chainhash.HashSize] = ':'
	buf = strconv.AppendUint(buf, uint64(index), 10)
	return string(buf)
}

// FromString decode from string
func (o *OutPoint) FromString(s string) error {

	strs := strings.Split(s, ":")
	if len(strs) != 2 {
		return fmt.Errorf("invalid outpoint: %s", s)
	}

	o.Hash = strs[0]
	v, err := strconv.ParseInt(strs[1], 10, 32)
	if err != nil {
		return err
	}
	o.Index = uint32(v)
	return nil
}

// NewOutPointFromString new out point
func NewOutPointFromString(s string) (*OutPoint, error) {

	o := &OutPoint{}
	strs := strings.Split(s, ":")
	if len(strs) != 2 {
		return nil, fmt.Errorf("invalid outpoint: %s", s)
	}
	o.Hash = strs[0]
	v, err := strconv.ParseInt(strs[1], 10, 32)
	if err != nil {
		return nil, err
	}
	o.Index = uint32(v)
	return o, nil
}

// IsUtxoAddress check if utxo address
func IsUtxoAddress(addr string) bool {

	if strings.Contains(addr, ":") && len(addr) > 2*chainhash.HashSize {
		return true
	}
	return false
}
