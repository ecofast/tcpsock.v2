package clientsock

import (
	"errors"
	"fmt"

	. "github.com/ecofast/rtl/sysutils"
	"tcpsock.v2"
	. "tcpsock.v2/samples/chatroom/protocol"
	. "tcpsock.v2/samples/chatroom/server/msgnode"
)

type FnWrite = func(b []byte) (n int, err error)
type FnClose = func() error

type ClientSock struct {
	sockHandle uint64
	onWrite    FnWrite
	onClose    FnClose
	cliChan    chan<- *MsgNode
	userName   [SizeOfUserName]byte
	roomID     uint8
	seatID     uint8
	recvBuf    []byte
	recvBufLen int
}

func New(handle uint64, fnWrite FnWrite, fnClose FnClose, cliChan chan<- *MsgNode) *ClientSock {
	return &ClientSock{
		sockHandle: handle,
		onWrite:    fnWrite,
		onClose:    fnClose,
		cliChan:    cliChan,
		roomID:     0xFF,
		seatID:     0xFF,
	}
}

func (self *ClientSock) SockHandle() uint64 {
	return self.sockHandle
}

func (self *ClientSock) Name() []byte {
	return self.userName[:]
}

func (self *ClientSock) EnterRoom(roomID, seatID uint8) {
	self.roomID = roomID
	self.seatID = seatID
}

func (self *ClientSock) RoomID() uint8 {
	return self.roomID
}

func (self *ClientSock) SeatID() uint8 {
	return self.seatID
}

func (self *ClientSock) Read(b []byte) (n int, err error) {
	count := len(b)
	if count+self.recvBufLen > tcpsock.RecvBufLenMax {
		return 0, errors.New("invalid data")
	}

	self.recvBuf = append(self.recvBuf, b[0:count]...)
	self.recvBufLen += count
	offsize := 0
	offset := 0
	var head PacketHead
	for self.recvBufLen-offsize > SizeOfPacketHead {
		offset = 0
		head.Len = uint16(uint16(self.recvBuf[offsize+1])<<8 | uint16(self.recvBuf[offsize+0]))
		pkglen := int(SizeOfPacketHead + head.Len)
		if pkglen >= tcpsock.RecvBufLenMax {
			offsize = self.recvBufLen
			break
		}
		if offsize+pkglen > self.recvBufLen {
			break
		}
		offset += SizeOfPacketHeadLen
		head.Cmd = uint16(uint16(self.recvBuf[offsize+offset+1])<<8 | uint16(self.recvBuf[offsize+offset+0]))
		switch head.Cmd {
		case PT_NORMAL:
			offset += SizeOfPacketHeadCmd
			self.process(self.recvBuf[offsize+offset : offsize+offset+int(head.Len)])
		default:
			//
		}
		offsize += pkglen
	}

	self.recvBufLen -= offsize
	if self.recvBufLen > 0 {
		self.recvBuf = self.recvBuf[offsize : offsize+self.recvBufLen]
	} else {
		self.recvBuf = nil
	}
	return len(b), nil
}

func (self *ClientSock) process(b []byte) {
	switch BytesToUInt16(b[:SizeOfMsgHeadProtoID]) {
	case CM_PING:
		self.Write(NewPacket(PT_NORMAL, SM_PING, 0, nil).Bytes())
	case CM_IDENTITY:
		copy(self.userName[:], b[SizeOfMsgHead:SizeOfMsgHead+SizeOfUserName])
		self.Write(NewPacket(PT_NORMAL, SM_IDENTITY, 0, nil).Bytes())
	case CM_REQROOMLIST:
		self.cliChan <- &MsgNode{
			Owner:   self,
			ProtoID: CM_REQROOMLIST,
		}
	case CM_ENTERROOM:
		if self.roomID != 0xFF {
			self.Write(NewPacket(PT_NORMAL, SM_ENTERROOM, 0, nil).Bytes())
			return
		}
		self.cliChan <- &MsgNode{
			Owner:   self,
			ProtoID: CM_ENTERROOM,
			Param:   BytesToUInt16(b[SizeOfMsgHeadProtoID : SizeOfMsgHeadProtoID+SizeOfMsgHeadParam]),
		}
	case CM_EXITROOM:
		self.cliChan <- &MsgNode{
			Owner:   self,
			ProtoID: CM_EXITROOM,
		}
	case CM_CHAT:
		self.cliChan <- &MsgNode{
			Owner:   self,
			ProtoID: CM_CHAT,
			Buf:     b[SizeOfMsgHead:],
		}
	default:
		fmt.Println("?????")
		self.Close()
	}
}

func (self *ClientSock) Write(b []byte) (n int, err error) {
	if self.onWrite != nil {
		return self.onWrite(b)
	}
	return len(b), nil
}

func (self *ClientSock) Close() error {
	if self.onClose != nil {
		return self.onClose()
	}
	return nil
}
