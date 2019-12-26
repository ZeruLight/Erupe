package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfListMember represents the MSG_MHF_LIST_MEMBER
type MsgMhfListMember struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfListMember) Opcode() network.PacketID {
	return network.MSG_MHF_LIST_MEMBER
}

// Parse parses the packet from binary
func (m *MsgMhfListMember) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfListMember) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}