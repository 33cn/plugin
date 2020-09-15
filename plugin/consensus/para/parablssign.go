// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"bytes"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"

	"github.com/pkg/errors"
)

const (
	maxRcvTxCount      = 100 //channel buffer, max 100 nodes, 1 height tx or 1 txgroup per node
	leaderSyncInt      = 15  //15s heartbeat sync interval
	defLeaderSwitchInt = 100 //每隔100个共识高度切换一次leader,大约6小时（按50个空块间隔计算）
	delaySubP2pTopic   = 10  //30s to sub p2p topic

	paraBlsSignTopic = "PARA-BLS-SIGN-TOPIC"
)

type blsClient struct {
	paraClient      *client
	selfID          string
	cryptoCli       crypto.Crypto
	blsPriKey       crypto.PrivKey
	blsPubKey       crypto.PubKey
	peersBlsPubKey  map[string]crypto.PubKey
	commitsPool     map[int64]*pt.ParaBlsSignSumDetails
	rcvCommitTxCh   chan []*pt.ParacrossCommitAction
	leaderOffset    int32
	leaderSwitchInt int32
	feedDog         uint32
	quit            chan struct{}
	mutex           sync.Mutex
}

func newBlsClient(para *client, cfg *subConfig) *blsClient {
	b := &blsClient{paraClient: para}
	b.selfID = cfg.AuthAccount
	cli, err := crypto.New("bls")
	if err != nil {
		panic("new bls crypto fail")
	}
	b.cryptoCli = cli
	b.peersBlsPubKey = make(map[string]crypto.PubKey)
	b.commitsPool = make(map[int64]*pt.ParaBlsSignSumDetails)
	b.rcvCommitTxCh = make(chan []*pt.ParacrossCommitAction, maxRcvTxCount)
	b.quit = make(chan struct{})
	b.leaderSwitchInt = defLeaderSwitchInt
	if cfg.BlsLeaderSwitchIntval > 0 {
		b.leaderSwitchInt = cfg.BlsLeaderSwitchIntval
	}

	return b
}

/*
1. 当前的leaderIndex和自己在nodegroup里面index一致，则自己是leader，负责发送共识交易
2. 当前leader负责每15s 发送一个喂狗消息，表明自己live
3. 每个node开启watchdog定时器，如果超时，则leaderIndex++, 寻找新的活的leader
4. node一旦收到新的喂狗消息，则把新消息的index更新为自己的leaderIndex， 如果和自己的leaderIndex冲突，则选大者
5. 每隔比如100个共识高度，就需要轮换leader，触发leaderIndex++，leader均衡轮换发送共识交易
*/
func (b *blsClient) procLeaderSync() {
	defer b.paraClient.wg.Done()
	if len(b.selfID) <= 0 {
		return
	}

	var feedDogTicker <-chan time.Time
	var watchDogTicker <-chan time.Time

	p2pTimer := time.After(delaySubP2pTopic * time.Second)
out:
	for {
		select {
		case <-feedDogTicker:
			//leader需要定期喂狗
			_, _, base, off, isLeader := b.getLeaderInfo()
			if isLeader {
				act := &pt.ParaP2PSubMsg{Ty: P2pSubLeaderSyncMsg}
				act.Value = &pt.ParaP2PSubMsg_SyncMsg{SyncMsg: &pt.LeaderSyncInfo{ID: b.selfID, BaseIdx: base, Offset: off}}
				err := b.paraClient.SendPubP2PMsg(paraBlsSignTopic, types.Encode(act))
				if err != nil {
					plog.Error("para.procLeaderSync feed dog", "err", err)
				}
				plog.Info("procLeaderSync feed dog", "id", b.selfID, "base", base, "off", off)
			}

		case <-watchDogTicker:
			//排除不在Nodegroup里面的Node
			if !b.isValidNodes(b.selfID) {
				plog.Info("procLeaderSync watchdog, not in nodegroup", "self", b.selfID)
				continue
			}
			//至少1分钟内要收到leader喂狗消息，否则认为leader挂了，index++
			if atomic.LoadUint32(&b.feedDog) == 0 {
				nodes, leader, _, off, _ := b.getLeaderInfo()
				if len(nodes) <= 0 {
					continue
				}
				atomic.StoreInt32(&b.leaderOffset, (off+1)%int32(len(nodes)))
				plog.Info("procLeaderSync watchdog", "fail node", nodes[leader], "newOffset", atomic.LoadInt32(&b.leaderOffset))
			}
			atomic.StoreUint32(&b.feedDog, 0)

		case <-p2pTimer:
			err := b.paraClient.SendSubP2PTopic(paraBlsSignTopic)
			if err != nil {
				plog.Error("procLeaderSync.SubP2PTopic", "err", err)
				p2pTimer = time.After(delaySubP2pTopic * time.Second)
				continue
			}
			feedDogTicker = time.NewTicker(leaderSyncInt * time.Second).C
			watchDogTicker = time.NewTicker(time.Minute).C
		case <-b.quit:
			break out
		}
	}
}

//处理leader sync tx, 需接受同步的数据，两个节点基本的共识高度相同, 两个共同leader需相同
func (b *blsClient) rcvLeaderSyncTx(sync *pt.LeaderSyncInfo) error {
	nodes, _, base, off, isLeader := b.getLeaderInfo()
	if len(nodes) <= 0 {
		return errors.Wrapf(pt.ErrParaNodeGroupNotSet, "id=%s", b.selfID)
	}
	syncLeader := (sync.BaseIdx + sync.Offset) % int32(len(nodes))
	//接受同步数据需要两个节点基本的共识高度相同, 两个共同leader需相同
	if sync.BaseIdx != base || nodes[syncLeader] != sync.ID {
		return errors.Wrapf(types.ErrNotSync, "peer base=%d,id=%s,self.Base=%d,id=%s", sync.BaseIdx, sync.ID, base, nodes[syncLeader])
	}
	//如果leader节点冲突，取大者
	if isLeader && off > sync.Offset {
		return errors.Wrapf(types.ErrNotSync, "self is leader, off=%d bigger than peer sync=%d", off, sync.Offset)
	}
	//更新同步过来的最新offset 高度
	atomic.CompareAndSwapInt32(&b.leaderOffset, b.leaderOffset, sync.Offset)

	//两节点不同步则不喂狗，以防止非同步或作恶节点喂狗
	atomic.StoreUint32(&b.feedDog, 1)
	return nil
}

func (b *blsClient) getLeaderInfo() ([]string, int32, int32, int32, bool) {
	//在未同步前 不处理聚合消息
	if !b.paraClient.commitMsgClient.isSync() {
		return nil, 0, 0, 0, false
	}
	nodes, _ := b.getSuperNodes()
	if len(nodes) <= 0 {
		return nil, 0, 0, 0, false
	}
	h := b.paraClient.commitMsgClient.getConsensusHeight()
	//间隔的除数再根据nodes取余数，平均覆盖所有节点
	baseIdx := int32((h / int64(b.leaderSwitchInt)) % int64(len(nodes)))
	offIdx := atomic.LoadInt32(&b.leaderOffset)
	leaderIdx := (baseIdx + offIdx) % int32(len(nodes))
	return nodes, leaderIdx, baseIdx, offIdx, nodes[leaderIdx] == b.selfID

}

func (b *blsClient) getSuperNodes() ([]string, string) {
	nodeStr, err := b.paraClient.commitMsgClient.getNodeGroupAddrs()
	if err != nil {
		return nil, ""
	}
	return strings.Split(nodeStr, ","), nodeStr
}

func (b *blsClient) isValidNodes(id string) bool {
	_, nodes := b.getSuperNodes()
	return strings.Contains(nodes, id)
}

//1. 要等到达成共识了才发送，不然处理未达成共识的各种场景会比较复杂，而且浪费手续费
func (b *blsClient) procAggregateTxs() {
	defer b.paraClient.wg.Done()
	if len(b.selfID) <= 0 {
		return
	}

out:
	for {
		select {
		case commits := <-b.rcvCommitTxCh:
			b.mutex.Lock()
			integrateCommits(b.commitsPool, commits)

			//commitsPool里面任一高度满足共识，则认为done
			nodes, _ := b.getSuperNodes()
			if !isMostCommitDone(len(nodes), b.commitsPool) {
				b.mutex.Unlock()
				continue
			}
			//自己是Leader,则聚合并发送交易
			_, _, _, _, isLeader := b.getLeaderInfo()
			if isLeader {
				b.sendAggregateTx(nodes)
			}
			//聚合签名总共消耗大约1.5ms
			//清空txsBuff，重新收集
			b.commitsPool = make(map[int64]*pt.ParaBlsSignSumDetails)
			b.mutex.Unlock()

		case <-b.quit:
			break out
		}
	}
}

func (b *blsClient) sendAggregateTx(nodes []string) error {
	dones := filterDoneCommits(len(nodes), b.commitsPool)
	plog.Info("sendAggregateTx filterDone", "commits", len(dones))
	if len(dones) <= 0 {
		return nil
	}
	acts, err := b.aggregateCommit2Action(nodes, dones)
	if err != nil {
		plog.Error("sendAggregateTx AggregateCommit2Action", "err", err)
		return err
	}
	b.paraClient.commitMsgClient.sendCommitActions(acts)
	return nil
}

func (b *blsClient) rcvCommitTx(tx *types.Transaction) error {
	if !b.isValidNodes(tx.From()) {
		plog.Error("rcvCommitTx is not valid node", "addr", tx.From())
		return pt.ErrParaNodeAddrNotExisted
	}

	txs := []*types.Transaction{tx}
	if count := tx.GetGroupCount(); count > 0 {
		group, err := tx.GetTxGroup()
		if err != nil {
			plog.Error("rcvCommitTx GetTxGroup ", "err", err)
			return errors.Wrap(err, "GetTxGroup")
		}
		txs = group.Txs
	}

	commits, err := b.checkCommitTx(txs)
	if err != nil {
		plog.Error("rcvCommitTx checkCommitTx ", "err", err)
		return errors.Wrap(err, "checkCommitTx")
	}

	if len(commits) > 0 {
		plog.Debug("rcvCommitTx tx", "addr", tx.From(), "height", commits[0].Status.Height)
	}

	b.rcvCommitTxCh <- commits
	return nil

}

func (b *blsClient) checkCommitTx(txs []*types.Transaction) ([]*pt.ParacrossCommitAction, error) {
	var commits []*pt.ParacrossCommitAction
	for _, tx := range txs {
		//验签
		if !tx.CheckSign() {
			return nil, errors.Wrapf(types.ErrSign, "hash=%s", common.ToHex(tx.Hash()))
		}
		var act pt.ParacrossAction
		err := types.Decode(tx.Payload, &act)
		if err != nil {
			return nil, errors.Wrap(err, "decode act")
		}
		if act.Ty != pt.ParacrossActionCommit {
			return nil, errors.Wrapf(types.ErrInvalidParam, "act ty=%d", act.Ty)
		}
		//交易签名和bls签名用户一致
		commit := act.GetCommit()
		if tx.From() != commit.Bls.Addrs[0] {
			return nil, errors.Wrapf(types.ErrFromAddr, "from=%s,bls addr=%s", tx.From(), commit.Bls.Addrs[0])
		}
		//验证bls 签名
		err = b.verifyBlsSign(tx.From(), commit)
		if err != nil {
			return nil, errors.Wrapf(pt.ErrBlsSignVerify, "from=%s", tx.From())
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

func hasCommited(addrs []string, addr string) (bool, int) {
	for i, a := range addrs {
		if a == addr {
			return true, i
		}
	}
	return false, 0
}

//整合相同高度commits
func integrateCommits(pool map[int64]*pt.ParaBlsSignSumDetails, commits []*pt.ParacrossCommitAction) {
	for _, cmt := range commits {
		if _, ok := pool[cmt.Status.Height]; !ok {
			pool[cmt.Status.Height] = &pt.ParaBlsSignSumDetails{Height: cmt.Status.Height}
		}
		a := pool[cmt.Status.Height]
		found, i := hasCommited(a.Addrs, cmt.Bls.Addrs[0])
		if found {
			a.Msgs[i] = types.Encode(cmt.Status)
			a.Signs[i] = cmt.Bls.Sign
			continue
		}

		a.Addrs = append(a.Addrs, cmt.Bls.Addrs[0])
		a.Msgs = append(a.Msgs, types.Encode(cmt.Status))
		a.Signs = append(a.Signs, cmt.Bls.Sign)
	}
}

//txBuff中任一高度满足done则认为ok，有可能某些未达成的高度是冗余的，达成共识的高度发给链最终判决
func isMostCommitDone(peers int, txsBuff map[int64]*pt.ParaBlsSignSumDetails) bool {
	if peers <= 0 {
		return false
	}

	for i, v := range txsBuff {
		most, _ := util.GetMostCommit(v.Msgs)
		if util.IsCommitDone(peers, most) {
			plog.Info("blssign.isMostCommitDone", "height", i, "most", most, "peers", peers)
			return true
		}
	}
	return false
}

//找出共识并达到2/3的commits， 并去除与共识不同的commits,为后面聚合签名做准备
func filterDoneCommits(peers int, pool map[int64]*pt.ParaBlsSignSumDetails) []*pt.ParaBlsSignSumDetails {
	var seq []int64
	for i, v := range pool {
		most, hash := util.GetMostCommit(v.Msgs)
		if !util.IsCommitDone(peers, most) {
			plog.Debug("blssign.filterDoneCommits not commit done", "height", i)
			continue
		}
		seq = append(seq, i)

		//只保留与most相同的commits做聚合签名使用
		a := &pt.ParaBlsSignSumDetails{Height: i, Msgs: [][]byte{[]byte(hash)}}
		for j, m := range v.Msgs {
			if bytes.Equal([]byte(hash), m) {
				a.Addrs = append(a.Addrs, v.Addrs[j])
				a.Signs = append(a.Signs, v.Signs[j])
			}
		}
		pool[i] = a
	}

	if len(seq) <= 0 {
		return nil
	}

	//从低到高找出连续的commits
	sort.Slice(seq, func(i, j int) bool { return seq[i] < seq[j] })
	var signs []*pt.ParaBlsSignSumDetails
	//共识高度要连续，不连续则退出
	lastSeq := seq[0] - 1
	for _, h := range seq {
		if lastSeq+1 != h {
			return signs
		}
		signs = append(signs, pool[h])
		lastSeq = h
	}
	return signs

}

//聚合多个签名为一个签名，并设置地址bitmap
func (b *blsClient) aggregateCommit2Action(nodes []string, commits []*pt.ParaBlsSignSumDetails) ([]*pt.ParacrossCommitAction, error) {
	var notify []*pt.ParacrossCommitAction
	for _, v := range commits {
		a := &pt.ParacrossCommitAction{Bls: &pt.ParacrossCommitBlsInfo{}}
		s := &pt.ParacrossNodeStatus{}
		types.Decode(v.Msgs[0], s)
		a.Status = s

		sign, err := b.aggregateSigns(v.Signs)
		if err != nil {
			return nil, errors.Wrapf(err, "bls aggregate=%s", v.Addrs)
		}
		a.Bls.Sign = sign.Bytes()
		bits, remains := util.SetAddrsBitMap(nodes, v.Addrs)
		plog.Debug("AggregateCommit2Action", "nodes", nodes, "addr", v.Addrs, "bits", common.ToHex(bits), "height", v.Height)
		if len(remains) > 0 {
			plog.Info("bls.signDoneCommits", "remains", remains)
		}
		a.Bls.AddrsMap = bits
		notify = append(notify, a)
	}
	return notify, nil
}

func (b *blsClient) aggregateSigns(signs [][]byte) (crypto.Signature, error) {
	var signatures []crypto.Signature
	for _, data := range signs {
		si, err := b.cryptoCli.SignatureFromBytes(data)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, si)
	}
	agg, err := crypto.ToAggregate(b.cryptoCli)
	if err != nil {
		return nil, types.ErrNotSupport
	}

	return agg.Aggregate(signatures)
}

func (b *blsClient) setBlsPriKey(secpPrkKey []byte) {
	b.blsPriKey = b.getBlsPriKey(secpPrkKey)
	b.blsPubKey = b.blsPriKey.PubKey()
	serial := b.blsPubKey.Bytes()
	plog.Debug("para commit get pub bls", "pubkey", common.ToHex(serial[:]))
}

func (b *blsClient) getBlsPriKey(key []byte) crypto.PrivKey {
	var newKey [common.Sha256Len]byte
	copy(newKey[:], key)
	for {
		pri, err := b.cryptoCli.PrivKeyFromBytes(newKey[:])
		if nil != err {
			plog.Debug("para commit getBlsPriKey try", "key", common.ToHex(newKey[:]))
			copy(newKey[:], common.Sha256(newKey[:]))
			continue
		}
		return pri
	}
}

//transfer secp256 Private key to bls pub key
func (b *blsClient) secp256Prikey2BlsPub(key string) (string, error) {
	secpPrkKey, err := getSecpPriKey(key)
	if err != nil {
		plog.Error("getSecpPriKey", "err", err)
		return "", err
	}
	blsPriKey := b.getBlsPriKey(secpPrkKey.Bytes())
	blsPubKey := blsPriKey.PubKey()
	serial := blsPubKey.Bytes()
	return common.ToHex(serial[:]), nil
}

func (b *blsClient) blsSign(commits []*pt.ParacrossCommitAction) error {
	for _, cmt := range commits {
		data := types.Encode(cmt.Status)

		cmt.Bls = &pt.ParacrossCommitBlsInfo{Addrs: []string{b.selfID}}
		sig := b.blsPriKey.Sign(data)
		sign := sig.Bytes()
		if len(sign) <= 0 {
			return errors.Wrapf(types.ErrInvalidParam, "addr=%s,height=%d", b.selfID, cmt.Status.Height)
		}
		cmt.Bls.Sign = sign
		plog.Info("bls sign msg", "data", common.ToHex(data), "height", cmt.Status.Height, "sign", len(cmt.Bls.Sign), "src", len(sign))
	}
	return nil
}

func (b *blsClient) getBlsPubKey(addr string) (crypto.PubKey, error) {
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

	s, err := common.FromHex(resp.BlsPubKey)
	if err != nil {
		plog.Error("commitmsg.getNode pubkey nok", "pubkey", resp.BlsPubKey)
		return nil, err
	}
	pubKey, err := b.cryptoCli.PubKeyFromBytes(s)
	if err != nil {
		plog.Error("verifyBlsSign.DeserializePublicKey", "key", addr)
		return nil, err
	}
	plog.Info("getBlsPubKey", "addr", addr, "pub", resp.BlsPubKey, "serial", pubKey.Bytes())
	b.peersBlsPubKey[addr] = pubKey

	return pubKey, nil
}

func (b *blsClient) verifyBlsSign(addr string, commit *pt.ParacrossCommitAction) error {
	//1. 获取对应公钥
	pubKey, err := b.getBlsPubKey(addr)
	if err != nil {
		return errors.Wrapf(err, "pub key not exist to addr=%s", addr)
	}
	//2.　获取bls签名
	sig, err := b.cryptoCli.SignatureFromBytes(commit.Bls.Sign)
	if err != nil {
		return errors.Wrapf(err, "DeserializeSignature key=%s", common.ToHex(commit.Bls.Sign))
	}

	//3. 获取签名前原始msg
	msg := types.Encode(commit.Status)

	//4. 验证bls 签名
	if !pubKey.VerifyBytes(msg, sig) {
		plog.Error("paracross.Commit bls sign verify", "title", commit.Status.Title, "height", commit.Status.Height,
			"addrsMap", common.ToHex(commit.Bls.AddrsMap), "sign", common.ToHex(commit.Bls.Sign), "addr", addr)
		plog.Error("paracross.commit bls sign verify", "data", common.ToHex(msg), "height", commit.Status.Height,
			"pub", common.ToHex(pubKey.Bytes()))
		return pt.ErrBlsSignVerify
	}
	return nil
}

func (b *blsClient) showTxBuffInfo() *pt.ParaBlsSignSumInfo {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	var ret pt.ParaBlsSignSumInfo

	reply, err := b.paraClient.SendFetchP2PTopic()
	if err != nil {
		plog.Error("fetch p2p topic", "err", err)
	}
	ret.Topics = append(ret.Topics, reply.Topics...)

	var seq []int64
	for k := range b.commitsPool {
		seq = append(seq, k)
	}
	sort.Slice(seq, func(i, j int) bool { return seq[i] < seq[j] })

	for _, s := range seq {
		sum := b.commitsPool[s]
		info := &pt.ParaBlsSignSumDetailsShow{Height: s}
		for i, a := range sum.Addrs {
			info.Addrs = append(info.Addrs, a)
			info.Msgs = append(info.Msgs, common.ToHex(sum.Msgs[i]))
		}
		ret.Info = append(ret.Info, info)
	}
	return &ret
}
