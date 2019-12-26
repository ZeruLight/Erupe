package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireExchangeShop represents the MSG_MHF_ACQUIRE_EXCHANGE_SHOP
type MsgMhfAcquireExchangeShop struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireExchangeShop) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_EXCHANGE_SHOP
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireExchangeShop) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireExchangeShop) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}