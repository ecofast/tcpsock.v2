package protocol

import (
	"encoding/binary"
)

const (
	SizeOfPacketHeadLen = 2
	SizeOfPacketHeadCmd = 2
	SizeOfPacketHead    = SizeOfPacketHeadLen + SizeOfPacketHeadCmd

	SizeOfMsgHeadProtoID = 2
	SizeOfMsgHeadParam   = 2
	SizeOfMsgHead        = SizeOfMsgHeadProtoID + SizeOfMsgHeadParam
)

/*
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃	PacketHead	┃	MsgHead	┃	MsgBody	┃	PacketHead	┃	MsgHead	┃	MsgBody	┃ ...... ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
*/

type PacketHead struct {
	Len uint16
	Cmd uint16
}

type MsgHead struct {
	ProtoID uint16
	Param   uint16
}

type MsgBody struct {
	Content []byte
}

type NetPacket struct {
	PacketHead
	MsgHead
	MsgBody
}

func (self *NetPacket) Bytes() []byte {
	buf := make([]byte, SizeOfPacketHead+SizeOfMsgHead+len(self.Content))
	binary.LittleEndian.PutUint16(buf[:SizeOfPacketHeadLen], self.PacketHead.Len)
	binary.LittleEndian.PutUint16(buf[SizeOfPacketHeadLen:SizeOfPacketHeadLen+SizeOfPacketHeadCmd], self.PacketHead.Cmd)
	binary.LittleEndian.PutUint16(buf[SizeOfPacketHead:SizeOfPacketHead+SizeOfMsgHeadProtoID], self.MsgHead.ProtoID)
	binary.LittleEndian.PutUint16(buf[SizeOfPacketHead+SizeOfMsgHeadProtoID:SizeOfPacketHead+SizeOfMsgHeadProtoID+SizeOfMsgHeadParam], self.MsgHead.Param)
	copy(buf[SizeOfPacketHead+SizeOfMsgHead:], self.MsgBody.Content[:])
	return buf
}

func NewPacket(cmd uint16, protoID, param uint16, content []byte) *NetPacket {
	if len(content) > 1<<16-SizeOfPacketHead-SizeOfMsgHead {
		return nil
	}
	p := &NetPacket{
		PacketHead{
			Len: uint16(SizeOfMsgHead + len(content)),
			Cmd: cmd,
		},
		MsgHead{
			ProtoID: protoID,
			Param:   param,
		},
		MsgBody{
			Content: nil,
		},
	}
	if content != nil && len(content) > 0 {
		p.MsgBody.Content = make([]byte, len(content))
		copy(p.MsgBody.Content, content)
	}
	return p
}

const (
	PT_NORMAL = 1

	cCmSmDif = 32767

	CM_PING = 1
	SM_PING = cCmSmDif + CM_PING

	CM_IDENTITY = 2
	SM_IDENTITY = cCmSmDif + CM_IDENTITY

	CM_REQROOMLIST = 3
	SM_REQROOMLIST = cCmSmDif + CM_REQROOMLIST

	CM_ENTERROOM = 4
	SM_ENTERROOM = cCmSmDif + CM_ENTERROOM

	CM_EXITROOM = 5
	SM_EXITROOM = cCmSmDif + CM_EXITROOM

	SM_NOTIFY = cCmSmDif + 6

	CM_CHAT = 7
	SM_CHAT = cCmSmDif + CM_CHAT
)

const (
	SizeOfUserName = 8
)
