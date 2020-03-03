package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetCogInfo represents the MSG_MHF_GET_COG_INFO
type MsgMhfGetCogInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetCogInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_COG_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetCogInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetCogInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
