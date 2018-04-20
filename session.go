// Copyright (C) 2018 ecofast(胡光耀). All rights reserved.
// Use of this source code is governed by a BSD-style license.

package tcpsock

import (
	"io"
)

// to do
//
type TcpSession interface {
	SockHandle() uint64
	io.ReadWriteCloser
}
