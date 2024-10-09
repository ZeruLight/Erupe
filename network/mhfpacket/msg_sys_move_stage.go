package mhfpacket

import (
	"erupe-ce/network"
	"erupe-ce/utils/bfutil"
	"erupe-ce/utils/byteframe"
)

// MsgSysMoveStage represents the MSG_SYS_MOVE_STAGE
type MsgSysMoveStage struct {
	AckHandle     uint32
	UnkBool       uint8
	StageIDLength uint8
	StageID       string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysMoveStage) Opcode() network.PacketID {
	return network.MSG_SYS_MOVE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysMoveStage) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.UnkBool = bf.ReadUint8()
	stageIDLength := bf.ReadUint8()
	m.StageID = string(bfutil.UpToNull(bf.ReadBytes(uint(stageIDLength))))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysMoveStage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
