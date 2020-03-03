package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfExchangeItem2Fpoint represents the MSG_MHF_EXCHANGE_ITEM_2_FPOINT
type MsgMhfExchangeItem2Fpoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfExchangeItem2Fpoint) Opcode() network.PacketID {
	return network.MSG_MHF_EXCHANGE_ITEM_2_FPOINT
}

// Parse parses the packet from binary
func (m *MsgMhfExchangeItem2Fpoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfExchangeItem2Fpoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
