package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfGetCafeDurationBonusInfo represents the MSG_MHF_GET_CAFE_DURATION_BONUS_INFO
type MsgMhfGetCafeDurationBonusInfo struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetCafeDurationBonusInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_CAFE_DURATION_BONUS_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetCafeDurationBonusInfo) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetCafeDurationBonusInfo) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
