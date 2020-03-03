package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfListMail represents the MSG_MHF_LIST_MAIL
type MsgMhfListMail struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfListMail) Opcode() network.PacketID {
	return network.MSG_MHF_LIST_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfListMail) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfListMail) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
