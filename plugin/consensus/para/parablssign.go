// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"bytes"
	"math/big"
	"sort"

	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/phoreproject/bls/g2pubs"
	"github.com/pkg/errors"
)

const (
	maxRcvTxCount = 100 //max 100 nodes, 1 height tx or 1 txgroup per node
)

type blsClient struct {
	paraClient     *client
	selfID         string
	blsPriKey      *g2pubs.SecretKey
	blsPubKey      *g2pubs.PublicKey
	peers          map[string]bool
	peersBlsPubKey map[string]*g2pubs.PublicKey
	txsBuff        map[int64]*pt.ParaBlsSignSumDetails
	rcvCommitTxCh  chan []*pt.ParacrossCommitAction
	quit           chan struct{}
}

func newBlsClient(para *client, cfg *subConfig) *blsClient {
	b := &blsClient{paraClient: para}
	b.selfID = cfg.AuthAccount
	b.peers = make(map[string]bool)
	b.peersBlsPubKey = make(map[string]*g2pubs.PublicKey)
	b.txsBuff = make(map[int64]*pt.ParaBlsSignSumDetails)
	b.rcvCommitTxCh = make(chan []*pt.ParacrossCommitAction, maxRcvTxCount)
	b.quit = make(chan struct{})

	return b
}

//1. 要等到达成共识了才发送，不然处理未达成共识的各种场景会比较复杂，而且浪费手续费
func (b *blsClient) procRcvSignTxs() {
	defer b.paraClient.wg.Done()
	if len(b.selfID) <= 0 {
		return
	}
	p2pTimer := time.NewTimer(time.Minute)
out:
	for {
		select {
		case commits := <-b.rcvCommitTxCh:
			collectSigns(b.txsBuff, commits)
			nodes := b.paraClient.commitMsgClient.authNodes.Load().([]string)
			if !isMostCommitDone(len(nodes), b.txsBuff) {
				continue
			}
			//清空txsBuff，重新收集
			txsBuff := b.txsBuff
			b.txsBuff = make(map[int64]*pt.ParaBlsSignSumDetails)

			//自己是Coordinator,则聚合交易
			if b.paraClient.bullyCli.IsSelfCoordinator() {
				dones := filterDoneCommits(len(nodes), txsBuff)
				if len(dones) > 0 {
					continue
				}
				acts, err := b.transferCommit2Action(dones)
				if err != nil {
					continue
				}
				b.paraClient.commitMsgClient.sendCommitActions(acts)
			}
		case <-p2pTimer.C:
			if len(b.selfID) > 0 {
				//tle := cfg.GetTitle()
				plog.Info("send p2p topic------------------------------")
				b.paraClient.subP2PTopic()
				plog.Info("rcv p2p topic-------------------------------")

			}
		case <-b.quit:
			break out
		}
	}

}

func (b *blsClient) rcvCommitTx(tx *types.Transaction) error {
	if !tx.CheckSign() {
		return types.ErrSign
	}

	if !b.paraClient.commitMsgClient.isValidNode(tx.From()) {
		b.updatePeers(tx.From(), false)
		return pt.ErrParaNodeAddrNotExisted
	}
	b.updatePeers(tx.From(), true)

	txs := []*types.Transaction{tx}
	if count := tx.GetGroupCount(); count > 0 {
		group, err := tx.GetTxGroup()
		if err != nil {
			return err
		}
		txs = group.Txs
	}

	commits, err := b.getCommitInfo(txs)
	if err != nil {
		return err
	}
	b.rcvCommitTxCh <- commits
	return nil

}

func (b *blsClient) getCommitInfo(txs []*types.Transaction) ([]*pt.ParacrossCommitAction, error) {
	var commits []*pt.ParacrossCommitAction
	for _, tx := range txs {
		var act pt.ParacrossAction
		err := types.Decode(tx.Payload, &act)
		if err != nil {
			return nil, errors.Wrap(err, "decode act")
		}
		if act.Ty != pt.ParacrossActionCommit {
			return nil, types.ErrInvalidParam
		}
		commit := act.GetCommit()
		if tx.From() != commit.Bls.Addrs[0] {
			return nil, types.ErrFromAddr
		}
		err = b.verifyBlsSign(tx.From(), commit)
		if err != nil {
			return nil, pt.ErrBlsSignVerify
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func collectSigns(txsBuff map[int64]*pt.ParaBlsSignSumDetails, commits []*pt.ParacrossCommitAction) {
	for _, cmt := range commits {
		if _, ok := txsBuff[cmt.Status.Height]; !ok {
			txsBuff[cmt.Status.Height] = &pt.ParaBlsSignSumDetails{Height: cmt.Status.Height}
		}
		a := txsBuff[cmt.Status.Height]
		for i, v := range a.Addrs {
			//节点更新交易参数的场景
			if v == cmt.Bls.Addrs[0] {
				a.Msgs[i] = types.Encode(cmt.Status)
				a.Signs[i] = cmt.Bls.Sign
				continue
			}
		}
		a.Addrs = append(a.Addrs, cmt.Bls.Addrs[0])
		a.Msgs = append(a.Msgs, types.Encode(cmt.Status))
		a.Signs = append(a.Signs, cmt.Bls.Sign)
	}
}

func isMostCommitDone(peers int, txsBuff map[int64]*pt.ParaBlsSignSumDetails) bool {
	for i, v := range txsBuff {
		most, _ := getMostCommit(v.Msgs)
		if isCommitDone(peers, most) {
			plog.Info("blssign.isMostCommitDone", "height", i, "most", most, "peers", peers)
			return true
		}
	}
	return false
}

func filterDoneCommits(peers int, txs map[int64]*pt.ParaBlsSignSumDetails) []*pt.ParaBlsSignSumDetails {
	var seq []int64
	for i, v := range txs {
		most, hash := getMostCommit(v.Msgs)
		if !isCommitDone(peers, most) {
			plog.Info("blssign.filterDoneCommits not commit done", "height", i)
			delete(txs, i)
			continue
		}
		seq = append(seq, i)

		//只保留相同的commits
		a := &pt.ParaBlsSignSumDetails{Msgs: [][]byte{[]byte(hash)}}
		for j, m := range v.Msgs {
			if bytes.Equal([]byte(hash), m) {
				a.Addrs = append(a.Addrs, v.Addrs[j])
				a.Signs = append(a.Signs, v.Signs[j])
			}
		}
		txs[i] = a
	}

	if len(seq) <= 0 {
		plog.Info("blssign.filterDoneCommits nil")
		return nil
	}

	sort.Slice(seq, func(i, j int) bool { return seq[i] < seq[j] })
	plog.Info("blssign.filterDoneCommits", "seq", seq)
	var signs []*pt.ParaBlsSignSumDetails
	//共识高度要连续，不连续则退出
	lastSeq := seq[0] - 1
	for _, h := range seq {
		if lastSeq+1 != h {
			return signs
		}
		signs = append(signs, txs[h])
		lastSeq = h
	}
	return signs

}

func (b *blsClient) transferCommit2Action(commits []*pt.ParaBlsSignSumDetails) ([]*pt.ParacrossCommitAction, error) {
	var notify []*pt.ParacrossCommitAction

	for _, v := range commits {
		a := &pt.ParacrossCommitAction{}
		s := &pt.ParacrossNodeStatus{}
		types.Decode(v.Msgs[0], s)
		a.Status = s

		sign, err := b.aggregateSigns(v.Signs)
		if err != nil {
			return nil, err
		}
		signData := sign.Serialize()
		copy(a.Bls.Sign, signData[:])
		nodes := b.paraClient.commitMsgClient.authNodes.Load().([]string)
		bits, remains := setAddrsBitMap(nodes, v.Addrs)
		if len(remains) > 0 {
			plog.Info("bls.signDoneCommits", "remains", remains)
		}
		a.Bls.AddrsMap = bits
		notify = append(notify, a)
	}
	return notify, nil
}

func (b *blsClient) aggregateSigns(signs [][]byte) (*g2pubs.Signature, error) {
	var signatures []*g2pubs.Signature
	for _, data := range signs {
		var s [48]byte
		copy(s[:], data)
		signKey, err := g2pubs.DeserializeSignature(s)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, signKey)
	}

	return g2pubs.AggregateSignatures(signatures), nil
}

func (b *blsClient) updatePeers(peer string, add bool) {
	if _, ok := b.peers[peer]; ok {
		if !add {
			delete(b.peers, peer)
		}
		return
	}
	if add {
		b.peers[peer] = true
	}

}

func (b *blsClient) setBlsPriKey(secpPrkKey []byte) {
	b.blsPriKey = getBlsPriKey(secpPrkKey)
	b.blsPubKey = g2pubs.PrivToPub(b.blsPriKey)
	serial := b.blsPubKey.Serialize()
	plog.Info("para commit get pub bls", "pubkey", common.ToHex(serial[:]))
}

//to repeat get prikey's hash until in range of bls's private key
func getBlsPriKey(key []byte) *g2pubs.SecretKey {
	var newKey [common.Sha256Len]byte
	copy(newKey[:], key[:])
	for {
		plog.Info("para commit getBlsPriKey", "keys", common.ToHex(newKey[:]))
		secret := g2pubs.DeserializeSecretKey(newKey)
		if nil != secret.GetFRElement() {
			serial := secret.Serialize()
			plog.Info("para commit getBlsPriKey", "final keys", common.ToHex(serial[:]), "string", secret.String())
			return secret
		}
		copy(newKey[:], common.Sha256(newKey[:]))
	}

}

func (b *blsClient) blsSign(commits []*pt.ParacrossCommitAction) error {
	for _, cmt := range commits {
		data := types.Encode(cmt.Status)
		plog.Debug("blsign msg", "data", common.ToHex(data), "height", cmt.Status.Height)
		sign := g2pubs.Sign(data, b.blsPriKey).Serialize()
		cmt.Bls = &pt.ParacrossCommitBlsInfo{Sign: sign[:], Addrs: []string{b.selfID}}
	}
	return nil
}

//设置nodes范围内的bitmap，如果addrs在node不存在，也不设置,返回未命中的addrs
func setAddrsBitMap(nodes, addrs []string) ([]byte, map[string]bool) {
	rst := big.NewInt(0)
	addrsMap := make(map[string]bool)
	for _, n := range addrs {
		addrsMap[n] = true
	}

	for i, a := range nodes {
		if _, exist := addrsMap[a]; exist {
			rst.SetBit(rst, i, 1)
			delete(addrsMap, a)
		}
	}
	return rst.Bytes(), addrsMap
}

func getMostCommit(commits [][]byte) (int, string) {
	stats := make(map[string]int)
	n := len(commits)
	for i := 0; i < n; i++ {
		if _, ok := stats[string(commits[i])]; ok {
			stats[string(commits[i])]++
		} else {
			stats[string(commits[i])] = 1
		}
	}
	most := -1
	var hash string
	for k, v := range stats {
		if v > most {
			most = v
			hash = k
		}
	}
	return most, hash
}

func isCommitDone(nodes, mostSame int) bool {
	return 3*mostSame > 2*nodes
}

func (b *blsClient) getBlsPubKey(addr string) (*g2pubs.PublicKey, error) {
	//先从缓存中获取
	if v, ok := b.peersBlsPubKey[addr]; ok {
		return v, nil
	}

	//缓存没有，则从statedb获取
	cfg := b.paraClient.GetAPI().GetConfig()
	ret, err := b.paraClient.GetAPI().QueryChain(&types.ChainExecutor{
		Driver:   "paracross",
		FuncName: "GetNodeAddrInfo",
		Param:    types.Encode(&pt.ReqParacrossNodeInfo{Title: cfg.GetTitle(), Addr: addr}),
	})
	if err != nil {
		plog.Error("commitmsg.GetNodeAddrInfo ", "err", err.Error())
		return nil, err
	}
	resp, ok := ret.(*pt.ParaNodeAddrIdStatus)
	if !ok {
		plog.Error("commitmsg.getNodeGroupAddrs rsp nok")
		return nil, err
	}

	//pubKeys := make([]*g2pubs.PublicKey, 0)
	val, err := common.FromHex(resp.BlsPubKey)
	if err != nil {
		plog.Error("verifyBlsSign.fromhex", "p", addr)
		return nil, err
	}
	k := [96]byte{}
	copy(k[:], val)
	pubKey, err := g2pubs.DeserializePublicKey(k)
	if err != nil {
		plog.Error("verifyBlsSign.DeserializePublicKey", "key", addr)
		return nil, err
	}

	b.peersBlsPubKey[addr] = pubKey

	return pubKey, nil
}

func (b *blsClient) verifyBlsSign(addr string, commit *pt.ParacrossCommitAction) error {
	//1. 获取对应公钥
	pubKey, err := b.getBlsPubKey(addr)
	if err != nil {
		plog.Error("verifyBlsSign pub　key not exist", "addr", addr)
		return err
	}

	//2.　获取bls签名
	signkey := [48]byte{}
	copy(signkey[:], commit.Bls.Sign)
	sign, err := g2pubs.DeserializeSignature(signkey)
	if err != nil {
		plog.Error("verifyBlsSign.DeserializeSignature", "key", common.ToHex(commit.Bls.Sign))
		return err
	}

	//3. 获取签名前原始msg
	msg := types.Encode(commit.Status)

	if !g2pubs.Verify(msg, pubKey, sign) {
		plog.Error("paracross.Commit bls sign verify", "title", commit.Status.Title, "height", commit.Status.Height,
			"addrsMap", common.ToHex(commit.Bls.AddrsMap), "sign", common.ToHex(commit.Bls.Sign), "addr", addr)
		plog.Error("paracross.commit bls sign verify", "data", common.ToHex(msg), "height", commit.Status.Height)
		return pt.ErrBlsSignVerify
	}
	return nil
}

func (b *blsClient) showTxBuffInfo() *pt.ParaBlsSignSumInfo {
	var seq []int64
	var ret pt.ParaBlsSignSumInfo
	for k := range b.txsBuff {
		seq = append(seq, k)
	}
	sort.Slice(seq, func(i, j int) bool { return seq[i] < seq[j] })

	for _, h := range seq {
		ret.Info = append(ret.Info, b.txsBuff[h])
	}
	return &ret
}
