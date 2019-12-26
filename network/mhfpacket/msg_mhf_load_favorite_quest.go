package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadFavoriteQuest represents the MSG_MHF_LOAD_FAVORITE_QUEST
type MsgMhfLoadFavoriteQuest struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadFavoriteQuest) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_FAVORITE_QUEST
}

// Parse parses the packet from binary
func (m *MsgMhfLoadFavoriteQuest) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadFavoriteQuest) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}