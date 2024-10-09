package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfReadMercenaryM represents the MSG_MHF_READ_MERCENARY_M
type MsgMhfReadMercenaryM struct {
	AckHandle uint32
	CharID    uint32
	MercID    uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadMercenaryM) Opcode() network.PacketID {
	return network.MSG_MHF_READ_MERCENARY_M
}

// Parse parses the packet from binary
func (m *MsgMhfReadMercenaryM) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.CharID = bf.ReadUint32()
	m.MercID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadMercenaryM) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
