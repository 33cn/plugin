// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package raft

import (
	"errors"
	"net"
	"time"
)

// 设置TCP keep-alive超时，接收stopc
type stoppableListener struct {
	*net.TCPListener
	stopc <-chan struct{}
}

// 监听tcp连接
func newStoppableListener(addr string, stopc <-chan struct{}) (*stoppableListener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &stoppableListener{ln.(*net.TCPListener), stopc}, nil
}

func (ln stoppableListener) Accept() (c net.Conn, err error) {
	connc := make(chan *net.TCPConn, 1)
	errc := make(chan error, 1)
	go func() {
		tc, err := ln.AcceptTCP()
		if err != nil {
			errc <- err
			return
		}
		connc <- tc
	}()
	select {
	case <-ln.stopc:
		return nil, errors.New("server stopped")
	case err := <-errc:
		return nil, err
	case tc := <-connc:
		err := tc.SetKeepAlive(true)
		if err != nil {
			return tc, err
		}
		err = tc.SetKeepAlivePeriod(3 * time.Minute)
		if err != nil {
			return tc, err
		}
		return tc, nil
	}
}
