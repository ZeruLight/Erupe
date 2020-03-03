package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSendMail represents the MSG_MHF_SEND_MAIL
type MsgMhfSendMail struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSendMail) Opcode() network.PacketID {
	return network.MSG_MHF_SEND_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfSendMail) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSendMail) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
