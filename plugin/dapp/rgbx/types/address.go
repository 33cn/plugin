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
		return "<nil>"
	}
	hash, _ := chainhash.NewHash(o.Hash)
	return FormatUtxo(hash.String(), o.Index)
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
	b, err := chainhash.NewHashFromStr(strs[0])
	if err != nil {
		return err
	}
	o.Hash = b.CloneBytes()

	v, err := strconv.ParseInt(strs[1], 10, 32)
	if err != nil {
		return err
	}
	o.Index = uint32(v)
	return nil
}

func (a *AssetAccount) Address() string {
	if a.GetUtxo() != nil {
		return a.GetUtxo().ToString()
	}
	return a.GetAddress()
}
