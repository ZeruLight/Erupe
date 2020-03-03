package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadBeatLevelAllRanking represents the MSG_MHF_READ_BEAT_LEVEL_ALL_RANKING
type MsgMhfReadBeatLevelAllRanking struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadBeatLevelAllRanking) Opcode() network.PacketID {
	return network.MSG_MHF_READ_BEAT_LEVEL_ALL_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfReadBeatLevelAllRanking) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadBeatLevelAllRanking) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
