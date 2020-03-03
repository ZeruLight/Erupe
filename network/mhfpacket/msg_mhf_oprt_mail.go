package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfOprtMail represents the MSG_MHF_OPRT_MAIL
type MsgMhfOprtMail struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOprtMail) Opcode() network.PacketID {
	return network.MSG_MHF_OPRT_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfOprtMail) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOprtMail) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
