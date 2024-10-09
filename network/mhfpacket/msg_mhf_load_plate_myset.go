package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfLoadPlateMyset represents the MSG_MHF_LOAD_PLATE_MYSET
type MsgMhfLoadPlateMyset struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadPlateMyset) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_PLATE_MYSET
}

// Parse parses the packet from binary
func (m *MsgMhfLoadPlateMyset) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadPlateMyset) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
