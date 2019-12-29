// Copyright (C) 2018 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

package tcpsock

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

const (
	RecvBufLenMax = 16 * 1024
	SendBufLenMax = 24 * 1024
	TcpBufLenMax  = 16 * 1024
)

type TcpConn struct {
	id         uint64
	owner      *tcpSock
	conn       net.Conn
	bufChan    chan []byte
	closeChan  chan struct{}
	closeOnce  sync.Once
	closedFlag int32
	onClose    OnTcpDisconnect
	onRead     func(p []byte) (n int, err error)
}

func newTcpConn(id uint64, owner *tcpSock, conn net.Conn, onClose OnTcpDisconnect) *TcpConn {
	return &TcpConn{
		id:        id,
		owner:     owner,
		conn:      conn,
		bufChan:   make(chan []byte, 20),
		closeChan: make(chan struct{}),
		onClose:   onClose,
	}
}

func (self *TcpConn) ID() uint64 {
	return self.id
}

func (self *TcpConn) Write(b []byte) (n int, err error) {
	if self.closed() {
		return 0, errors.New("connection closed")
	}

	cnt := len(b)
	if cnt == 0 || cnt > SendBufLenMax {
		return 0, errors.New("invalid data")
	}

	self.bufChan <- b
	return cnt, nil
}

func (self *TcpConn) Close() error {
	self.closeOnce.Do(func() {
		atomic.StoreInt32(&self.closedFlag, 1)
		close(self.closeChan)
		close(self.bufChan)
		self.conn.Close()
		if self.onClose != nil {
			self.onClose(self)
		}
		self.clear()
	})
	return nil
}

func (self *TcpConn) closed() bool {
	return atomic.LoadInt32(&self.closedFlag) == 1
}

func (self *TcpConn) RawConn() net.Conn {
	return self.conn
}

func (self *TcpConn) run() {
	startGoroutine(self.send, self.owner.waitGroup)
	startGoroutine(self.recv, self.owner.waitGroup)
}

func startGoroutine(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()
}

func (self *TcpConn) clear() {
	// place holder
}

func (self *TcpConn) send() {
	defer func() {
		recover()
		self.Close()
	}()

	for {
		if self.closed() {
			return
		}

		select {
		case <-self.owner.exitChan:
			return
		case <-self.closeChan:
			return
		case b := <-self.bufChan:
			if n, err := self.conn.Write(b); err != nil || n != len(b) {
				return
			}
		}
	}
}

func (self *TcpConn) recv() {
	defer func() {
		recover()
		self.Close()
	}()

	buf := make([]byte, RecvBufLenMax)
	for {
		select {
		case <-self.owner.exitChan:
			return
		case <-self.closeChan:
			return
		default:
		}

		cnt, err := self.conn.Read(buf)
		if err != nil || cnt == 0 {
			return
		}
		if self.onRead != nil {
			if n, err := self.onRead(buf[:cnt]); n != cnt || err != nil {
				return
			}
		}
	}
}
