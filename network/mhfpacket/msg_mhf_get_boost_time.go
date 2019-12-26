package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetBoostTime represents the MSG_MHF_GET_BOOST_TIME
type MsgMhfGetBoostTime struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetBoostTime) Opcode() network.PacketID {
	return network.MSG_MHF_GET_BOOST_TIME
}

// Parse parses the packet from binary
func (m *MsgMhfGetBoostTime) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetBoostTime) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}