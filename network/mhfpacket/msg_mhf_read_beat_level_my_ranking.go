package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadBeatLevelMyRanking represents the MSG_MHF_READ_BEAT_LEVEL_MY_RANKING
type MsgMhfReadBeatLevelMyRanking struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadBeatLevelMyRanking) Opcode() network.PacketID {
	return network.MSG_MHF_READ_BEAT_LEVEL_MY_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfReadBeatLevelMyRanking) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadBeatLevelMyRanking) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
