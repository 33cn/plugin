// persistPset TODO

package pbft

import (
	"fmt"
	"net"
	"strings"
	"time"

	pt "github.com/33cn/plugin/plugin/dapp/pbft/types"
	"github.com/golang/protobuf/proto"
)

// qidx用于索引集合Q
type qidx struct {
	d string // d是消息的digest
	n uint64 // n是消息的序列号
}

// msgID用于索引certStore
type msgID struct {
	v uint64 // v是消息所在的视图
	n uint64 // n是消息的序列号
}

// msgCert为certStore的索引值，储存了该消息的基本信息
// 包括该节点对于该消息的状态
type msgCert struct {
	digest         string
	sentPreprepare bool
	prePrepare     *pt.RequestPrePrepare
	sentPrepare    bool
	prepare        []*pt.RequestPrepare
	sentCommit     bool
	commit         []*pt.RequestCommit
	sentReply      bool
}

// replyCert为客户端的reply证书(可以同时是客户端和节点)
// 由时间戳索引
type replyCert struct {
	reply    []*pt.ClientReply
	sentData bool
}

// chkpCert为节点的checkpoint证书
type chkpCert struct {
	sentCheckpoint bool
	checkpoints    []*pt.RequestCheckpoint
}

// checkpointMessage记录了checkpoint信息 TODO
type checkpointMessage struct {
	seqNo uint64
	id    []byte
}

// stateUpdateTarget由于更新节点信息 TODO
type stateUpdateTarget struct {
	checkpointMessage
	replicas []uint64
}

// vcidx为viewChangeStore的索引值，用于索引viewChange消息，v对应view,id对应节点id
type vcidx struct {
	v  uint64
	id uint64
}

// ackidx为viewChangeAck的索引值，用于索引viewChangeAck消息
type ackidx struct {
	v         uint64
	vcsender  uint64
	acksender uint64
}

// Replica 为PBFT节点基本结构体，同时包括了基本的方法
type Replica struct {

	// net部分
	listen  net.Listener
	address string
	// uint64 & 普通变量部分
	id               uint64          // 节点的id
	f                uint64          // 最大可容忍的错误节点数
	N                uint64          // 最大的网络节点数
	replicaCount     uint64          // 网络中的节点数
	replicaF         uint64          // 当前节点数对应的容错数
	h                uint64          // 低水位线
	K                uint64          // 检查点周期
	logMultiplier    uint64          // 乘积常数
	L                uint64          // 日志LOG的长度，具体计算为K*logMultiplier
	seqNo            uint64          // 序列号n，需要被严格监控是否为连续的数字，对于副节点来说，意义在于监控主节点给的是否连续
	view             uint64          // 节点目前的视图v
	lastExec         uint64          // 上一个执行完成后的Request序列号(发送了Reply)
	stableCheckpoint uint64          // 该节点储存的全网稳定的检查点的序列号
	lastReply        *pt.ClientReply // 上一个发送的客户端回复
	// highStateTarget  *stateUpdateTarget // 观察到的最大的弱检查点认证,TODO

	// bool部分
	activeView bool // 若为false，意味着viewchange发生
	// skipInProgress bool // 当节点重新恢复，需要找到一个新的启动点，此时此变量设置为true TODO
	//stateTransferring bool // 当状态传输执行的时候，该变量设置为true TODO
	byzantine bool // 用于测试，某个节点是否会表现拜占庭，发生任意性行为

	// map部分
	clients         map[uint64]string                   // 目前所有的客户端地址
	replicas        map[uint64]string                   // 目前所有节点的地址，由id索引
	reqStore        map[string]*pt.RequestClient        // 客户端的请求
	outstandingReq  map[string]*pt.RequestClient        // 待完成的客户端请求
	executedReq     map[uint64]*pt.RequestClient        // 已完成的客户端请求
	repStore        map[string][]*pt.ClientReply        // 已经发送的回复
	pset            map[uint64]*pt.RequestViewChange_PQ // 对应论文P
	qset            map[qidx]*pt.RequestViewChange_PQ   // 对应论文Q
	sset            map[vcidx]*pt.RequestViewChange     // 对应论文S
	viewChangeStore map[vcidx]*pt.RequestViewChange     // 发送的视图变更请求
	ackStore        map[ackidx]*pt.RequestAck           // 发送的ack的请求
	newViewStore    map[uint64]*pt.RequestNewView       // 用于追踪最新的<new-view> (发送或者接受的)
	//chkpts		map[]
	//checkpointStore map[uint64][]*pt.RequestCheckpoint // 状态检查点，由lastExec索引，结果为digest
	certStore  map[msgID]*msgCert    // prePrep,prep,commit的证书
	chkpStore  map[uint64]*chkpCert  // checkpoint的证书
	replyStore map[string]*replyCert // reply的证书，时间戳索引，客户端独有

	// slice部分
	checkpointStore []*pt.Checkpoint // 该节点储存的状态检查点

	//checkpointStore	map[]		//TODO
	//viewChangeStore map[vcidx]		//TODO
	//newViewStore	map[]		//TODO
	//chkpts    	map[uint64]		// 由序列号索引checkpoints状态，TODO
	//hChkpts 		map[uint64]		// 对应每个节点的最大的检查点序列号，TODO

	// channel部分
	requestChan chan *pt.Request
	dataChan    chan *pt.BlockData

	// Timer部分
	vcTimer            *Timer
	vcTimeout          time.Duration
	vcResendTimer      *Timer
	vcResendTimeout    time.Duration
	newViewTimer       *Timer
	newViewTimeout     time.Duration
	lastNewViewTimeout time.Duration
}

// NewReplica 创建一个节点，为构造器
func NewReplica(isNode bool, nodeID uint64, clientID uint64, peersURL string, clientURL string, primaryID uint64, f uint64, N uint64, K uint64, logMultiplier uint64, byzantine bool) (chan *pt.Request, chan *pt.BlockData, bool, string) {

	rep := &Replica{}
	isClient := false

	if !isNode {
		isClient = true
		rep.f = f                              // 仅仅只需要知道f即可
		rep.clients = make(map[uint64]string)  // 只用初始化客户端
		rep.replicas = make(map[uint64]string) // 以及节点IP
		rep.clientsInit(clientURL)
		rep.address = rep.clients[clientID-1]        // 然后监听此端口即可，用于接发消息
		rep.replicasInit(peersURL)                   // 初始化参与共识的节点IP，要向这些地址发送ClientRequest
		rep.replyStore = make(map[string]*replyCert) //只用初始化回复证书，验证f+1即可
		rep.requestChan = make(chan *pt.Request)     // 仅仅用于发送ClientRequest
		rep.dataChan = make(chan *pt.BlockData)      // 用于出块的chan

		plog.Info("PBFT Client INFO", "Address", rep.address)
		rep.startReplica(isClient)
		return rep.requestChan, rep.dataChan, isClient, rep.address
	}

	// map部分
	rep.clients = make(map[uint64]string)                       // 网络中的客户端，索引没有意义，值为ip地址
	rep.replicas = make(map[uint64]string)                      // 网络中的节点，id为索引，值为ip地址
	rep.reqStore = make(map[string]*pt.RequestClient)           // 客户请求组，消息digest索引，值为客户请求
	rep.outstandingReq = make(map[string]*pt.RequestClient)     // 待处理组，同上
	rep.executedReq = make(map[uint64]*pt.RequestClient)        // 已完成请求，同上，消息序列号索引
	rep.repStore = make(map[string][]*pt.ClientReply)           // 客户回复组，客户端地址索引，可能可以去掉
	rep.qset = make(map[qidx]*pt.RequestViewChange_PQ)          // P集合
	rep.pset = make(map[uint64]*pt.RequestViewChange_PQ)        // Q集合
	rep.sset = make(map[vcidx]*pt.RequestViewChange)            // S集合
	rep.viewChangeStore = make(map[vcidx]*pt.RequestViewChange) // 视图变更组，由视图与节点id索引
	rep.ackStore = make(map[ackidx]*pt.RequestAck)              // 变更确认组，由视图，发送视图变更的节点，发送确认的节点索引
	rep.newViewStore = make(map[uint64]*pt.RequestNewView)      // 新视图组，
	rep.certStore = make(map[msgID]*msgCert)                    // 三阶段请求证书
	rep.chkpStore = make(map[uint64]*chkpCert)                  // 检查点证书
	rep.replyStore = make(map[string]*replyCert)                //回复证书

	// slice部分
	rep.checkpointStore = []*pt.Checkpoint{ToCheckpoint(0, "Gensis")} // 稳定检查点集合

	// uint64部分
	rep.id = nodeID
	rep.view = primaryID
	rep.replicasInit(peersURL)
	rep.clientsInit(clientURL)
	rep.address = rep.replicas[rep.id-1] // 节点address由节点id决定
	rep.f = f
	rep.N = N
	if rep.f*3+1 > rep.N {
		panic(fmt.Sprintf("need at least %d enough replicas to tolerate %d byzantine faults,but only %d replicas configured", rep.f*3+1, rep.f, rep.N))
	}
	rep.replicaCount = rep.computeNinNet()
	rep.replicaF = rep.computeFinNet()
	if rep.replicaF > rep.f || rep.replicaCount > rep.N {
		panic(fmt.Sprintf("Number of peers or faluty peers out of range"))
	}
	rep.seqNo = rep.lowWaterMark()
	rep.h = rep.lowWaterMark()
	rep.K = K
	rep.logMultiplier = logMultiplier
	if rep.logMultiplier < 2 {
		panic("Log multiplier must be greater than or equal to 2")
	}
	rep.L = rep.K * rep.logMultiplier
	rep.stableCheckpoint = rep.lowWaterMark()
	rep.lastExec = rep.seqNo
	rep.lastReply = nil

	// bool部分
	rep.activeView = true
	rep.byzantine = byzantine

	// channel部分
	rep.requestChan = make(chan *pt.Request)
	rep.dataChan = make(chan *pt.BlockData)

	// Timer部分
	rep.vcTimeout = 60 * time.Second
	rep.vcResendTimeout = 60 * time.Second
	rep.newViewTimeout = 60 * time.Second
	rep.lastNewViewTimeout = rep.newViewTimeout
	rep.restartAllTimers()
	// plog部分
	plog.Debug("PBFT Basic Info", "Max number of validating peers", rep.N)
	plog.Debug("PBFT Basic Info", "Max number of failing peers", rep.f)
	plog.Debug("PBFT Basic Info", "Now number of peers", rep.replicaCount)
	plog.Debug("PBFT Basic Info", "log size", rep.L)
	plog.Debug("PBFT Basic Info", "byzantine flag", rep.byzantine)

	for _, c := range rep.clients {
		if rep.address == c {
			isClient = true
		}
	}
	// 开启网络监听，消息传输，启动节点
	rep.startReplica(isClient)
	return rep.requestChan, rep.dataChan, isClient, rep.address
}

//=====================================================
// 节点基本操作
//=====================================================
// 初始化节点组信息
func (rep *Replica) replicasInit(peers string) {
	for id, addr := range strings.Split(peers, ",") {
		rep.replicasAdd(uint64(id), addr)
	}
}

// 删除节点组中的节点
func (rep *Replica) replicasDelete(id uint64) {
	delete(rep.replicas, id)
}

// 增加节点组中的节点
func (rep *Replica) replicasAdd(id uint64, addr string) {
	rep.replicas[id] = addr
}

// 初始化客户端组信息
func (rep *Replica) clientsInit(clients string) {
	for id, addr := range strings.Split(clients, ",") {
		rep.clientsAdd(uint64(id), addr)
	}
}

// 增加节点中的客户端
func (rep *Replica) clientsAdd(id uint64, addr string) {
	rep.clients[id] = addr
}

// 删除节点中的客户端
func (rep *Replica) clientsDelete(id uint64) {
	delete(rep.clients, id)
}

// 启动节点
func (rep *Replica) startReplica(replicaType bool) {
	rep.startListen(replicaType)
	go rep.sendMessage()
	go rep.recvMessage()
}

// 关闭节点
func (rep *Replica) closeReplica() {
	rep.closeListen()
	rep.vcTimerClose()
	rep.vcResendTimerClose()
	rep.nvTimerClose()
}

// 接收REQUEST消息
func (rep *Replica) recvRequest(REQ *pt.Request) {
	switch REQ.Value.(type) {
	case *pt.Request_Client:
		rep.recvClientRequest(REQ)
	case *pt.Request_Preprepare:
		rep.recvPreprepare(REQ)
	case *pt.Request_Prepare:
		rep.recvPrepare(REQ)
	case *pt.Request_Commit:
		rep.recvCommit(REQ)
	case *pt.Request_Reply:
		rep.recvReply(REQ)
	case *pt.Request_Checkpoint:
		rep.recvCheckpoint(REQ)
	case *pt.Request_Viewchange:
		rep.recvViewChange(REQ)
	case *pt.Request_Ack:
		rep.recvViewChangeAck(REQ)
	case *pt.Request_Newview:
		rep.recvNewView(REQ)
	default:
		plog.Debug("Request for unknown type, ignoring")
		return
	}
}

// 检查点之后更新节点的状态
func (rep *Replica) updateReplicaState(n uint64) {
	h := n
	// 更新节点的信息
	// 首先清除节点Log中序列号小于n的所有消息
	for idx, cert := range rep.certStore {
		if idx.n <= h {
			plog.Info("Replica Cleaning quorum certificate", "Replica", rep.id,
				"View", rep.view)
			// rep.persistDelRequestBatch(cert.digest)
			delete(rep.reqStore, cert.digest)
			delete(rep.certStore, idx)
		}
	}

	// 删除P，Q集合，如果之前View-change过非空，则会删除
	for n := range rep.pset {
		if n <= h {
			delete(rep.pset, n)
		}
	}

	for idx := range rep.qset {
		if idx.n <= h {
			delete(rep.qset, idx)
		}
	}

	// 然后移动节点的低水位线
	rep.h = h
	plog.Info("Replica low-waterMark update", "low-waterMark", rep.h)
}

//=====================================================
// Timer定时器处理
//=====================================================

//每个节点的定时器操作，主要是判定View-change和New-View
func (rep *Replica) restartAllTimers() {
	rep.vcTimer = &Timer{}
	rep.vcResendTimer = &Timer{}
	rep.newViewTimer = &Timer{}
	rep.vcTimerRestart()
	rep.vcResendTimerRestart()
	rep.nvTimerRestart()
}

// 用于发送View-change的定时器
func (rep *Replica) vcTimerStart(reason string, duration time.Duration) {
	rep.vcTimer.Reset(reason, duration)
}
func (rep *Replica) vcTimerRestart() {
	rep.vcTimer.init()
	go rep.vcTimerLoop()
}
func (rep *Replica) vcTimerStop() {
	rep.vcTimer.Stop()
}
func (rep *Replica) vcTimerClose() {
	rep.vcTimer.Close()
}
func (rep *Replica) vcTimerLoop() {
	for {
		select {
		case rep.vcTimer.pending = <-rep.vcTimer.startChan:
			plog.Info("PBFT-vcTimer Starting", "Replica", rep.id)
			rep.vcTimer.stopSignal = false
			rep.vcTimer.timerChan = time.After(rep.vcTimer.pending.duration)
		case <-rep.vcTimer.stopChan:
			plog.Info("PBFT-vcTimer Stop", "Replica", rep.id)
			rep.vcTimer.stopSignal = true
		case <-rep.vcTimer.closeChan:
			plog.Info("vcTimer has been closed")
			return
		case <-rep.vcTimer.timerChan:
			if !rep.vcTimer.stopSignal {
				plog.Info("PBFT-vcTimer Fired", "Replica", rep.id,
					"Reason", rep.vcTimer.pending.reason)
				rep.vcTimer.timerChan = nil
				switch rep.vcTimer.pending.reason {
				case "View-change":
					rep.sendViewChange()
				case "Testing":
					plog.Info("We are now testing", "Replica", rep.id)
				default:
					plog.Warn("Unknow type, vcTimer stop", "Replica", rep.id)
					return
				}
			} else {
				plog.Info("vcTimer is stoped, not fired", "Replica", rep.id)
			}
		}
	}
}

// 用于重发View-change的定时器
func (rep *Replica) vcResendTimerStart(reason string, duration time.Duration) {
	rep.vcResendTimer.Reset(reason, duration)
}
func (rep *Replica) vcResendTimerRestart() {
	rep.vcResendTimer.init()
	go rep.vcResendTimerLoop()
}
func (rep *Replica) vcResendTimerStop() {
	rep.vcResendTimer.Stop()
}
func (rep *Replica) vcResendTimerClose() {
	rep.vcResendTimer.Close()
}
func (rep *Replica) vcResendTimerLoop() {
	for {
		select {
		case rep.vcResendTimer.pending = <-rep.vcResendTimer.startChan:
			plog.Info("PBFT-vcResendTimer Starting", "Replica", rep.id)
			rep.vcResendTimer.stopSignal = false
			rep.vcResendTimer.timerChan = time.After(rep.vcResendTimer.pending.duration)
		case <-rep.vcResendTimer.stopChan:
			plog.Info("PBFT-vcResendTimer Stop", "Replica", rep.id)
			rep.vcResendTimer.stopSignal = true
		case <-rep.vcResendTimer.closeChan:
			plog.Info("vcResendTimer has been closed")
			return
		case <-rep.vcResendTimer.timerChan:
			if !rep.vcResendTimer.stopSignal {
				plog.Info("PBFT-vcResendTimer Fired", "Replica", rep.id,
					"Reason", rep.vcResendTimer.pending.reason)
				rep.vcResendTimer.timerChan = nil
				switch rep.vcResendTimer.pending.reason {
				case "Resend-View-change":
					rep.view--
					rep.sendViewChange()
				case "Testing":
					plog.Info("We are now testing", "Replica", rep.id)
				default:
					plog.Warn("Unknow type, vcTimer stop", "Replica", rep.id)
					return
				}
			} else {
				plog.Info("vcResendTimer is stoped, not fired", "Replica", rep.id)
			}
		}
	}
}

// 用于检测New-View的定时器,TODO
func (rep *Replica) nvTimerStart(reason string, duration time.Duration) {
	rep.newViewTimer.stopSignal = false
	rep.newViewTimer.Reset(reason, duration)
}
func (rep *Replica) nvTimerRestart() {
	rep.newViewTimer.init()
	go rep.nvTimerLoop()
}
func (rep *Replica) nvTimerStop() {
	rep.newViewTimer.Stop()
}
func (rep *Replica) nvTimerClose() {
	rep.newViewTimer.Close()
}
func (rep *Replica) nvTimerLoop() {
	for {
		select {
		case rep.newViewTimer.pending = <-rep.newViewTimer.startChan:
			plog.Debug("PBFT-newViewTimer Starting", "Replica", rep.id)
			rep.newViewTimer.stopSignal = false
			rep.newViewTimer.timerChan = time.After(rep.newViewTimer.pending.duration)
		case <-rep.newViewTimer.stopChan:
			plog.Debug("PBFT-newViewTimer Stop", "Replica", rep.id)
			rep.newViewTimer.stopSignal = true
		case <-rep.newViewTimer.closeChan:
			plog.Debug("newViewTimer has been closed")
			return
		case <-rep.newViewTimer.timerChan:
			if !rep.newViewTimer.stopSignal {
				plog.Info("PBFT-newViewTimer Fired", "Replica", rep.id,
					"Reason", rep.newViewTimer.pending.reason)
				rep.newViewTimer.timerChan = nil
				switch rep.newViewTimer.pending.reason {
				case "New-View":
					rep.sendViewChange()
				case "Testing":
					plog.Debug("We are now testing", "Replica", rep.id)
				default:
					plog.Warn("Unknow type, newViewTimer stop", "Replica", rep.id)
					return
				}
			} else {
				plog.Debug("newViewTimer is stoped, not fired", "Replica", rep.id)
			}
		}
	}
}

//=====================================================
// 节点辅助方法
//=====================================================
// 计算网络中目前的F，根据节点总数计算
func (rep *Replica) computeFinNet() uint64 {
	return rep.replicaCount / 3
}

// 计算网络中目前的N，根据节点总数计算
func (rep *Replica) computeNinNet() uint64 {
	return uint64(len(rep.replicas))
}

// 用于计算至少收到的证书数目，这里用N-f计算
func (rep *Replica) quorumLimit() uint64 {
	return rep.replicaCount - rep.replicaF
}

// 获取目前该节点对应的主节点ID
func (rep *Replica) primaryID() uint64 {
	return rep.view % (rep.replicaCount + 1)
}

// 计算节点的低水位线，也是稳定检查点
func (rep *Replica) lowWaterMark() uint64 {
	return rep.checkpointStore[len(rep.checkpointStore)-1].Sequence
}

// 主要是用于判断序列号是否在高低水位之间
func (rep *Replica) inW(n uint64) bool {
	return (n > rep.h) && (n <= rep.h+rep.L)
}

// 用于判断消息视图是否匹配，同时序列号在水位线之间
func (rep *Replica) inVW(v uint64, n uint64) bool {
	return rep.view == v && rep.inW(n)
}

// 判断序列号是否为检查点
func (rep *Replica) isCheckpoint(sequence uint64) bool {
	return sequence%rep.K == 0
}

// 获取节点状态的摘要，通过节点对客户端最近一次的回复获得
func (rep *Replica) stateDigest() string {
	state := rep.lastReply
	state.Replica = rep.view
	return DigestReply(state)
}

// 获得最近一次发送给客户端Client的回复，对于没有回复的情况返回空
func (rep *Replica) lastReplyToClient(client string) *pt.ClientReply {
	if reply, ok := rep.repStore[client]; ok {
		return reply[len(rep.repStore[client])-1]
	}
	return nil
}

// 获取对应v，n的证书，若不存在对应的证书，则新建
func (rep *Replica) getCert(v uint64, n uint64) (cert *msgCert) {
	idx := msgID{v, n}

	cert, ok := rep.certStore[idx]
	if ok {
		return
	}

	cert = &msgCert{}
	rep.certStore[idx] = cert
	return
}

// 获取对应n的证书，为检查点证书
func (rep *Replica) getChkpCert(n uint64) (cert *chkpCert) {

	cert, ok := rep.chkpStore[n]
	if ok {
		return
	}

	cert = &chkpCert{}
	rep.chkpStore[n] = cert
	return
}

// 获取对应timestamp的证书，若不存在对应的证书，则新建
func (rep *Replica) getReplyCert(t string) (cert *replyCert) {

	cert, ok := rep.replyStore[t]
	if ok {
		return
	}

	cert = &replyCert{}
	rep.replyStore[t] = cert
	return
}

//=====================================================
// 节点状态判定(证书)
//=====================================================

// 判定节点对于消息在视图v序列号n是否预准备(Pre-prepared)了
func (rep *Replica) prePrepared(digest string, v uint64, n uint64) bool {
	// 获取已经储存在节点日志中的请求
	_, mInLog := rep.reqStore[digest]

	// 对于需要判定的非空消息，若该消息没有被保存在请求Log中，则返回
	if digest != "" && !mInLog {
		plog.Debug("Pre-prepare Request not in Log")
		return false
	}

	// 对于在节点的Q集合中存在的同样digest同序列的请求，且视图相同
	// 意味着节点是预准备的(发送过Pre-prepare或者Prepare消息)
	// 这里在视图更新中的时候会使用，而正常情况下，Q集合为空，因此该判定正常情况不执行
	if q, ok := rep.qset[qidx{digest, n}]; ok && q.View == v {
		// plog.Info("This kind Request already in Q", "Replica", rep.id)
		return true
	}

	// 节点的证书中记录有prePrepare意味着节点接受或发送过prePrepare消息
	// 之所以进行这一步，因为对于初次发送的Pre-Prepare消息，还未储存在Q集合中，因此上面的判定未能通过
	// 这里暂时有一个小问题！需要注意！
	// 修改后为，对于主节点而言，证书中有Pre-prepare消息即认为是pre-prepared
	// 对于非主节点而言，要发送过Prepare消息的才能认为是pre-prepared
	cert := rep.certStore[msgID{v, n}]
	if cert != nil {
		p := cert.prePrepare
		if p != nil && p.View == v && p.Sequence == n && rep.primaryID() == rep.id && p.Digest == digest && cert.sentPreprepare == true {
			return true
		}
		if p != nil && p.View == v && p.Sequence == n && rep.primaryID() != rep.id && p.Digest == digest {
			return true
		}
	}
	plog.Debug("Request not pre-prepared", "Replica", rep.id, "View", rep.view, "Sequence", rep.seqNo)
	return false
}

// 判定节点对于消息在视图v序列号n是否有准备认证(Prepared certificate)了
func (rep *Replica) prepared(digest string, v uint64, n uint64) bool {

	// 若该消息未预准备，则返回
	if !rep.prePrepared(digest, v, n) {
		return false
	}

	// 对于P集合中的若能找到该序列号消息，且视图v和摘要digest都相同，则有认证
	// 同样，P集合在正常情况下为空
	if p, ok := rep.pset[n]; ok && p.View == v && p.Digest == digest {
		// plog.Info("This kind Request already in P", "Replica", rep.id)
		return true
	}

	// 准备计数，获取证书
	var count uint64
	cert := rep.certStore[msgID{v, n}]

	// 该消息证书为空
	if cert == nil {
		return false
	}

	for _, p := range cert.prepare {
		if p.View == v && p.Sequence == n && p.Digest == digest {
			count++
		}
	}
	plog.Debug("<prepare> Total count", "Replica", rep.id, "count", count)

	// quorumLimit 为 2f+1，这里因为自己不发消息，因此为2f
	return count >= rep.quorumLimit()-1
}

// 判定节点对于消息在视图v序列号n是否有提交认证(committed certificate)了
func (rep *Replica) committed(digest string, v uint64, n uint64) bool {

	// 若该消息未prepared，则一定未committed
	if !rep.prepared(digest, v, n) {
		return false
	}

	// 准备计数，获取证书
	var count uint64
	cert := rep.certStore[msgID{v, n}]

	// 该消息证书为空
	if cert == nil {
		return false
	}

	for _, c := range cert.commit {
		if c.View == v && c.Sequence == n {
			count++
		}
	}

	plog.Debug("<commit> Total count", "Replica", rep.id, "count", count)

	return count >= rep.quorumLimit()
}

func (rep *Replica) replied(c string, t string) bool {

	var count uint64
	cert := rep.replyStore[t]

	if cert == nil {
		return false
	}

	for _, r := range cert.reply {
		if r.Client == c && r.Timestamp == t {
			count++
		}
	}

	plog.Debug("<reply> Total count", "count", count)
	return count >= rep.f+1
}

// 判定节点对于检查点消息(序列号n)是否有足够的认证
func (rep *Replica) checkpointed(digest string, n uint64) bool {

	cert := rep.getChkpCert(n)
	// 若没有该序列号的检查点证书，则说明节点还没有到达该序列号，因此不处理
	if cert == nil {
		return false
	}

	// 若没有发送过对应序列号的检查点消息请求，则说明节点还没有到达该序列号，因此不处理
	if !cert.sentCheckpoint {
		return false
	}

	var count uint64
	for _, c := range cert.checkpoints {
		if c.Sequence == n && c.Digest == digest {
			count++
		}
	}

	plog.Debug("", "Replica", rep.id, "<Checkpoint> count", count)

	return count >= rep.quorumLimit()

}

//=====================================================
// 节点网络通信操作
//=====================================================

// 节点开启监听tcp端口
func (rep *Replica) startListen(replicaType bool) {
	var err error
	rep.listen, err = net.Listen("tcp", rep.address)
	if err != nil {
		plog.Error("tcp connect error", "err", err)
	}
	if replicaType {
		// 意味着是客户端(不参与共识的)
		plog.Info("Client listen start", "Address", rep.address)
	} else {
		// 意味着是共识节点
		plog.Info("Replica listen start", "Replica", rep.id, "Address", rep.address)
	}
}

// 节点关闭监听
func (rep *Replica) closeListen() error {
	return rep.listen.Close()
}

// 节点接收消息，并且将消息根据类别写入不同的channel里面
func (rep *Replica) recvMessage() {
	for {
		conn, err := rep.listen.Accept()
		if err != nil {
			plog.Error("PBFT Message accept error")
		}
		REQ := &pt.Request{}
		err = ReadMessage(conn, REQ)
		if err != nil {
			plog.Error("PBFT Message read error")
		}
		rep.recvRequest(REQ)
	}
}

// 节点发送消息，根据request和reply发送不同的方向
func (rep *Replica) sendMessage() {
	for req := range rep.requestChan {
		switch req.Value.(type) {
		case *pt.Request_Reply:
			client := req.GetReply().Client
			err := WriteMessage(client, req)
			if err != nil {
				plog.Error("PBFT", "Message write error")
			}
		default:
			err := rep.multicast(req)
			if err != nil {
				plog.Error("PBFT", "Message write error")
			}
		}
	}
}

// 节点向其他节点广播
func (rep *Replica) multicast(REQ proto.Message) error {
	for _, replica := range rep.replicas {
		err := WriteMessage(replica, REQ)
		if err != nil {
			return err
		}
	}
	return nil
}

//=====================================================
// 日志(LOG)基本操作
//=====================================================

// 存储回复信息
func (rep *Replica) persistReply(Reply *pt.ClientReply) {

	client := Reply.Client
	// 是否一定需要客户端发送的时间戳顺序呢？
	//if rep.lastReply == nil || rep.lastReply.Timestamp < Reply.Timestamp {
	//	rep.repStore[client] = append(rep.repStore[client], Reply)
	//}
	rep.repStore[client] = append(rep.repStore[client], Reply)
	rep.lastReply = Reply
}

// 存储检查点信息
func (rep *Replica) persistCheckpoint(Checkpoint *pt.Checkpoint) {

	rep.checkpointStore = append(rep.checkpointStore, Checkpoint)
	//
}

// 完全清除节点的证书
func (rep *Replica) clearCertStore() {
	rep.certStore = make(map[msgID]*msgCert)
}

//=====================================================
// 消息基本操作
//=====================================================

// 处理客户端请求
func (rep *Replica) recvClientRequest(REQ *pt.Request) {
	clientREQ := REQ.GetClient()
	digest := Hash(REQ)
	plog.Info("PBFT-receive Request", "Replica", rep.id, "Type", "<client-request>",
		"Digest", digest)

	// 这里的操作为LOG存储，以digest为索引，存储客户端请求
	rep.reqStore[digest] = clientREQ
	rep.outstandingReq[digest] = clientREQ

	if rep.activeView {
		// 除主节点外，其他节点启动vc定时器，如果一定时间没有接收到该消息的Pre-prepare，则启动vc
		if rep.primaryID() != rep.id {
			//plog.Info("If !Primary node do not receive pre-prepare after a time, Fire view-change",
			//	"Time", rep.vcTimeout)
			rep.vcTimerStart("View-change", rep.vcTimeout)
		}
	}

	if rep.primaryID() == rep.id && rep.activeView {
		// TODO 这里会停止一个定时器
		if len(rep.outstandingReq) > 1 {
			plog.Info("A client Request is now handling, so not handle another and delay",
				"Now Requests num", len(rep.outstandingReq))
			return
		}
		rep.sendPreprepare(clientREQ, digest)
	} else if rep.primaryID() != rep.id {
		// plog.Warn("Backups not send Pre-prepare", "ID", rep.id)
	} else {
		plog.Info("View-change happening, not deal with Pre-prepare")
	}
}

// 发送Pre-prepare消息
func (rep *Replica) sendPreprepare(REQ *pt.RequestClient, digest string) {
	// 分配编号
	n := rep.seqNo + 1
	// 验证客户端请求
	for _, cert := range rep.certStore {
		if pp := cert.prePrepare; pp != nil {
			if pp.View == rep.view && pp.Digest != "" &&
				pp.Digest == digest && pp.Sequence != n {
				// 意味着节点曾经在view中发送过同样的消息，但是目前给了该消息不同的序列号，对于这个不予处理
				plog.Warn("Other Pre-prepare found with same digest but different seqNo")
				return
			}
		}
	}

	// 现在判断节点的状态，若赋予的序列号不在高低水位之间，则不发送该消息
	if !rep.inW(n) {
		plog.Warn("Out of sequence number")
		return
	}

	rep.seqNo = n
	// 验证通过，说明节点未曾发送过该请求，且序列号合适
	// 新建请求
	prePrepareREQ := ToRequestPreprepare(rep.view, n, digest, REQ, rep.id)
	prePrepare := prePrepareREQ.GetPreprepare()
	// 设置证书
	cert := rep.getCert(rep.view, n)
	cert.digest = digest
	cert.prePrepare = prePrepare
	cert.sentPreprepare = true
	rep.persistQset()
	go func() {
		rep.requestChan <- prePrepareREQ
	}()
	// maybeSendCommit
}

// 处理Pre-prepare消息(来自主节点)，Pre-prepare消息与论文略有不同，为了能够验证消息正确性，把客户端的请求也加入到了数据之中
func (rep *Replica) recvPreprepare(REQ *pt.Request) {
	// plog.Debug("PBFT-receive Request", "Replica", rep.id, "Type", "<pre-prepare>")

	prePrepREQ := REQ.GetPreprepare()
	// 主节点不处理预准备消息
	if rep.primaryID() == rep.id && rep.replicaCount != 1 {
		// plog.Info("Primary not handle Pre-prepare, except there is only one replica", "Replica", rep.id)
		return
	}
	// 节点处于Viewchange，返回
	if !rep.activeView {
		plog.Info("View-change happen ignore Pre-prepare message")
		return
	}

	// Pre-prepare消息非来源主节点，返回
	if rep.primaryID() != prePrepREQ.Replica {
		plog.Warn("Pre-prepare message not from primary replica")
		return
	}

	// 消息视图不匹配或者不在高低水位线之间，返回
	if !rep.inVW(prePrepREQ.View, prePrepREQ.Sequence) {
		plog.Warn("Pre-prepare message not in Watermark or not in View")
		return
	}

	// 节点数不止一个的情况下，主节点不处理Preprepare，因此判断序列号
	// 否则就是主节点判断，因为主节点已经增加了序列号了，所以判断规则发送改变
	if rep.replicaCount != 1 {
		// 序列号有问题，直接发送<View-change>，不用等待计时器
		if prePrepREQ.Sequence != rep.seqNo+1 {
			plog.Warn("Pre-prepare message not continous")
			rep.vcTimerStop()
			rep.sendViewChange()
			return
		}
	} else {
		if prePrepREQ.Sequence != rep.seqNo {
			plog.Warn("Pre-prepare message not continous")
			rep.vcTimerStop()
			rep.sendViewChange()
			return
		}
	}
	// 验证满足后，说明主节点发送的没有问题
	rep.vcTimerStop()
	// 设置副节点的序列号，主要是用于监控主节点
	rep.seqNo = prePrepREQ.Sequence
	// 获取证书，对于非主节点而言，因为没有进行发送Pre-prepare过程，因此若第一次收到Pre-prepare消息，其证书为空
	// 对于非第一次收到消息，或者是主节点，若证书中的消息不匹配(因为是以同view同sequence发送的)这里认为主节点作恶，发送Viewchange消息
	cert := rep.getCert(prePrepREQ.View, prePrepREQ.Sequence)
	if cert.digest != "" && cert.digest != prePrepREQ.Digest {
		plog.Warn("Pre-prepare message has same view&seqNo but digest different")
		rep.sendViewChange()
		return
	}

	// 对于第一次收到该消息的副本，初始化证书
	cert.prePrepare = prePrepREQ
	cert.digest = prePrepREQ.Digest

	// 储存客户端消息，不管客户端有没有给副本发送消息，或者是发送过程中出现问题副本没有接受到消息，都要进行客户端消息存储判定
	// 这也对应着Pre-prepare消息的验证
	if request, reqok := rep.reqStore[prePrepREQ.Digest]; !reqok && prePrepREQ.Digest != "" {
		digest := DigestClientRequest(prePrepREQ.Request)
		// 对于rep没有找到消息情况下，消息不匹配的Pre-prepare请求，返回
		if prePrepREQ.Digest != digest {
			plog.Warn("Pre-prepare not match Client request")
			return
		}
		rep.reqStore[prePrepREQ.Digest] = prePrepREQ.GetRequest()
		rep.outstandingReq[prePrepREQ.Digest] = prePrepREQ.GetRequest()
		//instance.persistRequestBatch(digest)
	} else if prePrepREQ.Digest != "" {
		// 在找到了相应的请求，且请求非空情况下，进行消息比对，不匹配，返回
		digest := DigestClientRequest(request)
		if prePrepREQ.Digest != digest {
			plog.Warn("Pre-prepare not match Client request")
			return
		}
		rep.reqStore[prePrepREQ.Digest] = prePrepREQ.GetRequest()
		rep.outstandingReq[prePrepREQ.Digest] = prePrepREQ.GetRequest()
	} else {
		plog.Warn("Pre-prepare message digest nil")
		return
	}

	// 上述验证都通过之后，节点会同意主节点的分配，因此会发送Prepare信息，同时还会启动定时器
	// 启动定时器 TODO
	//

	// 对于没有对该消息发送过Prepare消息的节点，发送Prepare消息
	// 这里有个问题，是否是发送过该Prepare消息的节点就不能再发送了呢？
	if !cert.sentPrepare {
		// plog.Info("Backup multicast Prepare message")
		prepareREQ := ToRequestPrepare(prePrepREQ.View, prePrepREQ.Sequence, prePrepREQ.Digest, rep.id)
		cert.sentPrepare = true
		rep.persistQset()
		go func() {
			rep.requestChan <- prepareREQ
		}()
	}
}

// 处理Prepare消息(来自非主节点)
func (rep *Replica) recvPrepare(REQ *pt.Request) {

	prepREQ := REQ.GetPrepare()

	plog.Debug("PBFT-receive Request", "Replica", rep.id, "Type", "<prepare>", "From", prepREQ.Replica,
		"Sequence", prepREQ.Sequence)

	// 正常情况下，主节点不发送Prepare消息，对于主节点的直接返回
	if rep.primaryID() == prepREQ.Replica && rep.replicaCount != 1 {
		plog.Warn("Received prepare from primary, ignore, except one replica")
		return
	}

	// 消息是否处于视图内且位于高低水平线之间
	if !rep.inVW(prepREQ.View, prepREQ.Sequence) {
		plog.Warn("Prepare message not in Watermark or not in View")
		return
	}
	/*
		// 对于没有Pre-prepared的节点，暂时不处理Prepare信息，直至该节点Pre-prepared
		if !rep.prePrepared(prepREQ.Digest, prepREQ.View, prepREQ.Sequence) {
			// plog.Info("Replica do not send <pre-prepare> or <prepare> yet", "Replica", rep.id)
			go func() {
				rep.requestChan <- ToRequestPrepare(prepREQ.View, prepREQ.Sequence, prepREQ.Digest, prepREQ.Replica)
			}()
			return
		}
	*/
	// 获取消息证书
	cert := rep.getCert(prepREQ.View, prepREQ.Sequence)

	// 对于某个节点发送了多次相同Prepare消息，直接返回
	// 对于收到了不同digest的消息，在后面进行处理
	for _, prevPrepare := range cert.prepare {
		if prevPrepare.Replica == prepREQ.Replica {
			plog.Warn("Duplicate prepare message, ignore", "From", prepREQ.Replica)
			return
		}
	}

	// 添加来自不同节点的prepare消息，同时储存P集合
	cert.prepare = append(cert.prepare, prepREQ)
	// rep.persistPset() TODO
	rep.maybeSendCommit(prepREQ)
}

// 对于pre-prepared的节点，统计收到的prepare消息，若满足条件，则该消息在该节点准备成功，发送commit消息
func (rep *Replica) maybeSendCommit(REQ *pt.RequestPrepare) {
	cert := rep.getCert(REQ.View, REQ.Sequence)

	// 该消息若prepared并且节点未发送过commit
	if rep.prepared(REQ.Digest, REQ.View, REQ.Sequence) && !cert.sentCommit {
		commitREQ := ToRequestCommit(REQ.View, REQ.Sequence, REQ.Digest, rep.id)
		cert.sentCommit = true
		go func() {
			rep.requestChan <- commitREQ
		}()
	}
}

// 处理Commit消息(来自所有节点)
func (rep *Replica) recvCommit(REQ *pt.Request) {

	commitREQ := REQ.GetCommit()

	plog.Debug("PBFT-receive Request", "Replica", rep.id, "Type", "<commit>", "From", commitREQ.Replica,
		"Sequence", commitREQ.Sequence)

	/*
		// 对于没有Prepared的节点，暂时不处理Commit信息，直至该节点Prepared
		if !rep.prepared(commitREQ.Digest, commitREQ.View, commitREQ.Sequence) {
			// plog.Info("Replica not prepared yet", "Replica", rep.id)
			go func() {
				rep.requestChan <- ToRequestCommit(commitREQ.View, commitREQ.Sequence, commitREQ.Digest, commitREQ.Replica)
			}()
			return
		}
	*/
	if !rep.inVW(commitREQ.View, commitREQ.Sequence) {
		plog.Warn("Commit message not in Watermark or not in View")
		return
	}

	// 获取证书
	cert := rep.getCert(commitREQ.View, commitREQ.Sequence)

	for _, prevCommit := range cert.commit {
		if prevCommit.Replica == commitREQ.Replica {
			plog.Warn("Duplicate commit message, ignore", "From", commitREQ.Replica)
			return
		}
	}

	cert.commit = append(cert.commit, commitREQ)
	if rep.committed(commitREQ.Digest, commitREQ.View, commitREQ.Sequence) && !cert.sentReply {
		// 停止计时器 TODO
		cert.sentReply = true
		delete(rep.outstandingReq, commitREQ.Digest)
		rep.lastExec = commitREQ.Sequence
		rep.executedReq[commitREQ.Sequence] = rep.reqStore[commitREQ.Digest]
		rep.sendReply(commitREQ)
	}

	// View-change TODO

}

// 发送回复给客户端
func (rep *Replica) sendReply(REQ *pt.RequestCommit) {

	clientRequest := rep.reqStore[REQ.Digest]
	op := clientRequest.Op
	timestamp := clientRequest.Timestamp
	client := clientRequest.Client
	block := &pt.BlockData{Value: op.Value}

	plog.Info("PBFT-send Reply to Client", "Replica", rep.id, "Client-Address", client)

	reply := ToRequestReply(REQ.View, timestamp, client, rep.id, block)

	// 存储该回复
	rep.persistReply(reply.GetReply())

	go func() {
		rep.requestChan <- reply
	}()

	rep.maybeSendCheckpoint()
}

func (rep *Replica) recvReply(REQ *pt.Request) {

	reply := REQ.GetReply()
	client := reply.Client
	timestamp := reply.Timestamp

	cert := rep.getReplyCert(timestamp)

	for _, prevReply := range cert.reply {
		if prevReply.Replica == reply.Replica {
			plog.Warn("Duplicate commit message, ignore", "From", reply.Replica)
			return
		}
	}

	cert.reply = append(cert.reply, reply)

	if rep.replied(client, timestamp) && !cert.sentData {
		cert.sentData = true
		data := reply.Result
		go func() {
			rep.dataChan <- data
		}()
	}
	return
}

// 可能会抵达检查点，因此可能发送检查点请求
func (rep *Replica) maybeSendCheckpoint() {

	// 如果没有到检查点，则判断节点是否存在有没有处理的队列请求，若有，优先执行队列请求
	if !rep.isCheckpoint(rep.seqNo) {
		if len(rep.outstandingReq) > 0 {
			for _, outstanding := range rep.outstandingReq {
				clientREQ := ToRequestClient(outstanding.Op, outstanding.Timestamp, outstanding.Client)
				rep.recvClientRequest(clientREQ)
				return
			}
		}
		return
	}

	// 若到达检查点，则发送<Checkpoint>信息
	plog.Info("Preparing checkpoint", "Replica", rep.id)
	// 摘要由最后一个Reply来得到(论文没有给摘要的具体是什么)
	state := rep.stateDigest()
	checkpointREQ := ToRequestCheckpoint(rep.seqNo, state, rep.id)

	cert := rep.getChkpCert(rep.seqNo)
	cert.sentCheckpoint = true

	go func() {
		rep.requestChan <- checkpointREQ
	}()
}

// 用于处理节点发送的检查点信息
func (rep *Replica) recvCheckpoint(REQ *pt.Request) {
	plog.Debug("PBFT-receive <Checkpoint>", "Replica", rep.id)
	checkpointREQ := REQ.GetCheckpoint()

	if !rep.inW(checkpointREQ.Sequence) {
		plog.Warn("Checkpoint message not in Watermark")
		return
	}

	cert := rep.getChkpCert(checkpointREQ.Sequence)
	for _, prevChkp := range cert.checkpoints {
		if prevChkp.Replica == checkpointREQ.Replica {
			plog.Warn("Duplicate checkpoint message, ignore", "From", checkpointREQ.Replica)
			return
		}
	}

	cert.checkpoints = append(cert.checkpoints, checkpointREQ)

	if rep.checkpointed(checkpointREQ.Digest, checkpointREQ.Sequence) {
		// 保存检查点
		// 并且如果有未处理完的请求，则优先处理这些请求
		rep.updateReplicaState(checkpointREQ.Sequence)
		checkpoint := ToCheckpoint(checkpointREQ.Sequence, checkpointREQ.Digest)
		rep.stableCheckpoint = checkpoint.Sequence
		rep.persistCheckpoint(checkpoint)
		if len(rep.outstandingReq) > 0 {
			for _, outstanding := range rep.outstandingReq {
				clientREQ := ToRequestClient(outstanding.Op, outstanding.Timestamp, outstanding.Client)
				rep.recvClientRequest(clientREQ)
				return
			}
		}
	}
}
