package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCreateJoint represents the MSG_MHF_CREATE_JOINT
type MsgMhfCreateJoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCreateJoint) Opcode() network.PacketID {
	return network.MSG_MHF_CREATE_JOINT
}

// Parse parses the packet from binary
func (m *MsgMhfCreateJoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCreateJoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
