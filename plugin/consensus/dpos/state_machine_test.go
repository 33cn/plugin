// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
	"github.com/stretchr/testify/assert"
)

func init() {
	setParams(3, 3, 6)
}
func setParams(delegateNum int64, blockInterval int64, continueBlockNum int64) {
	dposDelegateNum = delegateNum           //委托节点个数，从配置读取，以后可以根据投票结果来定
	dposBlockInterval = blockInterval       //出块间隔，当前按3s
	dposContinueBlockNum = continueBlockNum //一个委托节点当选后，一次性持续出块数量
	dposCycle = dposDelegateNum * dposBlockInterval * dposContinueBlockNum
	dposPeriod = dposBlockInterval * dposContinueBlockNum
}

func printTask(now int64, task *Task) {
	fmt.Printf("now:%v|cycleStart:%v|cycleStop:%v|periodStart:%v|periodStop:%v|blockStart:%v|blockStop:%v|nodeId:%v\n",
		now,
		task.CycleStart,
		task.CycleStop,
		task.PeriodStart,
		task.PeriodStop,
		task.BlockStart,
		task.BlockStop,
		task.NodeID)
}
func assertTask(task *Task, t *testing.T) {
	assert.Equal(t, true, task.NodeID >= 0 && task.NodeID < dposDelegateNum)
	assert.Equal(t, true, task.CycleStart <= task.PeriodStart && task.PeriodStart <= task.BlockStart && task.BlockStop <= task.PeriodStop && task.PeriodStop <= task.CycleStop)

}
func TestDecideTaskByTime(t *testing.T) {

	now := time.Now().Unix()
	task := DecideTaskByTime(now)
	printTask(now, &task)
	assertTask(&task, t)

	setParams(2, 1, 6)
	now = time.Now().Unix()
	task = DecideTaskByTime(now)
	printTask(now, &task)
	assertTask(&task, t)

	setParams(21, 1, 12)
	now = time.Now().Unix()
	task = DecideTaskByTime(now)
	printTask(now, &task)
	assertTask(&task, t)

	setParams(21, 2, 12)
	now = time.Now().Unix()
	task = DecideTaskByTime(now)
	printTask(now, &task)
	assertTask(&task, t)

	setParams(2, 3, 12)

	/*
		for i := 0; i < 120; i++ {
			now = time.Now().Unix()
			task = DecideTaskByTime(now)
			printTask(now, &task)
			assertTask(&task, t)
			time.Sleep(time.Second * 1)
		}
	*/
}
