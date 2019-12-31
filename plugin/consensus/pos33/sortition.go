package pos33

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	vrf "github.com/33cn/chain33/common/vrf/secp256k1"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/pos33/types"
	secp256k1 "github.com/btcsuite/btcd/btcec"
)

const diffValue = 1.
const diffDelta = 0.03

var max = big.NewInt(0).Exp(big.NewInt(2), big.NewInt(256), nil)
var fmax = big.NewFloat(0).SetInt(max) // 2^^256

// 算法依据：
// 1. 通过签名，然后hash，得出的Hash值是在[0，max]的范围内均匀分布并且随机的, 那么Hash/max实在[1/max, 1]之间均匀分布的
// 2. 那么从N个选票中抽出M个选票，等价于计算N次Hash, 并且Hash/max < M/N

func calcuVrfHash(input *pt.VrfInput, priv crypto.PrivKey) ([]byte, []byte) {
	privKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), priv.Bytes())
	vrfPriv := &vrf.PrivateKey{PrivateKey: (*ecdsa.PrivateKey)(privKey)}
	in := types.Encode(input)
	vrfHash, vrfProof := vrfPriv.Evaluate(in)
	hash := vrfHash[:]
	return hash, vrfProof
}

func changeDiff(size, round int) int {
	if round <= 3 {
		return size
	}
	return size + round - 3
}

func (n *node) sort(seed []byte, height int64, round, step, allw int) []*pt.Pos33SortMsg {
	// 本轮难度：委员会票数 / (总票数 * 在线率)
	size := pt.Pos33VoterSize
	if step == 0 {
		size = pt.Pos33ProposerSize
	}
	//allw := client.allWeight(height)
	diff := float64(changeDiff(size, round)) / float64(allw)

	plog.Debug("sortition", "height", height, "round", round, "step", step, "seed", hexs(seed), "allw", allw)
	priv := n.getPriv("")
	input := &pt.VrfInput{Seed: seed, Height: height, Round: int32(round), Step: int32(step)}
	vrfHash, vrfProof := calcuVrfHash(input, priv)
	proof := &pt.HashProof{
		Input:    input,
		VrfHash:  vrfHash,
		VrfProof: vrfProof,
		Pubkey:   priv.PubKey().Bytes(),
	}

	var msgs []*pt.Pos33SortMsg
	var minHash []byte
	index := 0
	for tid := range n.getTicketsMap(height) {
		data := fmt.Sprintf("%x+%s", vrfHash, tid)
		hash := hash2([]byte(data))

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
		m := &pt.Pos33SortMsg{
			SortHash: &pt.SortHash{Hash: hash, Tid: tid},
			Proof:    proof,
		}
		msgs = append(msgs, m)
	}

	if len(msgs) == 0 {
		return nil
	}
	if step == 0 {
		return []*pt.Pos33SortMsg{msgs[index]}
	}
	sort.Sort(pt.Sorts(msgs))
	c := pt.Pos33VoterSize
	if len(msgs) > c {
		return msgs[:c]
	}
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

var errDiff = errors.New("diff error")

func (n *node) verifySort(height int64, step, allw int, seed []byte, m *pt.Pos33SortMsg) error {
	// 本轮难度：委员会票数 / (总票数 * 在线率)
	size := pt.Pos33VoterSize
	if step == 0 {
		size = pt.Pos33ProposerSize
	}
	if m == nil || m.Proof == nil || m.SortHash == nil || m.Proof.Input == nil {
		return fmt.Errorf("verifySort error: sort msg is nil")
	}

	round := m.Proof.Input.Round
	diff := float64(changeDiff(size, int(round))) / float64(allw)

	resp, err := n.GetAPI().Query(pt.Pos33TicketX, "Pos33TicketInfos", &pt.Pos33TicketInfos{TicketIds: []string{m.SortHash.Tid}})
	if err != nil {
		return err
	}
	reply := resp.(*pt.ReplyPos33TicketList)

	ok := false
	for _, t := range reply.Tickets {
		if t.TicketId == m.SortHash.Tid && getTicketHeight(t.TicketId) <= height {
			if t.MinerAddress == address.PubKeyToAddr(m.Proof.Pubkey) {
				ok = true
			}
			break
		}
	}
	if !ok {
		return fmt.Errorf("ticketID error")
	}

	input := &pt.VrfInput{Seed: seed, Height: height, Round: round, Step: int32(step)}
	in := types.Encode(input)
	err = vrfVerify(m.Proof.Pubkey, in, m.Proof.VrfProof, m.Proof.VrfHash)
	if err != nil {
		return err
	}
	data := fmt.Sprintf("%x+%s", m.Proof.VrfHash, m.SortHash.Tid)
	hash := hash2([]byte(data))
	if string(hash) != string(m.SortHash.Hash) {
		return fmt.Errorf("sort hash error")
	}

	y := big.NewInt(0).SetBytes(hash)
	z := big.NewFloat(0).SetInt(y)
	if z.Quo(z, fmax).Cmp(big.NewFloat(diff)) > 0 {
		return errDiff
	}

	return nil
}

func hash2(data []byte) []byte {
	return crypto.Sha256(crypto.Sha256(data))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
