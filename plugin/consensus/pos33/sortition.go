package pos33

import (
	"fmt"
	"math/big"

	"github.com/33cn/chain33/common/crypto"
	pb "github.com/33cn/chain33/types"
)

// RANDAROUNDS is rounds of generate rands
//const RANDAROUNDS = 3

// Diff is difficulty adjustment factor
const Diff = 1.123

var max = big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256), nil)
var fmax = big.NewFloat(0).SetInt(max) // 2^^256

func genRands(seed []byte, allw, w int, priv crypto.PrivKey, height int64) *ty.Pos33Rands {
	p := Diff * ty.Pos33MaxCommittee / float64(allw)
	pub := string(priv.PubKey().Bytes())

	// var rm ty.Pos33Rands
	// for k := 0; k < RANDAROUNDS; k++ {
	// if w > ty.Pos33MaxCommittee / 2 {
	// 	w = ty.Pos33MaxCommittee / 2
	// }
	var rs ty.Pos33Rands
	for j := 0; j < w; j++ {
		data := []byte(pub + string(seed) + fmt.Sprintf("%d%d", j, height))
		hash := crypto.Sha256(data)
		sig := priv.Sign(hash)
		rh := crypto.Sha256(sig.Bytes())

		y := big.NewInt(0).SetBytes(rh)
		z := big.NewFloat(0).SetInt(y)
		if z.Quo(z, fmax).Cmp(big.NewFloat(p)) > 0 {
			continue
		}
		// z := y.Mod(y, big.NewInt(int64(allw)))
		// if z.Int64() >= int64(ty.Pos33MaxCommittee) {
		// 	continue
		// }
		rs.Rands = append(rs.Rands, &ty.Pos33Rand{RandHash: rh, Weight: uint32(j), Sig: sig.Bytes()})
	}

	if len(rs.Rands) == 0 {
		return nil
	}
	// sign the rands
	rs.Height = height
	rs.Pub = pub
	return &rs
}

func checkRands(pub string, allw, w int, seed []byte, m *ty.Pos33Rands, height int64) error {
	if len(m.Rands) == 0 {
		return fmt.Errorf("%s len(rands)==0", hexString([]byte(pub)))
	}
	// if w > ty.Pos33MaxCommittee / 2 {
	// 	w = ty.Pos33MaxCommittee / 2
	// }
	if len(m.Rands) > w {
		return fmt.Errorf("%s len(rands)>w", hexString([]byte(pub)))
	}
	p := Diff * ty.Pos33MaxCommittee / float64(allw)

	for _, r := range m.Rands {
		if int(r.Weight) > w {
			return fmt.Errorf("%s Round >= RANDAROUNDS, w=%d", hexString([]byte(pub)), r.Weight)
		}

		data := []byte(pub + string(seed) + fmt.Sprintf("%d%d", r.Weight, height))
		sig := &pb.Signature{Ty: pb.ED25519, Pubkey: []byte(pub), Signature: r.Sig}
		hash := crypto.Sha256(data)
		if !pb.CheckSign(hash, "pos33", sig) {
			return fmt.Errorf("%s signature error", hexString([]byte(pub)))
		}
		if string(r.RandHash) != string(crypto.Sha256(r.Sig)) {
			return fmt.Errorf("%s rand hash error", hexString([]byte(pub)))
		}

		y := big.NewInt(0).SetBytes(r.RandHash)
		// z := y.Mod(y, big.NewInt(int64(allw)))
		// if z.Int64() > int64(ty.Pos33MaxCommittee) {
		// 	return fmt.Errorf("%s diff error", hexString([]byte(pub)))
		// }
		z := big.NewFloat(0).SetInt(y)
		if z.Quo(z, fmax).Cmp(big.NewFloat(p)) > 0 {
			return fmt.Errorf("%s diff error", hexString([]byte(pub)))
		}
	}
	return nil
}

/*
func sortition(rands []*ty.Pos33Rands, allw int, height int64) *committee {
	max := 0
	var committee map[string]*ty.Pos33Rands
	for i := 0; i < RANDAROUNDS; i++ {
		rm, ok := rands[i]
		if !ok {
			continue
		}
		l := 0
		for _, rs := range rm {
			l += len(rs.Rands)
		}
		if l > max {
			max = l
			committee = rm
		}
	}
	if len(committee) == 0 {
		return nil
	}

	low := allw * 2 / 3
	if low > ty.Pos33Members {
		low = ty.Pos33Members
	}

	maker := ""
	min := "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	for pub, rs := range committee {
		for _, r := range rs.Rands {
			sr := hexString(r.RandHash)
			if sr < min {
				min = sr
				maker = pub
			}
		}
	}
	return &committee{
		committee: committee,
		minHash:   min,
		bp:        maker,
		allw:      allw,
		sorted:    max > low,
		height:    height,
	}
}
*/
