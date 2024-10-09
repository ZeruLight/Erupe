package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfGetUdMonsterPoint represents the MSG_MHF_GET_UD_MONSTER_POINT
type MsgMhfGetUdMonsterPoint struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdMonsterPoint) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_MONSTER_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdMonsterPoint) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdMonsterPoint) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
