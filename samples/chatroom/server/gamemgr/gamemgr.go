package gamemgr

import (
	"sync"
	"time"

	"tcpsock.v2/samples/chatroom/protocol"
	. "tcpsock.v2/samples/chatroom/server/msgnode"
	"tcpsock.v2/samples/chatroom/server/player"
	"tcpsock.v2/samples/chatroom/server/room"
)

const (
	RoomNumMax = 10
)

type GameMgr struct {
	rooms        [RoomNumMax]*room.Room
	mutex        sync.Mutex
	firstMsgNode *MsgNode
	lastMsgNode  *MsgNode
}

var (
	gameMgr *GameMgr
)

func Setup() {
	gameMgr = &GameMgr{}
	for i := 0; i < RoomNumMax; i++ {
		gameMgr.rooms[i] = room.New(uint8(i))
	}
}

func Run(exitChan chan struct{}, waitGroup *sync.WaitGroup, cliChan <-chan *MsgNode) {
	defer waitGroup.Done()

	go gameMgr.run()
	go func() {
		for node := range cliChan {
			gameMgr.addMsgNode(node)
		}
	}()

	<-exitChan
}

func (self *GameMgr) addMsgNode(node *MsgNode) {
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.lastMsgNode != nil {
		self.lastMsgNode.Next = node
	}
	if self.firstMsgNode == nil {
		self.firstMsgNode = node
	}
	self.lastMsgNode = node
}

func (self *GameMgr) getMsgNode() *MsgNode {
	var node *MsgNode
	self.mutex.Lock()
	defer self.mutex.Unlock()
	if self.firstMsgNode != nil {
		node = self.firstMsgNode
		self.firstMsgNode = node.Next
	}
	if self.firstMsgNode == nil {
		self.lastMsgNode = nil
	}
	return node
}

func (self *GameMgr) run() {
	for {
		node := self.getMsgNode()
		if node == nil {
			time.Sleep(5 * time.Millisecond)
			continue
		}
		self.process(node)
	}
}

func (self *GameMgr) process(node *MsgNode) {
	switch node.ProtoID {
	case protocol.CM_REQROOMLIST:
		self.processReqRoomList(node.Owner)
	case protocol.CM_ENTERROOM:
		self.processEnterRoom(node.Owner, uint8(node.Param))
	case protocol.CM_EXITROOM:
		self.processLeaveRoom(node.Owner)
	case protocol.CM_CHAT:
		self.processChat(node.Owner, node.Buf)
	}
}

func (self *GameMgr) processReqRoomList(p player.Player) {
	buf := make([]byte, len(self.rooms))
	for i := range self.rooms {
		buf[i] = byte(i)
	}
	p.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.SM_REQROOMLIST, 0, buf).Bytes())
}

func (self *GameMgr) processEnterRoom(p player.Player, reqID uint8) {
	if reqID >= 0 && reqID < uint8(len(self.rooms)) {
		if seatID := self.rooms[reqID].Enter(p); seatID >= 0 {
			p.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.SM_ENTERROOM, uint16(reqID)<<8|uint16(seatID), nil).Bytes())
			return
		}
	}
	p.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.SM_ENTERROOM, 0xFFFF, nil).Bytes())
}

func (self *GameMgr) processLeaveRoom(p player.Player) {
	roomID := p.RoomID()
	if roomID >= 0 && roomID < uint8(len(self.rooms)) {
		if self.rooms[roomID].Leave(p) {
			p.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.SM_EXITROOM, 0, nil).Bytes())
			return
		}
	}
	p.Write(protocol.NewPacket(protocol.PT_NORMAL, protocol.SM_EXITROOM, 0xFFFF, nil).Bytes())
}

func (self *GameMgr) processChat(p player.Player, s []byte) {
	roomID := p.RoomID()
	if roomID >= 0 && roomID < uint8(len(self.rooms)) {
		self.rooms[roomID].Chat(p, s)
	}
}
