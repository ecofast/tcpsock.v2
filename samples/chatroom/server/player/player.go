package player

/*
import (
	"tcpsock.v2/samples/chatroom/protocol"
)
*/

type Player interface {
	SockHandle() uint64
	Write(b []byte) (int, error)
	Name() []byte
	EnterRoom(roomID, seatID uint8)
	RoomID() uint8
	SeatID() uint8
}
