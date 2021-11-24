// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"bytes"
	"sync"
	"sync/atomic"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
)

type paraTxBlocksJob struct {
	start        int64
	end          int64
	paraTxBlocks map[int64]*types.ParaTxDetail //有平行链交易的blocks
	headers      *types.ParaTxDetails
}

type jumpDldClient struct {
	paraClient *client
	downFail   int32
	wg         sync.WaitGroup
}

func newJumpDldCli(para *client, cfg *subConfig) *jumpDldClient {
	return &jumpDldClient{paraClient: para}
}

//校验按高度获取的block hash和前一步对应高度的blockhash比对
func verifyBlockHash(heights []*types.BlockInfo, blocks []*types.ParaTxDetail) error {
	heightMap := make(map[int64][]byte)
	for _, h := range heights {
		heightMap[h.Height] = h.Hash
	}
	for _, b := range blocks {
		if !bytes.Equal(heightMap[b.Header.Height], b.Header.Hash) {
			plog.Error("jumpDld.verifyBlockHash", "height", b.Header.Height,
				"heightsHash", common.ToHex(heightMap[b.Header.Height]), "tx", b.Header.Hash)
			return types.ErrBlockHashNoMatch
		}
	}
	return nil
}

func (j *jumpDldClient) getParaHeightList(startHeight, endHeight int64) ([]*types.BlockInfo, error) {
	var heightList []*types.BlockInfo
	title := j.paraClient.GetAPI().GetConfig().GetTitle()
	lastHeight := int64(-1)
	for {
		req := &types.ReqHeightByTitle{Height: lastHeight, Count: int32(types.MaxBlockCountPerTime), Direction: 1, Title: title}
		heights, err := j.paraClient.GetParaHeightsByTitle(req)
		if err != nil && err != types.ErrNotFound {
			plog.Error("jumpDld.getParaHeightList", "start", lastHeight, "count", req.Count, "title", title, "err", err)
			return heightList, err
		}
		if err == types.ErrNotFound || heights == nil || len(heights.Items) <= 0 {
			return heightList, nil
		}
		//分页查找，只获取范围内的高度
		plog.Info("jumpDld.getParaTxHeightList", "start", heights.Items[0].GetHeight(), "end", heights.Items[len(heights.Items)-1].GetHeight())
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
func splitHeights2Rows(heights []*types.BlockInfo, offset int) [][]*types.BlockInfo {
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
func getHeaderStartEndRange(startHeight, endHeight int64, arr [][]*types.BlockInfo, i int) (int64, int64) {
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
	for _, h := range job.headers.Items {
		headMap[h.Header.Height] = h
	}

	//收集header头尾区间有没有paraTxBlocks
	txBlocks := &types.ParaTxDetails{}
	for i := job.start; i <= job.end; i++ {
		if job.paraTxBlocks[i] != nil {
			txBlocks.Items = append(txBlocks.Items, job.paraTxBlocks[i])
		}
	}

	if len(txBlocks.Items) > 0 {
		for _, tx := range txBlocks.Items {
			// 1. 校验平行链交易的区块头hash 和之前读取的主链头对应高度的块hash
			if !bytes.Equal(tx.Header.Hash, headMap[tx.Header.Height].Header.Hash) {
				plog.Error("jumpDldClient.process verifyhash", "height", tx.Header.Height,
					"txHash", common.ToHex(tx.Header.Hash), "headerHash", common.ToHex(headMap[tx.Header.Height].Header.Hash))
				atomic.StoreInt32(&j.downFail, 1)
				return
			}
			// 2. 校验交易merkle根和之前读的主链头的交易rootHash
			if j.paraClient.GetAPI().GetConfig().IsFork(tx.Header.Height, "ForkRootHash") {
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
	err := j.paraClient.procLocalAddBlocks(job.headers)
	if err != nil {
		atomic.StoreInt32(&j.downFail, 1)
		plog.Error("jumpDldClient.process procLocalAddBlocks", "start", job.start, "end", job.end, "err", err)
	}

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

func (j *jumpDldClient) getParaTxsBlocks(blocksList []*types.BlockInfo, title string) (map[int64]*types.ParaTxDetail, error) {
	var hlist []int64
	for _, h := range blocksList {
		hlist = append(hlist, h.Height)
	}

	blocks, err := j.fetchHeightListBlocks(hlist, title)
	if err != nil {
		plog.Error("jumpDld.getParaTxsBlocks", "start", hlist[0], "end", hlist[len(hlist)-1], "title", title)
		return nil, err
	}

	err = verifyBlockHash(blocksList, blocks.Items)
	if err != nil {
		plog.Error("jumpDld.getParaTxsBlocks verifyTx", "start", hlist[0], "end", hlist[len(hlist)-1], "title", title)
		return nil, err
	}

	blocksMap := make(map[int64]*types.ParaTxDetail)
	for _, b := range blocks.Items {
		blocksMap[b.Header.Height] = b
	}
	return blocksMap, nil
}

func (j *jumpDldClient) getHeaders(start, end int64) (*types.ParaTxDetails, error) {
	blocks := &types.ReqBlocks{Start: start, End: end}
	headers, err := j.paraClient.GetBlockHeaders(blocks)
	if err != nil {
		plog.Error("jumpDld.getHeaders", "start", start, "end", end, "error", err)
		return nil, err
	}
	plog.Debug("jumpDld.getHeaders", "start", start, "end", end)
	paraTxHeaders := &types.ParaTxDetails{}
	for _, header := range headers.Items {
		paraTxHeaders.Items = append(paraTxHeaders.Items, &types.ParaTxDetail{Type: types.AddBlock, Header: header})
	}
	return paraTxHeaders, nil
}

func (j *jumpDldClient) procParaTxHeaders(startHeight, endHeight int64, paraBlocks map[int64]*types.ParaTxDetail, jobCh chan *paraTxBlocksJob) error {
	for s := startHeight; s <= endHeight; s += types.MaxBlockCountPerTime {
		end := s + types.MaxBlockCountPerTime - 1
		if end > endHeight {
			end = endHeight
		}
		headers, err := j.getHeaders(s, end)
		if err != nil {
			plog.Error("jumpDld.procParaTxHeaders", "start", startHeight, "end", endHeight, "err", err)
			return err
		}
		//每1000个header同步一次，这样可以更快更小粒度的使同步层获取区块执行
		job := &paraTxBlocksJob{start: s, end: end, paraTxBlocks: paraBlocks, headers: headers}
		jobCh <- job

		if atomic.LoadInt32(&j.downFail) != 0 || j.paraClient.isCancel() {
			return errors.New("verify fail or main thread cancel")
		}
	}
	return nil
}

//每1000header执行一次比全部获取出来更有效率，可以和同步层更好并行处理，节约时间，1000paraTxBlocks花时间很少，相比headers获取，串行获取时间可以忽略
func (j *jumpDldClient) getParaTxs(startHeight, endHeight int64, heights []*types.BlockInfo, jobCh chan *paraTxBlocksJob) error {
	title := j.paraClient.GetAPI().GetConfig().GetTitle()
	heightsRows := splitHeights2Rows(heights, int(types.MaxBlockCountPerTime))

	for i, row := range heightsRows {
		//获取每一排1000个paraTxBlocks
		paraBlocks, err := j.getParaTxsBlocks(row, title)
		if err != nil {
			plog.Error("jumpDld.getParaTxsBlocks", "err", err)
			return err
		}
		//根据1000个paraTxBlocks的头尾高度获取header的头尾高度，header的高度要包含paraTxBlocks高度
		headerStart, headerEnd := getHeaderStartEndRange(startHeight, endHeight, heightsRows, i)
		plog.Info("jumpDld.getParaTxs", "headerStart", headerStart, "headerEnd", headerEnd, "i", i)
		err = j.procParaTxHeaders(headerStart, headerEnd, paraBlocks, jobCh)
		if err != nil {
			plog.Error("jumpDld.procParaTxHeaders", "err", err)
			return err
		}

		if atomic.LoadInt32(&j.downFail) != 0 || j.paraClient.isCancel() {
			plog.Error("jumpDld.downFail", "downfail", atomic.LoadInt32(&j.downFail))
			return errors.New("verify fail or main thread cancel")
		}
	}

	return nil
}

//Jump Download 是选择有平行链交易的区块跳跃下载的功能，分为三个步骤：
//0. 只获取当前主链高度1w高度前的区块，默认没有分叉，都是addType　block
//1. 获取所有平行链交易的高度列表，大概5s以内
//2. 按高度列表获取含平行链交易的主链区块每次获取一段1000高度，并获取相关范围的主链headers，一起执行，单独获取headers和同步处理不能并行
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

	//1. 获取有平行链交易的块高度列表
	t1 := types.Now()
	heights, err := j.getParaHeightList(startHeight, endHeight)
	if err != nil {
		plog.Error("JumpDld.getParaHeightList", "err", err)
	}
	if len(heights) == 0 {
		plog.Error("JumpDld.getParaHeightList　no height found")
		return
	}
	plog.Info("tryJumpDownload.getParaHeightList", "time", types.Since(t1))

	//2. 按有平行链交易的高度列表获取平行链块
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
	plog.Info("tryJumpDownload.getParaTxs　done", "time", types.Since(t1))
}
