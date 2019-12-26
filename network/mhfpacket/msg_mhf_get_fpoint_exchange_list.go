package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetFpointExchangeList represents the MSG_MHF_GET_FPOINT_EXCHANGE_LIST
type MsgMhfGetFpointExchangeList struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetFpointExchangeList) Opcode() network.PacketID {
	return network.MSG_MHF_GET_FPOINT_EXCHANGE_LIST
}

// Parse parses the packet from binary
func (m *MsgMhfGetFpointExchangeList) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetFpointExchangeList) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}