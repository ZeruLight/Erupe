package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAddRewardSongCount represents the MSG_MHF_ADD_REWARD_SONG_COUNT
type MsgMhfAddRewardSongCount struct {
	AckHandle uint32
	PrayerID  uint32
	Unk1      uint16
	Unk2      uint8
	Unk3      []uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddRewardSongCount) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_REWARD_SONG_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfAddRewardSongCount) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.PrayerID = bf.ReadUint32()
	m.Unk1 = bf.ReadUint16()
	m.Unk2 = bf.ReadUint8()
	for i := uint8(0); i < m.Unk2; i++ {
		m.Unk3 = append(m.Unk3, bf.ReadUint16())
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddRewardSongCount) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
