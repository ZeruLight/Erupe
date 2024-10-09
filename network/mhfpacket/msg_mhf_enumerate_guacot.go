package mhfpacket

import (
	"errors"
	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEnumerateGuacot represents the MSG_MHF_ENUMERATE_GUACOT
type MsgMhfEnumerateGuacot struct {
	AckHandle uint32
	Unk0      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateGuacot) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_GUACOT
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateGuacot) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	bf.ReadUint16() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateGuacot) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
