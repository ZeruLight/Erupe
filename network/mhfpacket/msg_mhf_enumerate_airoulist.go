package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateAiroulist represents the MSG_MHF_ENUMERATE_AIROULIST
type MsgMhfEnumerateAiroulist struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateAiroulist) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_AIROULIST
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateAiroulist) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateAiroulist) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
