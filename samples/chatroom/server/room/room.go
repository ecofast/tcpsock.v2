package room

import (
	. "tcpsock.v2/samples/chatroom/protocol"
	"tcpsock.v2/samples/chatroom/server/player"
)

const (
	MaxUserPerRoom = 200
)

var (
	enterRoomHint = []byte(" 已进入房间")
	leaveRoomHint = []byte(" 已离开房间")
)

type Room struct {
	idx     uint8
	players [MaxUserPerRoom]player.Player
}

func New(idx uint8) *Room {
	return &Room{
		idx: idx,
	}
}

func (self *Room) IsFull() bool {
	for i := 0; i < MaxUserPerRoom; i++ {
		if self.players[i] == nil {
			return false
		}
	}
	return true
}

func (self *Room) Enter(p player.Player) int {
	name := p.Name()
	sz := len(name) + len(enterRoomHint)
	buf := make([]byte, sz)
	copy(buf[:SizeOfUserName], name)
	copy(buf[SizeOfUserName:], enterRoomHint)
	for i := range self.players {
		if self.players[i] == nil {
			p.EnterRoom(self.idx, uint8(i))
			self.broadcast(NewPacket(PT_NORMAL, SM_NOTIFY, 0, buf).Bytes())
			self.players[i] = p
			return i
		}
	}
	return -1
}

func (self *Room) Leave(p player.Player) bool {
	id := p.SeatID()
	if id >= 0 && id < MaxUserPerRoom {
		name := p.Name()
		sz := len(name) + len(leaveRoomHint)
		buf := make([]byte, sz)
		copy(buf[:SizeOfUserName], name)
		copy(buf[SizeOfUserName:], leaveRoomHint)
		self.players[id] = nil
		self.broadcast(NewPacket(PT_NORMAL, SM_NOTIFY, 0, buf).Bytes())
		return true
	}
	return false
}

func (self *Room) Chat(p player.Player, s []byte) {
	buf := make([]byte, SizeOfUserName+len(s))
	copy(buf[:SizeOfUserName], p.Name())
	copy(buf[SizeOfUserName:], s)
	self.broadcast(NewPacket(PT_NORMAL, SM_CHAT, 0, buf).Bytes())
}

func (self *Room) broadcast(b []byte) {
	for i := range self.players {
		if self.players[i] != nil {
			self.players[i].Write(b)
		}
	}
}
