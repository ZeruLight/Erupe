package mhfpacket

import (
	"errors"
	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfCheckDailyCafepoint represents the MSG_MHF_CHECK_DAILY_CAFEPOINT
type MsgMhfCheckDailyCafepoint struct {
	AckHandle uint32
	Unk       uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCheckDailyCafepoint) Opcode() network.PacketID {
	return network.MSG_MHF_CHECK_DAILY_CAFEPOINT
}

// Parse parses the packet from binary
func (m *MsgMhfCheckDailyCafepoint) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk = bf.ReadUint32()
	return nil
}

func (m *MsgMhfCheckDailyCafepoint) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
