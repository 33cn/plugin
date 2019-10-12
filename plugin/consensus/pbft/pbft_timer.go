// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pbft

import "time"

// timerStart 用于启动定时器
type timerStart struct {
	reason   string        //启动的原因
	duration time.Duration // 计时多久
}

// Timer 是定时器结构
type Timer struct {
	startChan  chan *timerStart
	stopChan   chan struct{}
	closeChan  chan struct{}
	timerChan  <-chan time.Time
	pending    *timerStart
	stopSignal bool
}

func (t *Timer) init() {
	t.startChan = make(chan *timerStart)
	t.stopChan = make(chan struct{})
	t.closeChan = make(chan struct{})
	t.timerChan = make(<-chan time.Time)
	t.stopSignal = false
}

// Reset 用于重置，然后开启定时器
func (t *Timer) Reset(resetReason string, timeout time.Duration) {
	t.startChan <- &timerStart{
		reason:   resetReason,
		duration: timeout,
	}
}

// Stop 用于停止目前的所有定时
func (t *Timer) Stop() {
	t.stopChan <- struct{}{}
}

// Close 用于关闭定时器
func (t *Timer) Close() {
	t.closeChan <- struct{}{}
}
