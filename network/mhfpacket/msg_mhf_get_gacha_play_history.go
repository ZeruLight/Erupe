package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfGetGachaPlayHistory represents the MSG_MHF_GET_GACHA_PLAY_HISTORY
type MsgMhfGetGachaPlayHistory struct {
	AckHandle uint32
	GachaID   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGachaPlayHistory) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GACHA_PLAY_HISTORY
}

// Parse parses the packet from binary
func (m *MsgMhfGetGachaPlayHistory) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.GachaID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGachaPlayHistory) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
