package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfLoadPartner represents the MSG_MHF_LOAD_PARTNER
type MsgMhfLoadPartner struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadPartner) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_PARTNER
}

// Parse parses the packet from binary
func (m *MsgMhfLoadPartner) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadPartner) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
