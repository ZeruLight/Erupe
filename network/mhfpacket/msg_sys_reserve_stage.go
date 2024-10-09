package mhfpacket

import (
	"errors"
	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysReserveStage represents the MSG_SYS_RESERVE_STAGE
type MsgSysReserveStage struct {
	AckHandle uint32
	Ready     uint8  // Bitfield but hex (0x11 or 0x01)
	StageID   string // NULL terminated string.
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserveStage) Opcode() network.PacketID {
	return network.MSG_SYS_RESERVE_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysReserveStage) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Ready = bf.ReadUint8()
	_ = bf.ReadUint8() // StageID length
	m.StageID = string(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserveStage) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
