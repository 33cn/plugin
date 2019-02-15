package pbft

import (
	"reflect"
	"time"

	pt "github.com/33cn/plugin/plugin/dapp/pbft/types"
	"github.com/golang/protobuf/proto"
)

//=====================================================
// P，Q 集合的操作(计算+保存)
//=====================================================

// 计算(更新)Q集合 消息：<n,v,d>形式
// Q 集合表示那些已经Pre-prepared的Request(节点发送过Pre-prepare 或者是 Prepare消息)
// 具有序列号n，视图v以及摘要为digest的Pre-prepare消息，并且在后面的view中没有以相同的n
// 发送过Pre-prepare或者Prepare消息
func (rep *Replica) calcQset() map[qidx]*pt.RequestViewChange_PQ {
	qset := make(map[qidx]*pt.RequestViewChange_PQ)

	// 获取目前节点的Q集合
	for n, q := range rep.qset {
		qset[n] = q
	}

	for idx, cert := range rep.certStore {
		// 节点未发送或接受过prePrepare消息，则该消息一定没有Pre-prepared，返回
		if cert.prePrepare == nil {
			continue
		}

		// 判定digest在v中以n是否已经Pre-prepared，若否，则返回
		digest := cert.digest
		if !rep.prePrepared(digest, idx.v, idx.n) {
			continue
		}

		// 寻找qset中是否有该序列号的消息，若存在且该消息已经在后来的view中Pre-prepared了
		// 则不会更新Q集合
		qi := qidx{digest, idx.n}
		if q, ok := qset[qi]; ok && q.View > idx.v {
			continue
		}

		// 满足上述条件后，更新Q集合
		qset[qi] = &pt.RequestViewChange_PQ{
			View:     idx.v,
			Sequence: idx.n,
			Digest:   digest,
		}
	}

	// 是否设置Q集合？
	//rep.qset = qset
	return qset
}

// 计算(更新)P集合 消息：<n,v,d>形式
// P 集合表示那些已经Prepared的Request(该Request在该节点已经有了prepared certificate)
// 即收集有超过判定的同视图v同序列号n的相同信息的其他节点的prepare消息，保存在节点的P集合中，同时
// 该Request没有在后面的view以相同的n获得prepared certificate
func (rep *Replica) calcPset() map[uint64]*pt.RequestViewChange_PQ {
	pset := make(map[uint64]*pt.RequestViewChange_PQ)

	// 获取目前节点的P集合，如果节点一直正常，则集合为空，因此这时候得到的是nil
	for n, p := range rep.pset {
		pset[n] = p
	}

	for idx, cert := range rep.certStore {
		// 节点未发送或接受过prePrepare消息，则该消息一定没有prepared certificate，返回
		if cert.prePrepare == nil {
			continue
		}

		// 判定digest在v中以n是否已经prepared，若否，则返回
		digest := cert.digest
		if !rep.prepared(digest, idx.v, idx.n) {
			continue
		}

		// 寻找pset中是否有该序列号的消息，若存在且该消息已经在后来的view中prepared certificate了
		// 则P集合中不增加该<prepare>，即P 集合代表的是节点保存了的，序列为n，视图在最新的<prepare>
		if p, ok := pset[idx.n]; ok && p.View > idx.v {
			continue
		}

		// 满足上述条件后，更新P集合
		pset[idx.n] = &pt.RequestViewChange_PQ{
			View:     idx.v,
			Sequence: idx.n,
			Digest:   digest,
		}
	}

	// 是否设置P集合？
	//rep.pset = pset
	return pset
}

// 获取更新后的Q集合，然后将其保存下来
func (rep *Replica) persistQset() {
	var qset []*pt.RequestViewChange_PQ

	for _, q := range rep.calcQset() {
		qset = append(qset, q)
	}

	rep.persistPQset("qset", qset)
}

// 获取更新后的P集合，然后将其保存下来
func (rep *Replica) persistPset() {
	var pset []*pt.RequestViewChange_PQ

	for _, p := range rep.calcPset() {
		pset = append(pset, p)
	}

	rep.persistPQset("pset", pset)
}

// 利用不同的键的值保存P、Q集合，P集合的键为pset，Q集合的键为qset，保存对应的PQ对象(经过protobuf二值化的)
func (rep *Replica) persistPQset(key string, set []*pt.RequestViewChange_PQ) {
	raw, err := proto.Marshal(&pt.PQset{Set: set})
	if raw == nil && err != nil {
		plog.Warn("Proto Marshall has Error")
	}
	//err = rep.consumer.StoreState(key, raw)
	//if err != nil {
	//	logger.Warningf("Replica %d could not persist pqset: %s: error: %s", instance.id, key, err)
	//}

}

//=====================================================
// View-change消息辅助操作
//=====================================================

// 判定Viewchange消息是否正确
// 是否满足P Q集合的view都小于要变更视图的view，且该消息的C集合中的序列都应该在给定的低水位线和低水位线+L之间
// 因为P Q表示的已经预准备和准备的存储的消息是没有在后面的view中出现的，如果不满足，说明该<View-change>消息已经太旧
func (rep *Replica) correctViewChange(vc *pt.RequestViewChange) bool {

	// 判定<View-change>中的P Q集合是否合理
	for _, pq := range append(vc.Pset, vc.Qset...) {
		if !(pq.View < vc.View && pq.Sequence > vc.H && pq.Sequence <= vc.H+rep.L) {
			plog.Debug("P or Q set in view-change invalid", "P/Q.View", pq.View, "<VC>.View", vc.View,
				"P/Q.Sequence", pq.Sequence, "<VC>.Sequence", vc.H)
			return false
		}
	}

	// 判定<View-change>中的C 集合是否合理
	for _, c := range vc.Cset {
		// PBFT: the paper says c.n > vc.h
		if !(c.Sequence >= vc.H && c.Sequence <= vc.H+rep.L) {
			plog.Debug("C set in view-change invalid",
				"C.Sequence", c.Sequence, "<VC>.Sequence", vc.H)
			return false
		}
	}

	return true
}

// 获取节点的ViewChange消息，获取的是节点目前接收到的所有的<View-change>
func (rep *Replica) getViewChanges() (vset []*pt.RequestViewChange) {

	if rep.viewChangeStore == nil {
		return nil
	}
	for _, vc := range rep.viewChangeStore {
		vset = append(vset, vc)
	}
	return
}

// 获取主节点中S集合已经验证过的<View-change>
func (rep *Replica) getSsetViewchanges() (vset []*pt.RequestViewChange) {

	// S集合为空，返回nil
	if rep.sset == nil {
		return nil
	}
	for _, vc := range rep.sset {
		vset = append(vset, vc)
	}
	return
}

// 因此，该方法对应了图4的第一步，寻找h
// 根据节点确认的Viewchange消息来决定初始的检查点位置
// checkpoint 表示能作为检查点的<checkpoint>
// ok 表示是否能找到这个检查点
// replicas 表示给这个检查点提交了<view-change>的节点id
func (rep *Replica) selectInitialCheckpoint(vset []*pt.RequestViewChange) (checkpoint pt.RequestViewChange_C, ok bool, replicas []uint64) {
	checkpoints := make(map[pt.RequestViewChange_C][]*pt.RequestViewChange)
	for _, vc := range vset {
		for _, c := range vc.Cset { // TODO, verify that we strip duplicate checkpoints from this set
			plog.Debug("Appending checkpoint...", "From replica", vc.Replica, "Add seqNo", vc.H, "Add h", c.Sequence)
			checkpoints[*c] = append(checkpoints[*c], vc)
		}
	}

	if len(checkpoints) == 0 {
		plog.Debug("No checkpoints to select", "Replica", rep.id)
		return
	}

	for idx, vcList := range checkpoints {
		// 这里表示某一个checkpoint对应的View-change消息没有满足weak certificate的要求
		// 为了确保f+1个View-change消息都达到了这个检查点(即weak certificate)
		// 然后我们选择号码高于至少f+1个非故障副本的日志中的低水位线的检查点作为检查点
		// vcList表示了检查点的集合，对应论文S'
		// S'是S的子集，满足其长度大于f+1
		if len(vcList) <= int(rep.replicaF) {
			// 若不满足，说明这个checkpoint点是没有weak certificate的，因此不能作为新的检查点
			plog.Debug("No weak certificate for Sequence", "Sequence", idx.Sequence, "Replica", rep.id)
			continue
		}

		var count uint64
		// vset表示了所有的验证过的View-change消息组成的集合，对应论文S
		// 下面的判定对应图4的，任意的S中的消息，这些消息的h都 <= 我们选择的n
		// 因此，要满足|S|>=2f
		for _, vc := range vset {
			if vc.H <= idx.Sequence {
				count++
			}
		}

		if count < rep.quorumLimit()-1 {
			plog.Info("No Quorum for sequence", "Sequence", idx.Sequence, "Replica", rep.id)
			continue
		}

		replicas = make([]uint64, len(vcList))
		for i, vc := range vcList {
			replicas[i] = vc.Replica
		}

		// 这里是满足论文中所说的，选择具有最高号码h的检查点，即可能有很多满足上述的检查点，这里我们选择最高号码的
		if checkpoint.Sequence <= idx.Sequence {
			checkpoint = idx
			ok = true
		}
	}

	// 如果我们找不到，则返回的为nil, false, nil
	return
}

// 根据表单，分配SequenceNumber，这里是对应论文中的4.5的图4
func (rep *Replica) assignSequenceNumbers(vset []*pt.RequestViewChange, h uint64) (msgList map[uint64]string) {

	// msgList：为序列号在前一步选择的h到h+L分配预准备请求
	msgList = make(map[uint64]string)

	maxN := h + 1

	// "for all n such that h < n <= h + L"
nLoop:
	for n := h + 1; n <= h+rep.L; n++ {
		// "∃m ∈ S..." 下面的循环对应A条件
		for _, m := range vset {
			// "对于 <n,d,v> ∈ m.P"，即对于S中的某一个<View-change>，去遍历他的P集合
			// P集合中的这个消息如果满足A1，A2两个条件，则作为该序列号下的预准备请求
			for _, em := range m.Pset {
				var count uint64
				// "A1. ∃2f+1 个消息 m' ∈ S"
			mpLoop:
				for _, mp := range vset {
					// 首先是对于满足A1的第一个，m'.h < n，因此对于 >= 的不计数
					if mp.H >= n {
						continue
					}
					// "∀<n,d',v'> ∈ m'.P"
					for _, emp := range mp.Pset {
						// 这里是A1的第二个条件，对于P集合中序列为n的消息，要么视图小于<View-change>的视图，要么同视图且与摘要相同
						if n == emp.Sequence && !(emp.View < em.View || (emp.View == em.View && emp.Digest == em.Digest)) {
							continue mpLoop
						}
					}
					count++
				}

				// 至少要收集到2f+1个才行，此时A1满足?
				// 2f还是2f+1？?
				if count < rep.quorumLimit() {
					continue
				}

				// 重置判定数，判定条件A2
				count = 0
				// "A2. ∃f+1 个消息 m' ∈ S"
				for _, mp := range vset {
					// "∃<n,d',v'> ∈ m'.Q"
					for _, emp := range mp.Qset {
						if n == emp.Sequence && emp.View >= em.View && emp.Digest == em.Digest {
							count++
						}
					}
				}

				// 至少要收集到f+1个才行，此时A2满足
				if count < rep.replicaF+1 {
					continue
				}

				// 那么选择REQUEST，以d为摘要付给序列n
				// 此时为序列n分配了请求，其摘要为d
				msgList[n] = em.Digest
				maxN = n

				// 然后我们继续循环直至为所有的都分配了请求
				continue nLoop
			}
		}

		// 对于不满足A中的条件的，我们尝试是否能用B来分配
		// 首先初始化判定数
		var count uint64
		// "else if ∃2f+1 messages m ∈ S"，对应B条件
	nullLoop:
		for _, m := range vset {

			// 首先要满足m的消息的h都要 < n
			if m.H >= n {
				continue nullLoop
			}
			// 然后要满足m的P集合没有n的入口，即P中没有序列号为n的储存信息
			for _, em := range m.Pset {
				if em.Sequence == n {
					continue nullLoop
				}
			}
			// 同时满足之后，计数才增加
			count++
		}

		// 如果我们找得到2f+1个这样的，那么B条件满足?
		// 2f还是2f+1?
		if count >= rep.quorumLimit() {
			// 那么选择空的REQUEST付给序列n
			msgList[n] = ""

			continue nLoop
		}

		// 对于A，B都不满足的情况下，我们没有办法给序列号分配请求，因此返回为空
		plog.Warn("Could not assign value to contents of sequence", "Sequence", n)
		return nil
	}

	// 这里n循环已经跑完了，即为h到h+L的所有序列号都分配了请求(可能有空请求)
	// 下面是用于删除其中最后一个非空请求前所有的空请求
	// maxN代表了最后一个非空请求的序列号
	for n, msg := range msgList {
		if n > maxN && msg == "" {
			delete(msgList, n)
		}
	}

	return
}

// 重新提交客户端请求，如果有必要的话
func (rep *Replica) resubmitRequest() {
	/*
		if rep.primaryID() != rep.id {
			plog.Warn("Only Primary can resubmit Client Request", "Replica", rep.id,
				"Now Primary", rep.view)
			return
		}
	*/

	var submissionOrder []*pt.RequestClient

	// 这一层循环是为了让新的主节点去重新提交在请求队列中还没有被处理的请求
outer:
	for d, reqBatch := range rep.outstandingReq {
		for _, cert := range rep.certStore {
			if cert.digest == d {
				plog.Debug("Replica already has certificate for request, not going to resubmit",
					"Replica", rep.id, "Request Digest", d)
				continue outer
			}
		}
		plog.Debug("Replica detect request must be resubmitted", "Replica", rep.id,
			"Request Digest", d)
		submissionOrder = append(submissionOrder, reqBatch)
	}
	// len为0表示没有待处理的请求
	if len(submissionOrder) == 0 {
		return
	}

	for _, clientREQ := range submissionOrder {
		// 现在对于那些还没有pre-Prepared的请求进行重新提交
		REQ := ToRequestClient(clientREQ.Op, clientREQ.Timestamp, clientREQ.Client)
		go func() {
			rep.requestChan <- REQ
		}()
	}
	return
}

//=====================================================
// View-change消息基本操作
//=====================================================

// 发送ViewChange消息
func (rep *Replica) sendViewChange() {
	// 停止 New-View 计时器
	rep.newViewTimer.Stop()

	// 删除newView的内容
	delete(rep.newViewStore, rep.view)
	// 根据论文内容，节点视图进入到v+1
	rep.view++
	rep.client = rep.replicas[rep.view - 1]
	// activeView False表示正在viewchange
	// 节点不会接收除了View-change(ack),New-view,checkpoint以外的消息
	rep.activeView = false

	// 正常情况下，节点的P Q集合都是空的，只有在发送<View-change>的时候才计算P Q集合
	rep.pset = rep.calcPset()
	rep.qset = rep.calcQset()

	// 清除节点中老的信息
	for idx := range rep.certStore {
		if idx.v < rep.view {
			delete(rep.certStore, idx)
		}
	}
	for idx := range rep.viewChangeStore {
		if idx.v < rep.view {
			delete(rep.viewChangeStore, idx)
		}
	}
	// 根据论文内容，在发送<View-change>之前，会更新这些集合，即P Q C

	// 产生<View-change>消息，不过不是Request类的
	vc := &pt.RequestViewChange{
		View:    rep.view,
		H:       rep.h,
		Replica: rep.id,
	}

	// 计算<View-change>的C集合
	for _, c := range rep.checkpointStore {
		vc.Cset = append(vc.Cset, &pt.RequestViewChange_C{
			Sequence: c.Sequence,
			Digest:   c.Digest,
		})
	}

	// 计算<View-change>的P集合，并验证
	for _, p := range rep.pset {
		if p.Sequence < rep.h {
			plog.Error("P set should not have sequence less than h(low-Watermark)", "Replica", rep.id)
			return
		}
		vc.Pset = append(vc.Pset, p)
	}

	// 计算<View-change>的Q集合，并验证
	for _, q := range rep.qset {
		if q.Sequence < rep.h {
			plog.Error("Q set should not have sequence less than h(low-Watermark)", "Replica", rep.id)
			return
		}
		vc.Qset = append(vc.Qset, q)
	}

	viewchangeRequest := ToRequestViewChange(vc.View, vc.H, vc.Cset, vc.Pset, vc.Qset, vc.Replica)
	// 得到需要发送的<View-change>，准备发送视图变更消息
	plog.Info("Sending <View-change>", "Replica", rep.id)
	// 将消息发送给所有节点，所有其他节点接受该消息
	go func() {
		rep.requestChan <- viewchangeRequest
	}()

	// 根据论文，一旦节点发送了<View-change>
	// 会清除节点日志的<Pre-prepare>,<Prepare>,<Commit>
	// 因为节点的上述信息都储存在certStore中，因此只需清空cerStore就行
	rep.clearCertStore()
	// 开启View-change重发定时器，如果一定事件没有收到足够的View-change消息，则重发
	rep.vcResendTimerStart("Resend-View-change", rep.vcResendTimeout) //TODO
}

// 接受ViewChange消息
func (rep *Replica) recvViewChange(vc *pt.Request) {
	plog.Info("PBFT-receive Request", "Replica", rep.id, "Type", "<view-change>")
	vcREQ := vc.GetViewchange()

	// 要变更的view是否大于等于节点的view
	if vcREQ.View < rep.view {
		plog.Warn("<view-change> is old view", "Replica", rep.id)
		return
	}

	// 该View-change消息是否合理， P Q集合的View是否都小于要变更的View，checkpoint是否都处于要变更的水位之间
	if !rep.correctViewChange(vcREQ) {
		plog.Warn("<view-change> not correct", "Replica", rep.id)
		return
	}

	// 是否已经接受过该View-change消息
	if _, ok := rep.viewChangeStore[vcidx{vcREQ.View, vcREQ.Replica}]; ok {
		plog.Warn("Already has the message", "View", vcREQ.View, "From", vcREQ.Replica)
		return
	}

	// 该<view-change>写入log中，viewchangeStore的消息仅仅是收到的，并不一定是与ack一起确认获得证书的消息
	rep.viewChangeStore[vcidx{vcREQ.View, vcREQ.Replica}] = vcREQ

	// 下面要验证一下view是否为最新的view，若节点的Viewchange的Log中，有超过f+1个<View-change>(至少有一个正确的信息)
	// 其消息的view比节点当前的view要大，则节点会同意这个<View-change>并发送视图变更信息，同时自己开始视图变更
	// 这一步对于没能判定到主节点失效而不能发送<View-change>的正常节点是有意义的
	replicas := make(map[uint64]bool)
	minView := uint64(0)
	for idx := range rep.viewChangeStore {

		// 寻找目前<view-change>LOG中，满足<view-change>.view > rep.view的消息
		if idx.v <= rep.view {
			continue
		}

		// 发送了这些消息的节点为true，并且寻找其中view最小的消息(该view作为最近的视图变更点)
		replicas[idx.id] = true
		if minView == 0 || idx.v < minView {
			minView = idx.v
		}
	}

	// 这一步仅仅只有我们收到了足够多的(f+1)个<view-change>，这些view-change的view比节点的view还要大
	// 才执行这一步(节点发送新的<view-change>)
	if len(replicas) >= int(rep.replicaF+1) {
		plog.Info("receive more than f+1 <view-change>, start to send <view-change>", "Replica", rep.id)
		// 这里rep.view先设置为minView-1，因为在sendViewChange中会+1
		rep.view = minView - 1
		rep.sendViewChange()
		return
	}

	// 这边确认完成后就应该发送View-change-ack消息， 对于老的<View-change>，这里在上面就直接return了，因为老的没有意义
	go rep.sendViewChangeAck(vcREQ)

	// 到这里表示没有收到f+1个满足 <view-change>.view > rep.view的消息
	// 因为对于2f个<View-change>的判定，节点一定要么察觉主节点的错误，要么接收了f+1个<View-change>
	// 可能是节点已经到minView；发送过<view-change>；然后开始收集<view-change>
	// 对于viewChange的log中满足<view-change>.view 与节点的view相同，则计数增加
	var count uint64
	for idx := range rep.viewChangeStore {
		if idx.v == rep.view {
			count++
		}
	}

	plog.Info("<View-change> Total count", "Replica", rep.id, "Next view", rep.view,
		"count", count)

	// 这里是根据4.5.1进行的，为了防止视图变更太快
	// !activeView: 表示发送过<view-change>，因此节点已经开始视图变更
	// vc.View == rep.view: 表示转移的视图相同
	// count: 满足表示收到了至少2f(这里为N-f)个消息(即有这么多正确的副本)
	if !rep.activeView &&
		vcREQ.View == rep.view &&
		count >= rep.quorumLimit()-1 {
		// 本来有个重新发送<view-change>的计时器关闭
		rep.vcResendTimerStop()
		// 开启一个新计时器(new view change)
		// 才启动new view 的定时器，等待时间根据上一次的时间成指数增加
		rep.nvTimerStart("New-View", rep.lastNewViewTimeout)
		rep.lastNewViewTimeout = 2 * rep.lastNewViewTimeout
		return //viewChangeQuorumEvent{}
	}

	return

}

//=====================================================
// View-change-ack消息基本操作
//=====================================================

// 发送ViewChangeAck消息
func (rep *Replica) sendViewChangeAck(vc *pt.RequestViewChange) {

	ackRequest := ToRequestAck(vc.View, rep.id, vc.Replica, DigestViewchange(vc))
	plog.Info("Sending <View-change-ack>", "Replica", rep.id)

	go func() {
		rep.requestChan <- ackRequest
	}()

	return
}

// 接受ViewChangeAck消息
// <Ack>的作用主要是允许主节点证明由错误节点发送的<View-change>的真实性
func (rep *Replica) recvViewChangeAck(ackREQ *pt.Request) {
	plog.Info("PBFT-receive Request", "Replica", rep.id, "Type", "<view-change-ack>")

	vcAck := ackREQ.GetAck()
	// <Ack>中的View代表了下一个主节点的id，只有新的主节点处理<Ack>
	newPrimary := vcAck.View
	if rep.id != newPrimary {
		plog.Info("Not handle <View-change-ack>, except New Primary")
		return
	}

	// 要处理ack，我们首先要在Log中能找到视图为view,id为ViewchangeSender的节点发送的<View-change>
	// 否则无法加入到S集合 无法进行认证
	if _, ok := rep.viewChangeStore[vcidx{vcAck.View, vcAck.ViewchangeSender}]; !ok {
		plog.Warn("Not receive this kind of <View-change>, so delay the <View-change-ack>",
			"Replica", rep.id, "Should receive <View-change> from", vcAck.ViewchangeSender)
		go func() {
			time.Sleep(50 * time.Millisecond)
			rep.requestChan <- ToRequestAck(vcAck.View, vcAck.Replica, vcAck.ViewchangeSender, vcAck.Digest)
		}()
		return
	}

	// 多余多重<View-change-ack>消息，进行忽略
	if _, ok := rep.ackStore[ackidx{vcAck.View, vcAck.ViewchangeSender, vcAck.Replica}]; ok {
		plog.Warn("Duplicate ack message", "From", vcAck.Replica)
		return
	}

	// 写入Log，ack在ackStore中
	rep.ackStore[ackidx{vcAck.View, vcAck.ViewchangeSender, vcAck.Replica}] = vcAck

	var count uint64

	// 验证是否节点收到了节点i发出的<View-change>消息的<Ack>超过2f个
	for idx := range rep.ackStore {
		if idx.v == vcAck.View && idx.vcsender == vcAck.ViewchangeSender {
			count++
		}
	}

	// 如果收到了2f个该<View-change>的<ack>消息，则可以加入到S集合中，如果没有收到，则返回
	// 在这里，每一次更新S集合，都会试图尝试论文中的序列号选择以及消息选择过程，以产生<New-View>
	if count >= rep.quorumLimit()-1 {
		rep.sset[vcidx{vcAck.View, vcAck.ViewchangeSender}] =
			rep.viewChangeStore[vcidx{vcAck.View, vcAck.ViewchangeSender}]
	} else {
		return
	}

	// 如果满足上述，则更新了新主节点的S集合，因此尝试选择序列号及分配消息
	nv, success := rep.createNewView()
	if nv == nil || !success {
		plog.Info("Could not create <New-View>", "Replica", rep.id)
		return
	}

	// 如果能够生成<New-View>，那么新主节点就能够广播<New-View>给所有副本
	rep.sendNewView(nv)

	return
}

//=====================================================
// New-View消息基本操作
//=====================================================

// 尝试通过节点的S集合产生<New-View>
func (rep *Replica) createNewView() (*pt.RequestNewView, bool) {

	// 如果不满足论文的S的规模要求，则返回，即论文所说的
	// 当primary选择了每个号码的请求时,决策过程停止。这可能需要等待超过n-f个消息
	if len(rep.sset) < int(rep.quorumLimit()-1) {
		plog.Info("|S| is not big enough", "Replica", rep.id, "|S|", len(rep.sset),
			"Need at least", rep.quorumLimit()-1)
		return nil, false
	}

	// 获取S集合中所有的<View-change>，这里的都是验证过的<View-change>，即有<Ack>v证书的
	vset := rep.getSsetViewchanges()

	// 根据S集合选取合适的检查点，满足条件
	// 号码高于至少f+1个非故障副本的日志中的低水位线的检查点选择具有最高号码h的检查点
	cp, ok, _ := rep.selectInitialCheckpoint(vset)
	if !ok {
		plog.Warn("Could not find consistent checkpoint", "Replica", rep.id)
		return nil, false
	}

	// 选择到合适的检查点之后，根据论文
	// primary选择在h和h+L之间的每个序列号n的新视图中预准备的请求，h就是选择的检查点
	// 如果选取的msgList非空，则能够给每个序列号分配预准备请求
	msgList := rep.assignSequenceNumbers(vset, cp.Sequence)
	if msgList == nil {
		plog.Info("Could not assign sequence for new view", "Replica", rep.id)
		return nil, false
	}

	// 产生New-View，Vset，Xset分别对应论文的V和X
	nv := &pt.RequestNewView{
		View:    rep.view,
		Vset:    vset,
		Xset:    msgList,
		Replica: rep.id,
	}

	return nv, true
}

// 发送<New-View>给所有副本
func (rep *Replica) sendNewView(nv *pt.RequestNewView) {

	if _, ok := rep.newViewStore[rep.view]; ok {
		plog.Debug("Already has <new-view>", "Replica", rep.id)
		return
	}

	plog.Info("New Primary sending <new-view>", "New-Primary", rep.id)

	newViewRequest := ToRequestNewView(nv.View, nv.Vset, nv.Xset, nv.Replica)

	// 新的主节点先进行<New-view>处理，然后再广播
	// 保存<New-View>
	rep.newViewStore[rep.view] = nv
	rep.processNewView()
	// 广播给所有副本
	go func() {
		rep.requestChan <- newViewRequest
	}()
}

// 处理<New-View>
func (rep *Replica) recvNewView(nvREQ *pt.Request) {

	plog.Info("PBFT-receive <New-View>", "Replica", rep.id)
	// 关闭<New-View>计时器
	rep.nvTimerStop()
	// 获取<New-View>
	nv := nvREQ.GetNewview()

	// 验证<New-View>的正确性，满足View非负且比当前节点视图等或高，且与节点视图相同
	if nv.View <= 0 || nv.View < rep.view || rep.primaryID() != nv.Replica {
		plog.Warn("Reject invalid <New-View>", "Replica", rep.id,
			"<New-View> From", nv.Replica, "Invalid View", nv.View)
		return
	}

	// 多重<New-View>，忽略
	if rep.newViewStore[rep.view] != nil {
		if rep.primaryID() == rep.id {
			plog.Info("New Primary has done New-view", "PrimaryID", rep.view)
		} else {
			plog.Warn("Duplicate <New-View>, Ignoring", "Replica", rep.id)
		}
		return
	}

	rep.newViewStore[rep.view] = nv

	// 验证<New-View>中的V集合是否合理，由于我们有了<Ack>，所以这里我们验证的<View-change>就是sset
	// 即<New-View>中的V集合，因此我们需要通过V集合验证X集合是否合理
	// 因此下面的处理过程对应论文中：执行图4中的决策过程来检查这些消息是否支持 primary 报告的决定
	rep.processNewView()
	return
}

// 处理进行NewView
func (rep *Replica) processNewView() {

	// 主节点会先进行这一步的处理，然后才广播<New-view>
	// 节点是否有该view的<New-View>，如果没有，则退出
	nv, ok := rep.newViewStore[rep.view]
	if !ok {
		plog.Debug("Ignore processNewView as could not find in its newViewStore",
			"Replica", rep.id, "New View", rep.view)
		return
	}

	// 节点在处理完<New-View>后会activeView为true，因此<New-View>仅在处理过程中有效
	if rep.activeView {
		plog.Info("Replica is active in view, ignore newView",
			"Replica", rep.id, "Active view", rep.view)
		return
	}

	// 对应论文，重复图4的过程，利用V集合验证X集合，如果不满足，移动视图到v+2
	cp, ok, _ := rep.selectInitialCheckpoint(nv.Vset)
	if !ok {
		plog.Warn("Could not determine the initial checkpoint", "Replica", rep.id)
		go func() {
			rep.sendViewChange()
		}()
		return
	}

	msgList := rep.assignSequenceNumbers(nv.Vset, cp.Sequence)
	if msgList == nil {
		plog.Warn("Could not assign sequence number", "Replica", rep.id)
		go func() {
			rep.sendViewChange()
		}()
		return
	}

	if !(len(msgList) == 0 && len(nv.Xset) == 0) && !reflect.DeepEqual(msgList, nv.Xset) {
		plog.Warn("Failed to verify new-view Xset", "Replica", rep.id)
		go func() {
			rep.sendViewChange()
		}()
		return
	}

	// 到这里，验证完成，以与primary类似的方式对新信息负责，不过备份会多播一条
	// 用于v+1的<Prepare>

	// 因此，首先是对于新信息的处理，节点状态的改变
	// 如果节点的稳定的检查点是比<New-View>中，最高号码的检查点小
	// 则节点的检查点会移动 这一部分TODO
	if rep.h < cp.Sequence {
		//TODO
	}

	// 否则，我们直接处理后面的，即备份会多播<Prepare>
	rep.processNewView2(nv)
}
func (rep *Replica) processNewView2(nv *pt.RequestNewView) {
	plog.Info("Accepting new-view", "Replica", rep.id, "To View", rep.view)

	// 停止两个计时器
	//rep.stopTimer()
	//rep.nullRequestTimer.Stop()

	// 视图变更完成，因为后面有<Prepare>，因此activeView为true
	rep.activeView = true
	// 删除上个视图的<New-View>
	delete(rep.newViewStore, rep.view-1)

	rep.seqNo = rep.h

	// 这一步更新了节点对于之前接收到的客户端请求，其序列号比更新后的要大的请求
	// 这些请求在节点的Log要重新更新视图
	for n, d := range nv.Xset {
		if n <= rep.h {
			continue
		}

		// 寻找节点中，消息序列号 > 修改后的稳定检查点的消息
		req, ok := rep.reqStore[d]
		// 如果不是空Request或者找不到，说明有问题
		if !ok && d != "" {
			plog.Error("Missing request, for assigned prepare after fetching, this indicates a serious bug",
				"Replica", rep.id, "For Sequence", n, "With Digest", d)
		}
		preprep := &pt.RequestPrePrepare{
			View:     rep.view,
			Sequence: n,
			Digest:   d,
			Request:  req,
			Replica:  rep.id,
		}
		cert := rep.getCert(rep.view, n)
		cert.prePrepare = preprep
		cert.digest = d
		if rep.id == rep.primaryID() {
			cert.sentPreprepare = true
		}
		if n > rep.seqNo {
			rep.seqNo = n
		}
		rep.persistQset()
	}

	// 对于非主节点来说，会多播<Prepare>消息
	if rep.primaryID() != rep.id {
		for n, d := range nv.Xset {
			prep := &pt.RequestPrepare{
				View:     rep.view,
				Sequence: n,
				Digest:   d,
				Replica:  rep.id,
			}
			if n > rep.h {
				cert := rep.getCert(rep.view, n)
				cert.sentPrepare = true
				prepREQ := ToRequestPrepare(prep.View, prep.Sequence, prep.Digest, prep.Replica)
				rep.recvPrepare(prepREQ)
			}
			prepREQ := ToRequestPrepare(prep.View, prep.Sequence, prep.Digest, prep.Replica)
			go func() {
				rep.requestChan <- prepREQ
			}()
		}
	} else {
		plog.Info("Replica is now Primary, attempting to resubmit Requests",
			"PrimaryID", rep.view)
		// 重新提交Request
		rep.resubmitRequest()
	}

	// 开启处理未处理的Request的计时器
	// rep.startTimerIfOutstandingRequests()

	// 视图变更处理完成
	plog.Info("Replica done cleaning view-change artifacts",
		"Replica", rep.id)

	return
}

// 带有状态转移的处理 TODO
