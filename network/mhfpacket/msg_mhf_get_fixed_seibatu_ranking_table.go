package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetFixedSeibatuRankingTable represents the MSG_MHF_GET_FIXED_SEIBATU_RANKING_TABLE
type MsgMhfGetFixedSeibatuRankingTable struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetFixedSeibatuRankingTable) Opcode() network.PacketID {
	return network.MSG_MHF_GET_FIXED_SEIBATU_RANKING_TABLE
}

// Parse parses the packet from binary
func (m *MsgMhfGetFixedSeibatuRankingTable) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetFixedSeibatuRankingTable) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
