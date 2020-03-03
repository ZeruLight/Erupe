package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateRengokuRanking represents the MSG_MHF_ENUMERATE_RENGOKU_RANKING
type MsgMhfEnumerateRengokuRanking struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateRengokuRanking) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_RENGOKU_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateRengokuRanking) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateRengokuRanking) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
