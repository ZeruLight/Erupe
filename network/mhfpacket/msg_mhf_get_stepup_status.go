package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetStepupStatus represents the MSG_MHF_GET_STEPUP_STATUS
type MsgMhfGetStepupStatus struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetStepupStatus) Opcode() network.PacketID {
	return network.MSG_MHF_GET_STEPUP_STATUS
}

// Parse parses the packet from binary
func (m *MsgMhfGetStepupStatus) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetStepupStatus) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
