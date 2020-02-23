package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdMyRanking represents the MSG_MHF_GET_UD_MY_RANKING
type MsgMhfGetUdMyRanking struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdMyRanking) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_MY_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdMyRanking) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdMyRanking) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
