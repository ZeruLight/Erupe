package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfInfoTournament represents the MSG_MHF_INFO_TOURNAMENT
type MsgMhfInfoTournament struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfInfoTournament) Opcode() network.PacketID {
	return network.MSG_MHF_INFO_TOURNAMENT
}

// Parse parses the packet from binary
func (m *MsgMhfInfoTournament) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfInfoTournament) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
