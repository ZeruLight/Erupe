package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateRanking represents the MSG_MHF_ENUMERATE_RANKING
type MsgMhfEnumerateRanking struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateRanking) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateRanking) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateRanking) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}