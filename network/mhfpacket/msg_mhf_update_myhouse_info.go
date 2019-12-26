package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateMyhouseInfo represents the MSG_MHF_UPDATE_MYHOUSE_INFO
type MsgMhfUpdateMyhouseInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateMyhouseInfo) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_MYHOUSE_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateMyhouseInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateMyhouseInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}