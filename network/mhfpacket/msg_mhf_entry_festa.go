package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEntryFesta represents the MSG_MHF_ENTRY_FESTA
type MsgMhfEntryFesta struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEntryFesta) Opcode() network.PacketID {
	return network.MSG_MHF_ENTRY_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfEntryFesta) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEntryFesta) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
