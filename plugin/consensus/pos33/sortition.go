package pos33

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"

	vrf "github.com/33cn/chain33/common/vrf/secp256k1"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
	secp256k1 "github.com/btcsuite/btcd/btcec"
)

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

func (client *Client) sort(seed []byte, height int64, round, step int) []*pt.Pos33SortitionMsg {
	// 本轮难度：委员会票数 / (总票数 * 在线率)
	size := pt.Pos33VoterSize
	if step == 0 {
		size = pt.Pos33ProposerSize
	}
	allw := client.allWeight(height)
	diff := float64(size) / (float64(allw) * onlinePersentOfAllW)

	var msgs []*pt.Pos33SortitionMsg
	var minHash []byte
	index := 0
	for tid, t := range client.ticketsMap {
		inputMsg := &pt.Pos33VrfInputMsg{Seed: seed, Height: height, Round: int32(round), Step: int32(step), TicketId: tid}
		priv := client.privmap[t.MinerAddress]
		privKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), priv.Bytes())
		vrfPriv := &vrf.PrivateKey{PrivateKey: (*ecdsa.PrivateKey)(privKey)}
		input := types.Encode(inputMsg)
		vrfHash, vrfProof := vrfPriv.Evaluate(input)
		hash := vrfHash[:]

		// 转为big.Float计算，比较难度diff
		y := big.NewInt(0).SetBytes(hash)
		z := big.NewFloat(0).SetInt(y)
		if z.Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
			continue
		}

		if minHash == nil {
			minHash = hash
		}
		if string(minHash) > string(hash) {
			minHash = hash
			index = len(msgs)
		}
		// 符合，表示抽中了

		m := &pt.Pos33SortitionMsg{Hash: hash, Proof: vrfProof[:], Input: inputMsg, Pubkey: priv.PubKey().Bytes()}
		msgs = append(msgs, m)
	}

	if len(msgs) == 0 {
		return nil
	}
	if step == 0 {
		return []*pt.Pos33SortitionMsg{msgs[index]}
	}
	sort.Sort(pt.Sorts(msgs))
	return msgs
}

func vrfVerify(pub []byte, input []byte, proof []byte, hash []byte) error {
	pubKey, err := secp256k1.ParsePubKey(pub, secp256k1.S256())
	if err != nil {
		plog.Error("vrfVerify", "err", err)
		return pt.ErrVrfVerify
	}
	vrfPub := &vrf.PublicKey{PublicKey: (*ecdsa.PublicKey)(pubKey)}
	vrfHash, err := vrfPub.ProofToHash(input, proof)
	if err != nil {
		plog.Error("vrfVerify", "err", err)
		return pt.ErrVrfVerify
	}
	plog.Debug("vrf verify", "ProofToHash", fmt.Sprintf("(%x, %x): %x", input, proof, vrfHash), "hash", hex.EncodeToString(hash))
	if !bytes.Equal(vrfHash[:], hash) {
		plog.Error("vrfVerify", "err", fmt.Errorf("invalid VRF hash"))
		return pt.ErrVrfVerify
	}
	return nil
}

func (client *Client) verifySort(height int64, step int, seed []byte, m *pt.Pos33SortitionMsg) error {
	// 本轮难度：委员会票数 / (总票数 * 在线率)
	size := pt.Pos33VoterSize
	if step == 0 {
		size = pt.Pos33ProposerSize
	}
	allw := client.allWeight(height)
	diff := float64(size) / (float64(allw) * onlinePersentOfAllW)

	resp, err := client.GetAPI().Query(pt.Pos33TicketX, "TicketInfos", &pt.Pos33TicketInfos{TicketIds: []string{m.Input.GetTicketId()}})
	if err != nil {
		return err
	}
	reply := resp.(*pt.ReplyPos33TicketList)

	ok := false
	for _, t := range reply.Tickets {
		if t.TicketId == m.Input.TicketId {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("ticketID error")
	}

	im := &pt.Pos33VrfInputMsg{TicketId: m.Input.TicketId, Seed: seed, Height: height, Round: m.Input.GetRound(), Step: int32(step)}
	input := types.Encode(im)
	err = vrfVerify(m.Pubkey, input, m.Proof, m.Hash)
	if err != nil {
		return err
	}

	y := big.NewInt(0).SetBytes(m.Hash)
	z := big.NewFloat(0).SetInt(y)
	if z.Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
		return fmt.Errorf("diff error")
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
