package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEnumerateTitle represents the MSG_MHF_ENUMERATE_TITLE
type MsgMhfEnumerateTitle struct {
	AckHandle uint32
	CharID    uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateTitle) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_TITLE
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateTitle) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.CharID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateTitle) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
