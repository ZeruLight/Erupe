package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetEarthValue represents the MSG_MHF_GET_EARTH_VALUE
type MsgMhfGetEarthValue struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetEarthValue) Opcode() network.PacketID {
	return network.MSG_MHF_GET_EARTH_VALUE
}

// Parse parses the packet from binary
func (m *MsgMhfGetEarthValue) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetEarthValue) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}