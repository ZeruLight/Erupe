package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetMyhouseInfo represents the MSG_MHF_GET_MYHOUSE_INFO
type MsgMhfGetMyhouseInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetMyhouseInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_MYHOUSE_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetMyhouseInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetMyhouseInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}