package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysWaitStageBinary represents the MSG_SYS_WAIT_STAGE_BINARY
type MsgSysWaitStageBinary struct {
	AckHandle     uint32
	BinaryType0   uint8
	BinaryType1   uint8
	Unk0          uint32 // Hardcoded 0
	StageIDLength uint8
	StageID       string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysWaitStageBinary) Opcode() network.PacketID {
	return network.MSG_SYS_WAIT_STAGE_BINARY
}

// Parse parses the packet from binary
func (m *MsgSysWaitStageBinary) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.BinaryType0 = bf.ReadUint8()
	m.BinaryType1 = bf.ReadUint8()
	m.Unk0 = bf.ReadUint32()
	m.StageIDLength = bf.ReadUint8()
	m.StageID = string(bf.ReadBytes(uint(m.StageIDLength)))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysWaitStageBinary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
