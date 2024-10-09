package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfStartBoostTime represents the MSG_MHF_START_BOOST_TIME
type MsgMhfStartBoostTime struct {
	AckHandle uint32
	Unk0      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStartBoostTime) Opcode() network.PacketID {
	return network.MSG_MHF_START_BOOST_TIME
}

// Parse parses the packet from binary
func (m *MsgMhfStartBoostTime) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStartBoostTime) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
