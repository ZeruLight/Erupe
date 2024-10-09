package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfGetSenyuDailyCount represents the MSG_MHF_GET_SENYU_DAILY_COUNT
type MsgMhfGetSenyuDailyCount struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetSenyuDailyCount) Opcode() network.PacketID {
	return network.MSG_MHF_GET_SENYU_DAILY_COUNT
}

// Parse parses the packet from binary
func (m *MsgMhfGetSenyuDailyCount) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetSenyuDailyCount) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
