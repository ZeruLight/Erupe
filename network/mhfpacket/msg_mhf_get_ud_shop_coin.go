package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdShopCoin represents the MSG_MHF_GET_UD_SHOP_COIN
type MsgMhfGetUdShopCoin struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdShopCoin) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_SHOP_COIN
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdShopCoin) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdShopCoin) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
