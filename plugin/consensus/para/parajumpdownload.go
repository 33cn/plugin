// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"sync"

	"bytes"
	"sync/atomic"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
)

type paraTxBlocksJob struct {
	start    int64
	end      int64
	txBlocks *types.ParaTxDetails //有平行链交易的blocks
}

type jumpDldClient struct {
	paraClient *client
	downFail   int32
	wg         sync.WaitGroup
	mtx        sync.Mutex
}

func (j *jumpDldClient) proSaveHeaders(job *types.Headers) {
	if atomic.LoadInt32(&j.downFail) != 0 || j.paraClient.isCancel() {
		return
	}
	err := j.paraClient.saveBatchMainHeaders(job)
	if err != nil {
		plog.Error("getAllHeaders---saveHeaders", "err", err)
		atomic.StoreInt32(&j.downFail, 1)
	}
}

func (j *jumpDldClient) saveHeaderJobs(ch chan *types.Headers) {
	defer j.wg.Done()

	for job := range ch {
		j.proSaveHeaders(job)
	}
}

func (j *jumpDldClient) getAllHeaders(startHeight, endHeight int64) error {
	jobsCh := make(chan *types.Headers, defaultJobBufferNum)
	j.wg.Add(1)
	go j.saveHeaderJobs(jobsCh)

	var ret error
	for i := startHeight; i <= endHeight; i += types.MaxBlockCountPerTime {
		end := i + types.MaxBlockCountPerTime - 1
		if end > endHeight {
			end = endHeight
		}
		blocks := &types.ReqBlocks{Start: i, End: end}
		headers, err := j.paraClient.GetBlockHeaders(blocks)
		if err != nil {
			ret = err
			break
		}
		plog.Info("paraJumpDownload.getAllHeaders", "start", headers.Items[0].Height, "end", headers.Items[len(headers.Items)-1].Height)
		jobsCh <- headers

		if j.paraClient.isCancel() {
			ret = errors.New("main thread cancel")
			break
		}
	}

	close(jobsCh)
	j.wg.Wait()
	return ret
}

//校验按高度获取的block hash和前一步对应高度的blockhash比对
func verifyBlockHahs(heights []*types.BlockInfo, blocks []*types.ParaTxDetail) error {
	heightMap := make(map[int64][]byte)
	for _, h := range heights {
		heightMap[h.Height] = h.Hash
	}
	for _, b := range blocks {
		if !bytes.Equal(heightMap[b.Header.Height], b.Header.Hash) {
			plog.Error("jumpDld.verifyBlockHahs", "height", b.Header.Height,
				"heightsHash", common.ToHex(heightMap[b.Header.Height]), "tx", b.Header.Hash)
			return types.ErrBlockHashNoMatch
		}
	}
	return nil
}

func (j *jumpDldClient) getParaHeights(startHeight, endHeight int64) ([]*types.BlockInfo, error) {
	var heightList []*types.BlockInfo
	title := j.paraClient.GetAPI().GetConfig().GetTitle()
	lastHeight := int64(-1)
	for {
		req := &types.ReqHeightByTitle{Height: lastHeight, Count: int32(types.MaxBlockCountPerTime), Direction: 1, Title: title}
		heights, err := j.paraClient.GetParaHeightsByTitle(req)
		if err != nil && err != types.ErrNotFound {
			plog.Error("jumpDld.getParaTxs getHeights", "start", lastHeight, "count", req.Count, "title", title, "err", err)
			return heightList, err
		}
		if err == types.ErrNotFound || heights == nil || len(heights.Items) <= 0 {
			return heightList, nil
		}
		//分页查找，只获取范围内的高度
		for _, h := range heights.Items {
			if h.Height >= startHeight && h.Height <= endHeight {
				heightList = append(heightList, h)
			}

		}
		lastHeight = heights.Items[len(heights.Items)-1].Height
		if lastHeight >= endHeight {
			return heightList, nil
		}

		if atomic.LoadInt32(&j.downFail) != 0 || j.paraClient.isCancel() {
			return nil, errors.New("verify fail or main thread cancel")
		}
	}
}

//把不连续的平行链区块高度按offset分成二维数组，方便后面处理
func getHeightsArry(heights []*types.BlockInfo, offset int) [][]*types.BlockInfo {
	var ret [][]*types.BlockInfo
	for i := 0; i < len(heights); i += offset {
		end := i + offset
		if end > len(heights) {
			end = len(heights)
		}
		ret = append(ret, heights[i:end])
	}
	return ret
}

//按高度每次获取实际1000个有平行链交易的区块，这些区块并不一定连续，为了连续处理有交易和没有交易的区块，需要特殊设置起始结束高度，
//但每次处理的起始高度和结束高度都包含了有交易的1000个平行链高度
func getStartEndHeight(startHeight, endHeight int64, arr [][]*types.BlockInfo, i int) (int64, int64) {
	single := arr[i]
	s := startHeight
	e := single[len(single)-1].Height
	if i > 0 {
		s = arr[i-1][len(arr[i-1])-1].Height + 1
	}
	if i == len(arr)-1 {
		e = endHeight
	}

	return s, e
}

func (j *jumpDldClient) verifyTxMerkleRoot(tx *types.ParaTxDetail, headMap map[int64]*types.ParaTxDetail) error {
	var verifyTxs []*types.Transaction
	for _, t := range tx.TxDetails {
		verifyTxs = append(verifyTxs, t.Tx)
	}
	verifyTxRoot := merkle.CalcMerkleRoot(j.paraClient.GetAPI().GetConfig(), tx.Header.Height, verifyTxs)
	if !bytes.Equal(verifyTxRoot, tx.ChildHash) {
		plog.Error("jumpDldClient.verifyTxMerkelHash", "height", tx.Header.Height,
			"calcHash", common.ToHex(verifyTxRoot), "rcvHash", common.ToHex(tx.ChildHash))
		return types.ErrCheckTxHash
	}
	txRootHash := merkle.GetMerkleRootFromBranch(tx.Proofs, tx.ChildHash, tx.Index)
	if !bytes.Equal(txRootHash, headMap[tx.Header.Height].Header.TxHash) {
		plog.Error("jumpDldClient.verifyRootHash", "height", tx.Header.Height,
			"txHash", common.ToHex(txRootHash), "headerHash", common.ToHex(headMap[tx.Header.Height].Header.TxHash))

		return types.ErrCheckTxHash
	}
	return nil
}

func (j *jumpDldClient) process(job *paraTxBlocksJob) {
	if atomic.LoadInt32(&j.downFail) != 0 || j.paraClient.isCancel() {
		return
	}
	headMap := make(map[int64]*types.ParaTxDetail)
	headers, err := j.paraClient.getBatchMainHeadersFromDb(job.start, job.end)
	if err != nil {
		plog.Error("jumpDldClient.process getBatchHeader", "start", job.start, "end", job.end)
		atomic.StoreInt32(&j.downFail, 1)
		return
	}
	for _, h := range headers.Items {
		headMap[h.Header.Height] = h
	}
	if job.txBlocks != nil {
		for _, tx := range job.txBlocks.Items {
			// 1. 校验平行链交易的区块头hash 和之前读取的主链头对应高度的块hash
			if !bytes.Equal(tx.Header.Hash, headMap[tx.Header.Height].Header.Hash) {
				plog.Error("jumpDldClient.process verifyhash", "height", tx.Header.Height,
					"txHash", common.ToHex(tx.Header.Hash), "headerHash", common.ToHex(headMap[tx.Header.Height].Header.Hash))
				atomic.StoreInt32(&j.downFail, 1)
				return
			}
			// 2. 校验交易merkle根和之前读的主链头的交易rootHash
			if tx.Header.Height >= j.paraClient.subCfg.MainVrfMerkleRootForkHeight {
				err := j.verifyTxMerkleRoot(tx, headMap)
				if err != nil {
					atomic.StoreInt32(&j.downFail, 1)
					return
				}
			}
			// verify ok, attach tx block to header
			headMap[tx.Header.Height].TxDetails = tx.TxDetails
		}
	}
	err = j.paraClient.procLocalAddBlocks(headers)
	if err != nil {
		atomic.StoreInt32(&j.downFail, 1)
		plog.Error("jumpDldClient.process procLocalAddBlocks", "start", job.start, "end", job.end, "err", err)
	}
	j.paraClient.rmvBatchMainBlocks(job.start, job.end)

}

func (j *jumpDldClient) processTxJobs(ch chan *paraTxBlocksJob) {
	defer j.wg.Done()

	for job := range ch {
		j.process(job)
	}
}

//按高度list请求平行链区块，服务器有可能返回少于请求高度，少于时候需要继续请求
func (j *jumpDldClient) fetchHeightListBlocks(hlist []int64, title string) (*types.ParaTxDetails, error) {
	index := 0
	retBlocks := &types.ParaTxDetails{}
	for {
		list := hlist[index:]
		req := &types.ReqParaTxByHeight{Items: list, Title: title}
		blocks, err := j.paraClient.GetParaTxByHeight(req)
		if err != nil {
			plog.Error("jumpDld.getParaTxs fetchHeightListBlocks", "start", list[0], "end", list[len(list)-1], "title", title)
			return nil, err
		}
		retBlocks.Items = append(retBlocks.Items, blocks.Items...)
		index += len(blocks.Items)
		if index == len(hlist) {
			return retBlocks, nil
		}
		//从逻辑上应该不会有大于场景出现
		if index > len(hlist) {
			plog.Error("jumpDld.getParaTxs fetchHeightListBlocks len", "index", index, "len", len(hlist), "start", list[0], "end", list[len(list)-1], "title", title)
			return nil, err
		}
	}
}

func (j *jumpDldClient) getParaTxs(startHeight, endHeight int64, heights []*types.BlockInfo, ch chan *paraTxBlocksJob) error {
	title := j.paraClient.GetAPI().GetConfig().GetTitle()
	heightsArr := getHeightsArry(heights, int(types.MaxBlockCountPerTime))

	for i, single := range heightsArr {
		var hlist []int64
		for _, h := range single {
			hlist = append(hlist, h.Height)
		}

		blocks, err := j.fetchHeightListBlocks(hlist, title)
		if err != nil {
			plog.Error("jumpDld.getParaTxs getParaTx", "start", hlist[0], "end", hlist[len(hlist)-1], "title", title)
			return err
		}

		err = verifyBlockHahs(single, blocks.Items)
		if err != nil {
			plog.Error("jumpDld.getParaTxs verifyTx", "start", hlist[0], "end", hlist[len(hlist)-1], "title", title)
			return err
		}
		s, e := getStartEndHeight(startHeight, endHeight, heightsArr, i)
		plog.Info("jumpDld.getParaTxs fillTxJob", "start", s, "end", e, "i", i, "len", len(single))
		paraTxs := &paraTxBlocksJob{start: s, end: e, txBlocks: blocks}
		ch <- paraTxs

		if atomic.LoadInt32(&j.downFail) != 0 || j.paraClient.isCancel() {
			return errors.New("verify fail or main thread cancel")
		}
	}

	return nil
}

//Jump Download 是选择有平行链交易的区块跳跃下载的功能，分为三个步骤：
//0. 只获取当前主链高度1w高度前的区块，默认没有分叉，都是addType　block
//1. 获取完整的主链header，为了后面平行链交易的校验和平行链的空块产生，云节点主链1000个header大概30ms，固态硬盘更快
//2. 获取所有平行链交易的高度列表，大概5s以内
//3. 按高度列表获取平行链区块并获取一段执行一段
func (j *jumpDldClient) tryJumpDownload() {
	curMainHeight, err := j.paraClient.GetLastHeightOnMainChain()
	if err != nil {
		plog.Error("tryJumpDownload getMain height", "err", err.Error())
		return
	}

	//如果切换不成功，则不进行多服务下载
	_, localBlock, err := j.paraClient.switchLocalHashMatchedBlock()
	if err != nil {
		plog.Error("tryJumpDownload switch local height", "err", err.Error())
		return
	}

	startHeight := localBlock.MainHeight + 1
	endHeight := curMainHeight - maxRollbackHeight
	if !(endHeight > startHeight && endHeight-startHeight > maxRollbackHeight) {
		plog.Info("tryJumpDownload.quit", "start", startHeight, "end", endHeight)
		return
	}
	plog.Info("tryJumpDownload", "start", startHeight, "end", endHeight)
	t1 := types.Now()
	//1. get all main headers
	err = j.getAllHeaders(startHeight, endHeight)
	if err != nil {
		plog.Error("JumpDld.getAllHeaders", "err", err)
		return
	}
	plog.Info("tryJumpDownload.getAllHeaders", "time", types.Since(t1))

	//2. 获取有平行链交易的块高度列表
	t1 = types.Now()
	heights, err := j.getParaHeights(startHeight, endHeight)
	if err != nil {
		plog.Error("JumpDld.getParaHeights", "err", err)
	}
	if len(heights) == 0 {
		plog.Error("JumpDld.getParaHeights　no height found")
		return
	}
	plog.Info("tryJumpDownload.getParaHeights", "time", types.Since(t1))

	//3. 按有平行链交易的高度列表获取平行链块
	jobsCh := make(chan *paraTxBlocksJob, defaultJobBufferNum)
	j.wg.Add(1)
	go j.processTxJobs(jobsCh)

	t1 = types.Now()
	err = j.getParaTxs(startHeight, endHeight, heights, jobsCh)
	if err != nil {
		//需要close　processTxJobs　后再返回
		plog.Error("tryJumpDownload.getParaTxs", "err", err)
	}

	close(jobsCh)
	j.wg.Wait()
	plog.Info("tryJumpDownload.getParaTxs", "time", types.Since(t1))
	plog.Info("tryJumpDownload done")
}
