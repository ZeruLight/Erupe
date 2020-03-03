package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfStampcardPrize represents the MSG_MHF_STAMPCARD_PRIZE
type MsgMhfStampcardPrize struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStampcardPrize) Opcode() network.PacketID {
	return network.MSG_MHF_STAMPCARD_PRIZE
}

// Parse parses the packet from binary
func (m *MsgMhfStampcardPrize) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStampcardPrize) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
