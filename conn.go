// Copyright (C) 2018 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

package tcpsock

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	RecvBufLenMax = 4 * 1024
	SendBufLenMax = 24 * 1024
	TcpBufLenMax  = 4 * 1024
)

type queueNode struct {
	buf  []byte
	next *queueNode
}

type TcpConn struct {
	id         uint64
	owner      *tcpSock
	conn       net.Conn
	closeChan  chan struct{}
	closeOnce  sync.Once
	closedFlag int32
	onClose    OnTcpDisconnect

	mutex         sync.Mutex
	firstSendNode *queueNode
	lastSendNode  *queueNode

	sendBuf [SendBufLenMax]byte
	bufLen  int

	onRead func(p []byte) (n int, err error)
}

func newTcpConn(id uint64, owner *tcpSock, conn net.Conn, onClose OnTcpDisconnect) *TcpConn {
	return &TcpConn{
		id:        id,
		owner:     owner,
		conn:      conn,
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

	node := &queueNode{}
	node.buf = make([]byte, cnt)
	copy(node.buf, b)
	self.mutex.Lock()
	if self.lastSendNode != nil {
		self.lastSendNode.next = node
	}
	if self.firstSendNode == nil {
		self.firstSendNode = node
	}
	self.lastSendNode = node
	self.mutex.Unlock()
	return cnt, nil
}

func (self *TcpConn) Close() error {
	self.closeOnce.Do(func() {
		atomic.StoreInt32(&self.closedFlag, 1)
		close(self.closeChan)
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
	self.mutex.Lock()
	defer self.mutex.Unlock()
	self.firstSendNode = nil
	self.lastSendNode = nil
	self.bufLen = 0
}

func (self *TcpConn) getSendBuf() []byte {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	for self.firstSendNode != nil {
		node := self.firstSendNode
		self.firstSendNode = node.next
		copy(self.sendBuf[self.bufLen:], node.buf)
		self.bufLen += len(node.buf)
		if self.bufLen >= TcpBufLenMax {
			break
		}
	}

	if self.firstSendNode == nil {
		self.lastSendNode = nil
	}

	l := self.bufLen
	if l > TcpBufLenMax {
		l = TcpBufLenMax
	}
	if l > 0 {
		b := make([]byte, l)
		copy(b, self.sendBuf[:l])
		self.bufLen -= l
		if self.bufLen > 0 {
			copy(self.sendBuf[0:], self.sendBuf[l:])
		}
		if self.bufLen < 0 {
			self.bufLen = 0
		}
		return b
	}

	return nil
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

		default:
			if b := self.getSendBuf(); b != nil {
				if n, err := self.conn.Write(b); err != nil || n != len(b) {
					// to do
					//
					return
				}
			} else {
				time.Sleep(5 * time.Millisecond)
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
				// to do
				//
				return
			}
		}
	}
}
