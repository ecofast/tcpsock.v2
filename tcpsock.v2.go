// Copyright (C) 2018 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

// Package tcpsock provides easy to use interfaces for TCP I/O.
// It's designed especially for developing online games.
package tcpsock

import (
	"sync"
)

type OnTcpConnect = func(conn *TcpConn) TcpSession
type OnTcpDisconnect = func(conn *TcpConn)
type OnTcpError = func(conn *TcpConn, err error)
type OnTcpIterate = func(id uint64, session TcpSession)

type tcpSock struct {
	exitChan     chan struct{}
	waitGroup    *sync.WaitGroup
	onConnect    OnTcpConnect
	onDisconnect OnTcpDisconnect
}
