package types

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	pb "github.com/33cn/chain33/types"
	// pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

// sortRounds is rounds of generate rands
const sortRounds = 1

// persent of allw online
const onlinePersentOfAllW = 1. - 0.33

var max = big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256), nil)
var fmax = big.NewFloat(0).SetInt(max) // 2^^256

// 算法依据：
// 1. 通过签名，然后hash，得出的Hash值是在[0，max]的范围内均匀分布并且随机的, 那么Hash/max实在[1/max, 1]之间均匀分布的
// 2. 那么从N个选票中抽出M个选票，等价于计算N次Hash, 并且Hash/max < M/N
// 3. 由于Hash的随机性，可能某次的M太小或者过大，所以需要加以限定，过大则保留部分
// 4. 如果M太小，需要重新抽签，但是抽签是打包到交易里的，不符合节点利益。
// 5. 为了避免M过小，那么每次进行R轮计算，这样M大约为之前的R倍。 我们只要累加前几轮的M，大于给定最小值。
// 6. 签名是为了校验, 必须是自己私钥生成的
// 7. 最后对Hashs排序，作为委员会打包顺序的依据

// GenRands 计算抽签hash，allw是当前总票数, w是自己抵押的票数
func GenRands(allw, w int, priv crypto.PrivKey, blockHeight int64, blockHash []byte, step int) (*Pos33Rands, *pb.Signature) {
	// 本轮难度：委员会票数 / (总票数 * 在线率)
	size := Pos33VerifierSize
	if step == 0 {
		size = Pos33ProposerSize
	}
	diff := float64(size) / (float64(allw) * onlinePersentOfAllW)

	data := []byte(string(blockHash) + fmt.Sprintf(":%d%d", blockHeight, step))
	hash := crypto.Sha256(data)
	// 签名，为了可以验证
	sig := priv.Sign(hash)
	signature := &pb.Signature{Ty: pb.ED25519, Pubkey: priv.PubKey().Bytes(), Signature: sig.Bytes()}
	addr := address.PubKeyToAddress(signature.Pubkey).String()

	var minHash []byte
	var pos int

	var rs Pos33Rands
	for j := 0; j < w; j++ {
		// 基于blockHash，blockHeight，voute_i, round_i计算hash
		rdata := []byte(string(sig.Bytes()) + fmt.Sprintf(":%s:%d", addr, j))
		rhash := crypto.Sha256(rdata)

		// 转为big.Float计算，比较难度diff
		y := big.NewInt(0).SetBytes(rhash)
		z := big.NewFloat(0).SetInt(y)
		if z.Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
			continue
		}

		if minHash == nil {
			minHash = rhash
		}
		if string(minHash) > string(rhash) {
			minHash = rhash
			pos = len(rs.Rands)
		}
		// 符合，表示抽中了
		rs.Rands = append(rs.Rands, &Pos33Rand{Hash: rhash, Index: uint32(j), Addr: addr})
	}

	if len(rs.Rands) == 0 {
		return nil, nil
	}
	rs.Rands[0], rs.Rands[pos] = rs.Rands[pos], rs.Rands[0]
	if step == 0 {
		rs.Rands = rs.Rands[:1]
	} else {
		rs.Rands = rs.Rands[:min(len(rs.Rands), Pos33VerifierSize)]
	}

	return &rs, signature
}

// CheckRands 检验抽签hash，allw是当前总票数, w是抵押的票数
func CheckRands(addr string, allw, w int, rs *Pos33Rands, blockHeight int64, blockHash []byte, sig *pb.Signature, step int) error {
	size := Pos33VerifierSize
	if step == 0 {
		size = Pos33ProposerSize
	}
	diff := float64(size) / (float64(allw) * onlinePersentOfAllW)

	data := []byte(string(blockHash) + fmt.Sprintf(":%d%d", blockHeight, step))
	hash := crypto.Sha256(data)

	if !pb.CheckSign(hash, "pos33", sig) {
		return fmt.Errorf("%s signature error", addr)
	}
	if addr != address.PubKeyToAddress(sig.Pubkey).String() {
		return fmt.Errorf("%s NOT match signature", addr)
	}

	if len(rs.Rands) == 0 {
		return fmt.Errorf("sortition is nil")
	}
	// 抽中的票数不能大于抵押票数
	if len(rs.Rands) > w {
		return fmt.Errorf("%s len(rands)>w", addr)
	}

	mp := make(map[int]bool)

	for _, r := range rs.Rands {
		if int(r.Index) > w {
			return fmt.Errorf("%s Round >= RANDAROUNDS, w=%d", addr, r.Index)
		}
		if addr != r.Addr {
			return errors.New("rand address error")
		}

		rdata := []byte(string(sig.Signature) + fmt.Sprintf(":%s:%d", r.Addr, r.Index))
		rhash := crypto.Sha256(rdata)
		if string(r.Hash) != string(rhash) {
			return fmt.Errorf("%s rand hash error", addr)
		}

		y := big.NewInt(0).SetBytes(r.Hash)
		z := big.NewFloat(0).SetInt(y)
		if z.Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
			return fmt.Errorf("%s diff error", addr)
		}
		if _, ok := mp[int(r.Index)]; ok {
			return fmt.Errorf("%s reuse %d index sortition", addr, r.Index)
		}
		mp[int(r.Index)] = true
	}
	return nil
}

// Sortition 统计每个action的投票，计算出共识委员会选票
func Sortition(msgs []*Pos33ElectMsg, step int) *Pos33Rands {
	if len(msgs) == 0 {
		return nil
	}

	var rs Pos33Rands
	for _, a := range msgs {
		for _, r := range a.Rands.Rands {
			rs.Rands = append(rs.Rands, r)
		}
	}

	sort.Sort(&rs)
	if step == 0 {
		rs.Rands = rs.Rands[:min(len(rs.Rands), Pos33ProposerSize)]
	} else {
		rs.Rands = rs.Rands[:min(len(rs.Rands), Pos33VerifierSize)]
	}
	return &rs
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
