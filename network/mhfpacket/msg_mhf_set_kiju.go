package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfSetKiju represents the MSG_MHF_SET_KIJU
type MsgMhfSetKiju struct {
	AckHandle uint32
	Unk1      uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetKiju) Opcode() network.PacketID {
	return network.MSG_MHF_SET_KIJU
}

// Parse parses the packet from binary
func (m *MsgMhfSetKiju) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk1 = bf.ReadUint16()
	return nil
	//panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetKiju) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
