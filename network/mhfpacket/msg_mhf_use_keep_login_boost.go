package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUseKeepLoginBoost represents the MSG_MHF_USE_KEEP_LOGIN_BOOST
type MsgMhfUseKeepLoginBoost struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUseKeepLoginBoost) Opcode() network.PacketID {
	return network.MSG_MHF_USE_KEEP_LOGIN_BOOST
}

// Parse parses the packet from binary
func (m *MsgMhfUseKeepLoginBoost) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUseKeepLoginBoost) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
