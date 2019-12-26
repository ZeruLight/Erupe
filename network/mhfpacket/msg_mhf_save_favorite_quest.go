package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSaveFavoriteQuest represents the MSG_MHF_SAVE_FAVORITE_QUEST
type MsgMhfSaveFavoriteQuest struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSaveFavoriteQuest) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_FAVORITE_QUEST
}

// Parse parses the packet from binary
func (m *MsgMhfSaveFavoriteQuest) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSaveFavoriteQuest) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}