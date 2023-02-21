// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"context"
	"sync"

	"time"

	"strings"

	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/types"
)

const (
	maxRollbackHeight      = 10000    // 据当前主链高度回滚的余量
	defaultInvNumPerJob    = 20       // 20inv task per job
	defaultJobBufferNum    = 20       // channel buffer num for done job process
	maxBlockSize           = 20000000 // 单次1000block size累积超过20M 需保存到localdb
	downTimesFastThreshold = 450      //  单个server 下载超过450次，平均20次用20s来计算，下载7分钟左右检查有没有差别比较大的
	downTimesSlowThreshold = 30       //  慢的server小于30次，则小于快server的15倍，需要剔除
	maxServerRspTimeout    = 15
)

type connectCli struct {
	ip        string
	conn      types.Chain33Client
	downTimes int64
	timeout   uint32
}

//invertory 是每次请求的最小单位，每次请求最多MaxBlockCountPerTime
type inventory struct {
	start     int64
	end       int64
	curHeight int64
	txs       *types.ParaTxDetails
	isDone    bool
	isSaveDb  bool
	connCli   *connectCli
}

type downloadJob struct {
	mDldCli     *multiDldClient
	parentBlock *types.ParaTxDetail
	wg          sync.WaitGroup
	invs        []*inventory
}

func newDownLoadJob(cli *multiDldClient) *downloadJob {
	return &downloadJob{
		mDldCli: cli,
	}
}

type multiDldClient struct {
	paraClient     *client
	jobBufferNum   uint32
	invNumPerJob   int64
	serverTimeout  uint32
	conns          []*connectCli
	connsCheckDone bool
	wg             sync.WaitGroup
	mtx            sync.Mutex
}

func newMultiDldCli(para *client, cfg *subConfig) *multiDldClient {
	multi := &multiDldClient{
		paraClient:    para,
		invNumPerJob:  defaultInvNumPerJob,
		jobBufferNum:  defaultJobBufferNum,
		serverTimeout: maxServerRspTimeout,
	}
	if cfg.MultiDownInvNumPerJob > 0 {
		multi.invNumPerJob = cfg.MultiDownInvNumPerJob
	}
	if cfg.MultiDownJobBuffNum > 0 {
		multi.jobBufferNum = cfg.MultiDownJobBuffNum
	}

	if cfg.MultiDownServerRspTime > 0 {
		multi.serverTimeout = cfg.MultiDownServerRspTime
	}
	return multi
}

func (m *multiDldClient) getInvs(startHeight, endHeight int64) []*inventory {
	var invs []*inventory
	if endHeight > startHeight && endHeight-startHeight > maxRollbackHeight {
		for i := startHeight; i < endHeight; i += m.paraClient.subCfg.BatchFetchBlockCount {
			inv := new(inventory)
			inv.txs = &types.ParaTxDetails{}
			inv.start = i
			inv.end = i + m.paraClient.subCfg.BatchFetchBlockCount - 1
			if inv.end > endHeight {
				inv.end = endHeight
				invs = append(invs, inv)
				return invs
			}
			invs = append(invs, inv)
		}
	}
	return invs
}

func (m *multiDldClient) testConn(conn *connectCli, inv *inventory) {
	defer m.wg.Done()

	recv := make(chan bool, 1)
	testInv := &inventory{start: inv.start, end: inv.end, curHeight: inv.start, connCli: conn}

	cfg := m.paraClient.GetAPI().GetConfig()
	go func() {
		_, err := requestMainBlocks(cfg, testInv)
		if err != nil {
			plog.Info("multiServerDownload.testconn ip error", "ip", conn.ip, "err", err.Error())
			recv <- false
			return
		}
		recv <- true
	}()

	t := time.NewTimer(time.Second * time.Duration(conn.timeout))
	defer t.Stop()
	select {
	case <-t.C:
		plog.Info("multiServerDownload.testconn ip timeout", "ip", conn.ip)
		return
	case ret := <-recv:
		if ret {
			m.mtx.Lock()
			m.conns = append(m.conns, conn)
			m.mtx.Unlock()
		}
		return
	}
}

func (m *multiDldClient) getConns(inv *inventory) error {
	cfg := m.paraClient.GetAPI().GetConfig()
	paraRemoteGrpcIps := cfg.GetModuleConfig().RPC.ParaChain.MainChainGrpcAddr
	ips := strings.Split(paraRemoteGrpcIps, ",")
	var conns []*connectCli
	for _, ip := range ips {
		conn, err := grpcclient.NewMainChainClient(cfg, ip)
		if err == nil {
			conns = append(conns, &connectCli{conn: conn, ip: ip, timeout: m.serverTimeout})
		}
	}

	if len(conns) == 0 {
		plog.Info("tryMultiServerDownload no valid ips")
		return types.ErrNotFound
	}

	plog.Info("tryMultiServerDownload test connects, wait 15s...")
	for _, conn := range conns {
		m.wg.Add(1)
		go m.testConn(conn, inv)
	}
	m.wg.Wait()

	if len(m.conns) == 0 {
		plog.Info("tryMultiServerDownload not valid ips")
		return types.ErrNotFound
	}

	plog.Info("multiServerDownload test connects done")
	for _, conn := range m.conns {
		plog.Info("multiServerDownload ok ip", "ip", conn.ip)
	}

	return nil
}

//缺省不打开，因为有些节点下载时间不稳定，容易超时出错，后面看怎么优化
func (m *multiDldClient) tryMultiServerDownload() {
	curMainHeight, err := m.paraClient.GetLastHeightOnMainChain()
	if err != nil {
		plog.Error("tryMultiServerDownload getMain height", "err", err.Error())
		return
	}

	//如果切换不成功，则不进行多服务下载
	_, localBlock, err := m.paraClient.switchLocalHashMatchedBlock()
	if err != nil {
		plog.Error("tryMultiServerDownload switch local height", "err", err.Error())
		return
	}

	//获取批量下载区间和数量，给curMainHeight留10000的回滚buffer
	totalInvs := m.getInvs(localBlock.MainHeight+1, curMainHeight-maxRollbackHeight)
	totalInvsNum := int64(len(totalInvs))
	if totalInvsNum == 0 {
		plog.Info("tryMultiServerDownload no invs need download")
		return
	}

	//获取可用IP 链接
	err = m.getConns(totalInvs[0])
	if err != nil {
		return
	}

	plog.Info("tryMultiServerDownload", "start", localBlock.MainHeight+1, "end", curMainHeight-maxRollbackHeight, "totalInvs", totalInvsNum)

	jobsCh := make(chan *downloadJob, m.jobBufferNum)
	m.wg.Add(1)
	go m.processDoneJobs(jobsCh)

	preBlock := &types.ParaTxDetail{
		Type:   types.AddBlock,
		Header: &types.Header{Hash: localBlock.MainHash, Height: localBlock.MainHeight},
	}
	for i := int64(0); i < totalInvsNum; i += m.invNumPerJob {
		end := i + m.invNumPerJob
		if end > totalInvsNum {
			end = totalInvsNum
		}
		job := newDownLoadJob(m)
		job.invs = append(job.invs, totalInvs[i:end]...)
		job.parentBlock = preBlock
		job.GetBlocks()
		if m.paraClient.isCancel() {
			break
		}
		jobsCh <- job
		plog.Info("tryMultiServerDownload", "start", i, "end", end, "total", totalInvsNum)
		preBlock = job.invs[len(job.invs)-1].getLastBlock(job)

	}
	close(jobsCh)
	m.wg.Wait()
	plog.Info("tryMultiServerDownload done")
}

func (i *inventory) getFirstBlock(d *downloadJob) *types.ParaTxDetail {
	if i.isSaveDb {
		block, err := d.mDldCli.paraClient.getMainBlockFromDb(i.start)
		if err != nil {
			panic(err)
		}
		return block
	}
	return i.txs.Items[0]
}

func (i *inventory) getLastBlock(d *downloadJob) *types.ParaTxDetail {
	if !i.isDone {
		return nil
	}
	if i.isSaveDb {
		block, err := d.mDldCli.paraClient.getMainBlockFromDb(i.end)
		if err != nil {
			panic(err)
		}
		return block
	}
	return i.txs.Items[len(i.txs.Items)-1]
}

func (m *multiDldClient) processDoneJobs(ch chan *downloadJob) {
	defer m.wg.Done()

	for job := range ch {
		job.process()
	}
}

func (d *downloadJob) resetInv(i *inventory) {
	if i.isSaveDb {
		d.mDldCli.paraClient.rmvBatchMainBlocks(i.start, i.curHeight)
	}
	i.curHeight = i.start
	i.txs.Items = nil
	i.isSaveDb = false
}

func (d *downloadJob) process() {
	for _, inv := range d.invs {
		if inv.isSaveDb {
			for i := inv.start; i <= inv.end; i++ {
				block, err := d.mDldCli.paraClient.getMainBlockFromDb(i)
				if err != nil {
					panic(err)
				}
				_, err = d.mDldCli.paraClient.procLocalBlock(block)
				if err != nil {
					panic(err)
				}
			}
			d.mDldCli.paraClient.blockSyncClient.handleLocalChangedMsg()
		} else {
			//block需要严格顺序执行，数据库错误，panic 重新来过
			err := d.mDldCli.paraClient.procLocalAddBlocks(inv.txs)
			if err != nil {
				panic(err)
			}
		}
		//release memory
		inv.txs = nil
	}
}

func (d *downloadJob) getPreVerifyBlock(inv *inventory) (*types.ParaTxDetail, error) {
	if inv.isSaveDb {
		lastBlock, err := d.mDldCli.paraClient.getMainBlockFromDb(inv.curHeight - 1)
		if err != nil {
			return nil, err
		}
		return lastBlock, nil
	}
	if len(inv.txs.Items) != 0 {
		return inv.txs.Items[len(inv.txs.Items)-1], nil
	}
	return nil, nil
}

func (d *downloadJob) verifyDownloadBlock(inv *inventory, blocks *types.ParaTxDetails) error {
	//返回区块内部校验
	err := verifyMainBlocksInternal(blocks)
	if err != nil {
		plog.Error("verifyDownloadBlock internal", "ip", inv.connCli.ip)
		return err
	}

	//跟已下载的区块校验
	verifyBlock, err := d.getPreVerifyBlock(inv)
	if err != nil {
		plog.Error("verifyDownloadBlock.getPreVerifyBlock", "ip", inv.connCli.ip)
		return err
	}
	if verifyBlock != nil {
		err = verifyMainBlockHash(getVerifyHash(verifyBlock), blocks.Items[0])
		if err != nil {
			plog.Error("verifyDownloadBlock.verfiy", "ip", inv.connCli.ip)
			return err
		}
	}
	return nil
}

func (d *downloadJob) checkInv(lastRetry, pre *types.ParaTxDetail, inv *inventory) error {
	if !inv.isDone {
		return types.ErrNotFound
	}

	if lastRetry == pre {
		return nil
	}

	return verifyMainBlockHash(getVerifyHash(pre), inv.getFirstBlock(d))

}

// 对一个job里面的invs之间头尾做校验， 如果后一个跟之前的校验不过，放入retry，retry后面的一个inv暂时跳过校验，继续和后面的做校验
func (d *downloadJob) verifyInvs() []*inventory {
	var retryItems []*inventory
	pre := d.parentBlock
	var lastRetry *types.ParaTxDetail
	for _, inv := range d.invs {
		err := d.checkInv(lastRetry, pre, inv)
		if err != nil {
			plog.Info("verifyInvs error", "height", inv.start)
			retryItems = append(retryItems, inv)
			lastRetry = inv.getLastBlock(d)
		}
		pre = inv.getLastBlock(d)

	}

	return retryItems

}

func (d *downloadJob) checkDownLoadRate() {
	if d.mDldCli.connsCheckDone {
		return
	}

	var found bool
	for _, conn := range d.mDldCli.conns {
		if conn.downTimes >= downTimesFastThreshold {
			found = true
			break
		}
	}

	if found {
		var fastConns []*connectCli
		for _, conn := range d.mDldCli.conns {
			if conn.downTimes > downTimesSlowThreshold {
				fastConns = append(fastConns, conn)
			}
		}
		d.mDldCli.conns = fastConns
		d.mDldCli.connsCheckDone = true
	}

}

func requestMainBlocks(cfg *types.Chain33Config, inv *inventory) (*types.ParaTxDetails, error) {
	req := &types.ReqParaTxByTitle{IsSeq: false, Start: inv.curHeight, End: inv.end, Title: cfg.GetTitle()}
	txs, err := inv.connCli.conn.GetParaTxByTitle(context.Background(), req)
	if err != nil {
		return nil, err
	}

	for i, item := range txs.Items {
		if item != nil && item.Header.Height != int64(i)+req.Start {
			plog.Error("requestMainBlocks block notmatch", "expect", int64(i)+req.Start, "height", item.Header.Height)
			return nil, types.ErrBlockHeightNoMatch
		}
	}

	//只获取最前面的有效交易
	return validMainBlocks(txs), nil
}

func requestMainBlockWithTime(cfg *types.Chain33Config, inv *inventory) *types.ParaTxDetails {
	retCh := make(chan *types.ParaTxDetails, 1)
	go func() {
		tx, err := requestMainBlocks(cfg, inv)
		if err != nil {
			plog.Error("requestMainBlockWithTime err", "start", inv.start, "end", inv.end, "ip", inv.connCli.ip, "err", err.Error())
			close(retCh)
			return
		}
		retCh <- tx
	}()

	t := time.NewTimer(time.Second * time.Duration(inv.connCli.timeout))
	defer t.Stop()
	select {
	case <-t.C:
		plog.Debug("requestMainBlockWithTime timeout", "start", inv.start, "end", inv.end, "ip", inv.connCli.ip)
		return nil
	case ret, ok := <-retCh:
		if !ok {
			return nil
		}
		return ret
	}
}

func (d *downloadJob) getInvBlocks(inv *inventory, connPool chan *connectCli) {
	cfg := d.mDldCli.paraClient.GetAPI().GetConfig()
	start := time.Now()
	defer func() {
		connPool <- inv.connCli
		d.wg.Done()
	}()

	inv.curHeight = inv.start
	plog.Debug("getInvBlocks begin", "start", inv.start, "end", inv.end, "ip", inv.connCli.ip)
	for {
		txs := requestMainBlockWithTime(cfg, inv)
		if txs == nil || len(txs.Items) == 0 {
			d.resetInv(inv)
			plog.Error("getInvBlocks reqMainBlock nil", "ip", inv.connCli.ip)
			return
		}

		err := d.verifyDownloadBlock(inv, txs)
		if err != nil {
			d.resetInv(inv)
			return
		}

		//save  之前save到db，后面区块全部save到db
		if inv.isSaveDb {
			d.mDldCli.paraClient.saveBatchMainBlocks(txs)
		} else {
			inv.txs.Items = append(inv.txs.Items, txs.Items...)
		}

		//check done
		if txs.Items[len(txs.Items)-1].Header.Height == inv.end {
			inv.connCli.downTimes++
			plog.Info("downloadjob getInvs done", "start", inv.start, "end", inv.end, "time", time.Since(start).Nanoseconds()/1000000, "downtimes", inv.connCli.downTimes, "ip", inv.connCli.ip)
			inv.isDone = true
			return
		}
		if !inv.isSaveDb && types.Size(inv.txs) > maxBlockSize {
			d.mDldCli.paraClient.saveBatchMainBlocks(inv.txs)
			inv.txs.Items = nil
			inv.isSaveDb = true
		}
		inv.curHeight = txs.Items[len(txs.Items)-1].Header.Height + 1

	}
}

// getInvs download the block
func (d *downloadJob) getInvs(invs []*inventory) {
	connPool := make(chan *connectCli, len(d.mDldCli.conns))
	for _, conn := range d.mDldCli.conns {
		connPool <- conn
	}

	for _, inv := range invs {
		if d.mDldCli.paraClient.isCancel() {
			break
		}
		inv.connCli = <-connPool
		d.wg.Add(1)
		go d.getInvBlocks(inv, connPool)

	}
	//等待下载任务
	d.wg.Wait()
}

// GetBlocks get blocks information
func (d *downloadJob) GetBlocks() {
	invs := d.invs
	for {
		d.getInvs(invs)
		invs = d.verifyInvs()
		d.checkDownLoadRate()
		if len(invs) == 0 {
			return
		}

		if d.mDldCli.paraClient.isCancel() {
			return
		}

	}
}
