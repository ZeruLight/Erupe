package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadLastWeekBeatRanking represents the MSG_MHF_READ_LAST_WEEK_BEAT_RANKING
type MsgMhfReadLastWeekBeatRanking struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadLastWeekBeatRanking) Opcode() network.PacketID {
	return network.MSG_MHF_READ_LAST_WEEK_BEAT_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfReadLastWeekBeatRanking) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadLastWeekBeatRanking) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
