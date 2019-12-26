package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdRanking represents the MSG_MHF_GET_UD_RANKING
type MsgMhfGetUdRanking struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdRanking) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdRanking) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdRanking) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}