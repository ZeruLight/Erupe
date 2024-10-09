package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfPostCafeDurationBonusReceived represents the MSG_MHF_POST_CAFE_DURATION_BONUS_RECEIVED
type MsgMhfPostCafeDurationBonusReceived struct {
	AckHandle   uint32
	CafeBonusID []uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostCafeDurationBonusReceived) Opcode() network.PacketID {
	return network.MSG_MHF_POST_CAFE_DURATION_BONUS_RECEIVED
}

// Parse parses the packet from binary
func (m *MsgMhfPostCafeDurationBonusReceived) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	ids := int(bf.ReadUint32())
	for i := 0; i < ids; i++ {
		m.CafeBonusID = append(m.CafeBonusID, bf.ReadUint32())
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostCafeDurationBonusReceived) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
