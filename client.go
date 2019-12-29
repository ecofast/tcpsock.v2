// Copyright (C) 2018 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

package tcpsock

import (
	"errors"
	"net"
	"sync"
	"time"
)

const (
	TcpDialTimeoutInSecs = 2
)

type TcpClient struct {
	svrAddr string
	*tcpSock
	*TcpConn
}

func NewTcpClient(svrAddr string, onConnect OnTcpConnect, onDisconnect OnTcpDisconnect) *TcpClient {
	if svrAddr == "" {
		panic(errors.New("invalid param of addr for NewTcpServer"))
	}
	if onConnect == nil {
		panic(errors.New("invalid param of onConnect for NewTcpServer"))
	}
	if onDisconnect == nil {
		panic(errors.New("invalid param of onDisconnect for NewTcpServer"))
	}

	return &TcpClient{
		svrAddr: svrAddr,
		tcpSock: &tcpSock{
			exitChan:     make(chan struct{}),
			waitGroup:    &sync.WaitGroup{},
			onConnect:    onConnect,
			onDisconnect: onDisconnect,
		},
	}
}

func (self *TcpClient) Open() {
	if conn, err := net.DialTimeout("tcp", self.svrAddr, TcpDialTimeoutInSecs*time.Second); err == nil {
		self.waitGroup.Add(1)
		go func() {
			c := newTcpConn(0, self.tcpSock, conn, self.connClose)
			session := self.onConnect(c)
			if session != nil {
				c.onRead = session.Read
			}
			self.TcpConn = c
			c.run()
			self.waitGroup.Done()
		}()
		time.Sleep(100 * time.Millisecond)
	}
}

func (self *TcpClient) Close() error {
	close(self.exitChan)
	self.waitGroup.Wait()
	return nil
}

func (self *TcpClient) connClose(conn *TcpConn) {
	if self.onDisconnect != nil {
		self.onDisconnect(conn)
	}
}
