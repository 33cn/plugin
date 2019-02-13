package types

import (
	"encoding/hex"
	"encoding/json"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	// ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

// const var
const (
	Pos33AllWeight      = "POS33-AllWeight"
	Pos33Weight         = "POS33-Weight-"
	Pos33DelegatePrefix = "POS33-Delegate:"
	Pos33Miner          = types.Coin * 10000
	Pos33BlockReword    = types.Coin * 15
	Pos33VoteReword     = types.Coin / 2
	Pos33MaxCommittee   = 10
	Pos33DepositPeriod  = 40320
	Pos33FundKeyAddr    = ""
)

/*
type IPos33Msg interface {
	Verify() bool
}

func Pos33Verify(m IPos33Msg) bool {
	return m.Verify()
}

func (m *Pos33Vote) Verify() bool {
	pm := &Pos33Msg{Data: Encode(&Pos33Commit{m.BlockHash}), Ty: Pos33Msg_COMMIT, Height: m.BlockHeight, Sig: m.BlockSig}
	return pm.Verify()
}

func (m *Pos33Msg) Verify() bool {
	sig := m.Sig
	m.Sig = nil
	defer func() {
		m.Sig = sig
	}()
	return CheckSign(Encode(m), "pos33", sig)
}

func (m *Pos33Msg) Sign(priv crypto.PrivKey) {
	m.Sig = nil
	s := priv.Sign(Encode(m))
	m.Sig = &Signature{Ty: ED25519, Pubkey: priv.PubKey().Bytes(), Signature: s.Bytes()}
}
*/

// Verify is verify vote msg
func (v *Pos33Vote) Verify() bool {
	s := v.Sig
	v.Sig = nil
	b := crypto.Sha256(types.Encode(v))
	v.Sig = s
	return types.CheckSign(b, "", s)
}

// Sign is sign vote msg
func (v *Pos33Vote) Sign(priv crypto.PrivKey) {
	v.Sig = nil
	b := crypto.Sha256(types.Encode(v))
	sig := priv.Sign(b)
	v.Sig = &types.Signature{Ty: types.ED25519, Pubkey: priv.PubKey().Bytes(), Signature: sig.Bytes()}
}

// // Verify is verify block
// func (m *types.Block) Verify() bool {
// 	return m.CheckSign()
// }

// // Sign is sign block
// func (m *types.Block) Sign(priv crypto.PrivKey) {
// 	s := priv.Sign(m.Hash())
// 	m.Signature = &Signature{Ty: ED25519, Pubkey: priv.PubKey().Bytes(), Signature: s.Bytes()}
// }

// ToString is rands to string
func (m *Pos33Rands) ToString() string {
	s := ""
	for _, r := range m.Rands {
		s += hex.EncodeToString(r.RandHash) + " "
	}
	return s
}

// // ToString is block to string
// func (b *Block) ToString() string {
// 	d, err := json.MarshalIndent(b, "", "  ")
// 	if err != nil {
// 		return err.Error()
// 	}
// 	return string(d)
// }

// ToString is reword to string
func (act *Pos33RewordAction) ToString() string {
	b, err := json.MarshalIndent(act, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func (m Pos33Rands) Len() int { return len(m.Rands) }
func (m Pos33Rands) Less(i, j int) bool {
	return string(m.Rands[i].RandHash) < string(m.Rands[j].RandHash)
}
func (m Pos33Rands) Swap(i, j int) { m.Rands[i], m.Rands[j] = m.Rands[j], m.Rands[i] }
