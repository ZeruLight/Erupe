package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfLoadPlateBox represents the MSG_MHF_LOAD_PLATE_BOX
type MsgMhfLoadPlateBox struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadPlateBox) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_PLATE_BOX
}

// Parse parses the packet from binary
func (m *MsgMhfLoadPlateBox) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadPlateBox) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
