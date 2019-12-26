package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetKeepLoginBoostStatus represents the MSG_MHF_GET_KEEP_LOGIN_BOOST_STATUS
type MsgMhfGetKeepLoginBoostStatus struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetKeepLoginBoostStatus) Opcode() network.PacketID {
	return network.MSG_MHF_GET_KEEP_LOGIN_BOOST_STATUS
}

// Parse parses the packet from binary
func (m *MsgMhfGetKeepLoginBoostStatus) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetKeepLoginBoostStatus) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}