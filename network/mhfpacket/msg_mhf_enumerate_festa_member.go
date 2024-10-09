package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEnumerateFestaMember represents the MSG_MHF_ENUMERATE_FESTA_MEMBER
type MsgMhfEnumerateFestaMember struct {
	AckHandle uint32
	FestaID   uint32
	GuildID   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateFestaMember) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_FESTA_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateFestaMember) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.FestaID = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	bf.ReadUint16() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateFestaMember) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
