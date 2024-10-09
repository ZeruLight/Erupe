package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfReleaseEvent represents the MSG_MHF_RELEASE_EVENT
type MsgMhfReleaseEvent struct {
	AckHandle uint32
	RaviID    uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReleaseEvent) Opcode() network.PacketID {
	return network.MSG_MHF_RELEASE_EVENT
}

// Parse parses the packet from binary
func (m *MsgMhfReleaseEvent) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.RaviID = bf.ReadUint32()
	bf.ReadUint32() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReleaseEvent) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
