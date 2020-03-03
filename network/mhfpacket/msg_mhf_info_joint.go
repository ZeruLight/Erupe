package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfInfoJoint represents the MSG_MHF_INFO_JOINT
type MsgMhfInfoJoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfInfoJoint) Opcode() network.PacketID {
	return network.MSG_MHF_INFO_JOINT
}

// Parse parses the packet from binary
func (m *MsgMhfInfoJoint) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfInfoJoint) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
