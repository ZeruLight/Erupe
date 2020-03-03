package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnterTournamentQuest represents the MSG_MHF_ENTER_TOURNAMENT_QUEST
type MsgMhfEnterTournamentQuest struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnterTournamentQuest) Opcode() network.PacketID {
	return network.MSG_MHF_ENTER_TOURNAMENT_QUEST
}

// Parse parses the packet from binary
func (m *MsgMhfEnterTournamentQuest) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnterTournamentQuest) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
