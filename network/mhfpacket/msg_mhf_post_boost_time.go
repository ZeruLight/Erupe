package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPostBoostTime represents the MSG_MHF_POST_BOOST_TIME
type MsgMhfPostBoostTime struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostBoostTime) Opcode() network.PacketID {
	return network.MSG_MHF_POST_BOOST_TIME
}

// Parse parses the packet from binary
func (m *MsgMhfPostBoostTime) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostBoostTime) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}