// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"context"
	"sync"

	"strings"

	"github.com/33cn/chain33/rpc/grpcclient"
	"github.com/33cn/chain33/types"
)

const (
	maxRollbackHeight      int64 = 10000
	defaultInvNumPerJob          = 20       // 1000 block per inv,  20inv per job
	defaultJobBufferNum          = 20       // channel buffer num for done job process
	maxBlockSize                 = 20000000 // 单次1000block size累积超过20M 需保存到localdb
	downTimesFastThreshold       = 600      //  单个server 下载超过600次，平均20次用20s，下载10分钟左右检查有没有差别比较大的
	downTimesSlowThreshold       = 40       //  慢的server小于40次，则小于快server的15倍，需要剔除
)

type connectCli struct {
	ip        string
	conn      types.Chain33Client
	downTimes int64
	isFail    bool
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
	conns          []*connectCli
	connsCheckDone bool
	multiDldOpen   bool
	wg             sync.WaitGroup
}

func (m *multiDldClient) getInvs(startHeight, endHeight int64) []*inventory {
	var invs []*inventory
	if endHeight > startHeight && endHeight-startHeight > maxRollbackHeight {
		for i := startHeight; i < endHeight; i += types.MaxBlockCountPerTime {
			inv := new(inventory)
			inv.txs = &types.ParaTxDetails{}
			inv.start = i
			inv.end = i + types.MaxBlockCountPerTime - 1
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

func (m *multiDldClient) tryMultiServerDownload() {
	if !m.multiDldOpen {
		return
	}

	paraRemoteGrpcIps := types.Conf("config.consensus.sub.para").GStr("ParaRemoteGrpcClient")
	ips := strings.Split(paraRemoteGrpcIps, ",")
	var conns []*connectCli
	for _, ip := range ips {
		conn, err := grpcclient.NewMainChainClient(ip)
		if err == nil {
			conns = append(conns, &connectCli{conn: conn, ip: ip})
		}
	}
	if len(conns) == 0 {
		plog.Info("multiDownload not valid ips")
		return
	}
	m.conns = conns

	curMainHeight, err := m.paraClient.GetLastHeightOnMainChain()
	if err != nil {
		return
	}

	//如果切换不成功，则不进行多服务下载
	_, localBlock, err := m.paraClient.switchLocalHashMatchedBlock()
	if err != nil {
		return
	}

	//获取批量下载区间和数量，给curMainHeight留10000的回滚buffer
	totalInvs := m.getInvs(localBlock.MainHeight+1, curMainHeight-maxRollbackHeight)
	totalInvsNum := int64(len(totalInvs))
	if totalInvsNum == 0 {
		plog.Info("multiDownload no invs need download")
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
}

func (i *inventory) getFirstBlock(d *downloadJob) *types.ParaTxDetail {
	if i.isSaveDb {
		block, err := d.getBlockFromDb(i.start)
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
		block, err := d.getBlockFromDb(i.end)
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

func (d *downloadJob) process() {
	for _, inv := range d.invs {
		if inv.isSaveDb {
			for i := inv.start; i <= inv.end; i++ {
				block, err := d.getBlockFromDb(i)
				if err != nil {
					panic(err)
				}
				_, err = d.mDldCli.paraClient.procLocalBlock(block)
				if err != nil {
					panic(err)
				}

			}
			continue
		}

		//block需要严格顺序执行，数据库错误，panic 重新来过
		err := d.mDldCli.paraClient.procLocalBlocks(inv.txs)
		if err != nil {
			panic(err)
		}

	}
}

func (d *downloadJob) getPreVerifyBlock(inv *inventory) (*types.ParaTxDetail, error) {
	if inv.isSaveDb {

		lastBlock, err := d.getBlockFromDb(inv.curHeight - 1)
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
			d.rmvBatchMainBlocks(inv)
			inv.curHeight = inv.start
			inv.txs.Items = nil
			inv.isSaveDb = false
			plog.Error("verifyDownloadBlock.verfiy", "ip", inv.connCli.ip)
			return err
		}
	}
	return nil
}

func (d *downloadJob) saveMainBlock(height int64, block *types.ParaTxDetail) error {
	set := &types.LocalDBSet{}

	key := calcTitleMainHeightKey(types.GetTitle(), height)
	kv := &types.KeyValue{Key: key, Value: types.Encode(block)}
	set.KV = append(set.KV, kv)

	return d.mDldCli.paraClient.setLocalDb(set)
}

func (d *downloadJob) saveBatchMainBlocks(txs *types.ParaTxDetails) error {
	set := &types.LocalDBSet{}

	for _, block := range txs.Items {
		key := calcTitleMainHeightKey(types.GetTitle(), block.Header.Height)
		kv := &types.KeyValue{Key: key, Value: types.Encode(block)}
		set.KV = append(set.KV, kv)
	}

	return d.mDldCli.paraClient.setLocalDb(set)
}

func (d *downloadJob) rmvBatchMainBlocks(inv *inventory) error {
	set := &types.LocalDBSet{}

	for i := inv.start; i <= inv.curHeight; i++ {
		key := calcTitleMainHeightKey(types.GetTitle(), i)
		kv := &types.KeyValue{Key: key, Value: nil}
		set.KV = append(set.KV, kv)
	}

	return d.mDldCli.paraClient.setLocalDb(set)
}

func (d *downloadJob) getBlockFromDb(height int64) (*types.ParaTxDetail, error) {
	key := calcTitleMainHeightKey(types.GetTitle(), height)
	set := &types.LocalDBGet{Keys: [][]byte{key}}

	value, err := d.mDldCli.paraClient.getLocalDb(set, len(set.Keys))
	if err != nil {
		return nil, err
	}
	if len(value) == 0 || value[0] == nil {
		return nil, types.ErrNotFound
	}

	var tx types.ParaTxDetail
	err = types.Decode(value[0], &tx)
	if err != nil {
		return nil, err
	}
	return &tx, nil
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
			plog.Info("verifyInvs", "height", inv.start)
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

	var fastConns, slowConns []*connectCli

	for _, conn := range d.mDldCli.conns {
		if conn.downTimes >= downTimesFastThreshold {
			fastConns = append(fastConns, conn)
		}
		if conn.downTimes <= downTimesSlowThreshold {
			slowConns = append(slowConns, conn)
		}
	}

	if len(fastConns) > 0 {
		for _, conn := range slowConns {
			conn.isFail = true
			plog.Info("paramultiDownload.checkDownLoadRate removed server", "ip", conn.ip, "times", conn.downTimes)
		}
		d.mDldCli.connsCheckDone = true
	}

}

func (d *downloadJob) requestMainBlocks(inv *inventory) (*types.ParaTxDetails, error) {
	req := &types.ReqParaTxByTitle{IsSeq: false, Start: inv.curHeight, End: inv.end, Title: types.GetTitle()}
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

func (d *downloadJob) getInvBlocks(inv *inventory, connPool chan *connectCli) {
	defer func() {
		d.wg.Done()
	}()

	inv.curHeight = inv.start
	plog.Debug("getInvBlocks begin", "start", inv.start, "end", inv.end)
	for {
		txs, err := d.requestMainBlocks(inv)
		if err != nil {
			plog.Error("getInvBlocks connect error", "err", err, "ip", inv.connCli.ip)
			return
		}
		if len(txs.Items) == 0 {
			plog.Error("getInvBlocks not items down", "ip", inv.connCli.ip)
			continue
		}

		err = d.verifyDownloadBlock(inv, txs)
		if err != nil {
			continue
		}

		//save  之前save到db，后面区块全部save到db
		if inv.isSaveDb {
			d.saveBatchMainBlocks(txs)
		} else {
			inv.txs.Items = append(inv.txs.Items, txs.Items...)
		}

		//check done
		if txs.Items[len(txs.Items)-1].Header.Height == inv.end {
			inv.connCli.downTimes++
			plog.Info("getInvs done", "start", inv.start, "end", inv.end, "downtimes", inv.connCli.downTimes, "ip", inv.connCli.ip)
			inv.isDone = true
			connPool <- inv.connCli
			return
		}
		if !inv.isSaveDb && types.Size(inv.txs) > maxBlockSize {
			d.saveBatchMainBlocks(inv.txs)
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
		if !conn.isFail {
			connPool <- conn
		}

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
