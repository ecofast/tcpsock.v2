package msgnode

import (
	"tcpsock.v2/samples/chatroom/server/player"
)

type MsgNode struct {
	Owner   player.Player
	ProtoID uint16
	Param   uint16
	_       uint32
	Buf     []byte
	Next    *MsgNode
}
