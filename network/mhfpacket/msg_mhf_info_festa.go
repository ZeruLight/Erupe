package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfInfoFesta represents the MSG_MHF_INFO_FESTA
type MsgMhfInfoFesta struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfInfoFesta) Opcode() network.PacketID {
	return network.MSG_MHF_INFO_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfInfoFesta) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfInfoFesta) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}