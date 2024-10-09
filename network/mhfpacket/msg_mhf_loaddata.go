package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfLoaddata represents the MSG_MHF_LOADDATA
type MsgMhfLoaddata struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoaddata) Opcode() network.PacketID {
	return network.MSG_MHF_LOADDATA
}

// Parse parses the packet from binary
func (m *MsgMhfLoaddata) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoaddata) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
