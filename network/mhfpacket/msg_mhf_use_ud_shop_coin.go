package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUseUdShopCoin represents the MSG_MHF_USE_UD_SHOP_COIN
type MsgMhfUseUdShopCoin struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUseUdShopCoin) Opcode() network.PacketID {
	return network.MSG_MHF_USE_UD_SHOP_COIN
}

// Parse parses the packet from binary
func (m *MsgMhfUseUdShopCoin) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUseUdShopCoin) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
