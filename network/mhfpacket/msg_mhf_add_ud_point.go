package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAddUdPoint represents the MSG_MHF_ADD_UD_POINT
type MsgMhfAddUdPoint struct {
	AckHandle uint32
	Unk1      uint32
	Unk2      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAddUdPoint) Opcode() network.PacketID {
	return network.MSG_MHF_ADD_UD_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfAddUdPoint) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()

	return nil
	//panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAddUdPoint) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
