package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadPartner represents the MSG_MHF_LOAD_PARTNER
type MsgMhfLoadPartner struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadPartner) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_PARTNER
}

// Parse parses the packet from binary
func (m *MsgMhfLoadPartner) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadPartner) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}