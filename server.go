// Copyright (C) 2018 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

package tcpsock

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"

	. "github.com/ecofast/rtl/sysutils"
)

const (
	numOfConnInit = 100
	NumOfConnMax  = 10000
)

type OnCheckIP = func(ip net.Addr) bool

type TcpServer struct {
	listener *net.TCPListener
	*tcpSock
	autoIncID uint64
	count     uint32
	mutex     sync.RWMutex
	sessions  map[uint64]TcpSession
	onCheckIP OnCheckIP
}

func NewTcpServer(addr string, onConnect OnTcpConnect, onDisconnect OnTcpDisconnect, onCheckIP OnCheckIP) *TcpServer {
	if addr == "" {
		panic(errors.New("invalid param of addr for NewTcpServer"))
	}
	if onConnect == nil {
		panic(errors.New("invalid param of onConnect for NewTcpServer"))
	}
	if onDisconnect == nil {
		panic(errors.New("invalid param of onDisconnect for NewTcpServer"))
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	CheckError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	CheckError(err)

	return &TcpServer{
		listener: listener,
		tcpSock: &tcpSock{
			exitChan:     make(chan struct{}),
			waitGroup:    &sync.WaitGroup{},
			onConnect:    onConnect,
			onDisconnect: onDisconnect,
		},
		sessions:  make(map[uint64]TcpSession, numOfConnInit),
		onCheckIP: onCheckIP,
	}
}

func (self *TcpServer) Serve() {
	go self.run()
}

func (self *TcpServer) run() {
	self.waitGroup.Add(1)
	defer func() {
		self.listener.Close()
		self.waitGroup.Done()
	}()

	for {
		select {
		case <-self.exitChan:
			return
		default:
		}

		conn, err := self.listener.AcceptTCP()
		if err != nil {
			continue
		}

		if !self.checkConn(conn.RemoteAddr()) {
			conn.Close()
			continue
		}

		atomic.AddUint32(&self.count, 1)
		self.waitGroup.Add(1)
		go func() {
			c := newTcpConn(atomic.AddUint64(&self.autoIncID, 1), self.tcpSock, conn, self.connClose)
			session := self.onConnect(c)
			if session != nil {
				c.onRead = session.Read
				self.addSession(c.ID(), session)
			}
			c.run()
			self.waitGroup.Done()
		}()
	}
}

func (self *TcpServer) Close() {
	self.listener.Close()
	close(self.exitChan)
	self.waitGroup.Wait()
}

func (self *TcpServer) Count() uint32 {
	return atomic.LoadUint32(&self.count)
}

func (self *TcpServer) Iterate(fn OnTcpIterate) {
	self.mutex.RLock()
	defer self.mutex.RUnlock()
	for id, session := range self.sessions {
		fn(id, session)
	}
}

func (self *TcpServer) Send(id uint64, b []byte) {
	if len(b) == 0 {
		return
	}
	
	self.mutex.RLock()
	if v, ok := self.sessions[id]; ok {
		v.Write(b)
	}
	self.mutex.RUnlock()
}

func (self *TcpServer) Kick(id uint64) {
	self.mutex.Lock()
	delete(self.sessions, id)
	self.mutex.Unlock()
}

func (self *TcpServer) GetSession(id uint64) TcpSession {
	var ret TcpSession
	self.mutex.RLock()
	if v, ok := self.sessions[id]; ok {
		ret = v
	}
	self.mutex.RUnlock()
	return ret
}

func (self *TcpServer) checkConn(ip net.Addr) bool {
	if self.Count() >= NumOfConnMax {
		return false
	}

	if (self.onCheckIP != nil) && (!self.onCheckIP(ip)) {
		return false
	}

	return true
}

func (self *TcpServer) connClose(conn *TcpConn) {
	atomic.AddUint32(&self.count, ^uint32(0))
	if self.onDisconnect != nil {
		self.onDisconnect(conn)
	}
	self.delSession(conn.ID())
}

func (self *TcpServer) addSession(id uint64, session TcpSession) {
	self.mutex.Lock()
	self.sessions[id] = session
	self.mutex.Unlock()
}

func (self *TcpServer) delSession(id uint64) {
	self.mutex.Lock()
	delete(self.sessions, id)
	self.mutex.Unlock()
}
