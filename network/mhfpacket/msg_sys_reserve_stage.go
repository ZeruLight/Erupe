package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserveStage represents the MSG_SYS_RESERVE_STAGE
type MsgSysReserveStage struct {
	AckHandle     uint32
	Unk0          uint8 // Made with: `16 * x | 1;`, unknown `x` values.
	StageIDLength uint8
	StageID       string // NULL terminated string.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserveStage) Opcode() network.PacketID {
	return network.MSG_SYS_RESERVE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysReserveStage) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.StageIDLength = bf.ReadUint8()
	m.StageID = string(bf.ReadBytes(uint(m.StageIDLength)))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserveStage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
