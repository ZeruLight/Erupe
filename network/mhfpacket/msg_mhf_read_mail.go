package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadMail represents the MSG_MHF_READ_MAIL
type MsgMhfReadMail struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadMail) Opcode() network.PacketID {
	return network.MSG_MHF_READ_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfReadMail) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadMail) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
