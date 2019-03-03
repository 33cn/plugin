package types

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	pb "github.com/33cn/chain33/types"
	// pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

// sortRounds is rounds of generate rands
const sortRounds = 3

// persent of allw online
const onlinePersentOfAllW = 0.8

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
func GenRands(allw, w int, priv crypto.PrivKey, blockHeight int64, blockHash []byte) []*Pos33Rands {
	// 本轮难度：委员会票数 / (总票数 * 在线率)
	diff := Pos33CommitteeSize / (float64(allw) * onlinePersentOfAllW)

	sorted := false // 是否有符合的票数
	rss := make([]*Pos33Rands, sortRounds)
	// 每张票计算sortRounds轮
	for i := 0; i < sortRounds; i++ {
		var rs Pos33Rands
		for j := 0; j < w; j++ {
			// 基于blockHash，blockHeight，voute_i, round_i计算hash
			data := []byte(string(blockHash) + fmt.Sprintf("%d%d%d", i, j, blockHeight))
			hash := crypto.Sha256(data)
			// 签名，为了可以验证
			sig := priv.Sign(hash)
			// 最终的抽签Hash
			rh := crypto.Sha256(sig.Bytes())

			// 转为big.Float计算，比较难度diff
			y := big.NewInt(0).SetBytes(rh)
			z := big.NewFloat(0).SetInt(y)
			if z.Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
				continue
			}
			// 符合，表示抽中了
			signature := &pb.Signature{Ty: pb.ED25519, Pubkey: priv.PubKey().Bytes(), Signature: sig.Bytes()}
			rs.Rands = append(rs.Rands, &Pos33Rand{RandHash: rh, Index: uint32(j), Sig: signature})
			rss[i] = &rs
			sorted = true
		}
	}

	if !sorted {
		return nil
	}

	return rss
}

// CheckRands 检验抽签hash，allw是当前总票数, w是抵押的票数
func CheckRands(addr string, allw, w int, rss []*Pos33Rands, blockHeight int64, blockHash []byte) error {
	diff := Pos33CommitteeSize / (float64(allw) * onlinePersentOfAllW)
	if len(rss) != sortRounds {
		return fmt.Errorf("%s len(rss)==%d", addr, len(rss))
	}
	for i := 0; i < sortRounds; i++ {
		rs := rss[i]
		if rs == nil {
			continue
		}
		if len(rs.Rands) == 0 {
			continue
		}
		// 抽中的票数不能大于抵押票数
		if len(rs.Rands) > w {
			return fmt.Errorf("%s len(rands)>w", addr)
		}

		rsm := make(map[string]bool)
		for j, r := range rs.Rands {
			if int(r.Index) > w {
				return fmt.Errorf("%s Round >= RANDAROUNDS, w=%d", addr, r.Index)
			}
			if _, ok := rsm[string(r.RandHash)]; ok {
				return fmt.Errorf("rand hash repeated")
			}
			rsm[string(r.RandHash)] = true

			data := []byte(string(blockHash) + fmt.Sprintf("%d%d%d", i, j, blockHeight))
			hash := crypto.Sha256(data)
			if !pb.CheckSign(hash, "pos33", r.Sig) {
				return fmt.Errorf("%s signature error", addr)
			}
			if addr != address.PubKeyToAddress(r.Sig.Pubkey).String() {
				return fmt.Errorf("%s signature error", addr)
			}

			if string(r.RandHash) != string(crypto.Sha256(r.Sig.Signature)) {
				return fmt.Errorf("%s rand hash error", addr)
			}

			y := big.NewInt(0).SetBytes(r.RandHash)
			z := big.NewFloat(0).SetInt(y)
			if z.Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
				return fmt.Errorf("%s diff error", addr)
			}
		}
	}
	return nil
}

// Sortition 统计每个action的投票，计算出共识委员会选票
func Sortition(acts []*Pos33ElecteAction) *Pos33Rands {
	// 方便统计，将action的票放到rss 中
	rss := make([]*Pos33Rands, sortRounds)
	for _, a := range acts {
		if len(a.Rands) == 0 {
			continue
		}
		for i, rs := range a.Rands {
			if rss[i] == nil {
				rss[i] = new(Pos33Rands)
			}
			if rs == nil {
				continue
			}
			// 累加sortRounds的票
			rss[i].Rands = append(rss[i].Rands, rs.Rands...)
		}
	}

	// 如果rss[0]中的票<min, 继续累加rss[1]的，以此类推
	min := Pos33MinCommittee
	rs := new(Pos33Rands)
	k := 0
	for i := 0; i < sortRounds; i++ {
		if rss[i] == nil {
			continue
		}
		if len(rs.Rands)+len(rss[i].Rands) >= min {
			k = i // 第i轮票数已经够了
			break
		}
		rs.Rands = append(rs.Rands, rss[i].Rands...)
	}

	if len(rs.Rands) < min {
		panic("can't go here")
	}

	// 如果k轮的总票数超出了max, 对第k轮票数排序，截取
	max := Pos33MaxCommittee
	if len(rs.Rands)+len(rss[k].Rands) > max {
		sort.Sort(rss[k])
		rss[k].Rands = rss[k].Rands[:max-len(rs.Rands)]
	}
	rs.Rands = append(rs.Rands, rss[k].Rands...)

	// 对最终符合的票数排序, 即委员会
	if k > 0 { // k == 0 时上面已经排过了
		sort.Sort(rs)
	}
	return rs
}
