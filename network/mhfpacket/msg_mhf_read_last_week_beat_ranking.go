package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfReadLastWeekBeatRanking represents the MSG_MHF_READ_LAST_WEEK_BEAT_RANKING
type MsgMhfReadLastWeekBeatRanking struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadLastWeekBeatRanking) Opcode() network.PacketID {
	return network.MSG_MHF_READ_LAST_WEEK_BEAT_RANKING
}

// Parse parses the packet from binary
func (m *MsgMhfReadLastWeekBeatRanking) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadInt32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadLastWeekBeatRanking) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
