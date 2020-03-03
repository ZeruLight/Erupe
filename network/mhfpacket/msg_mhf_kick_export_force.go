package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfKickExportForce represents the MSG_MHF_KICK_EXPORT_FORCE
type MsgMhfKickExportForce struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfKickExportForce) Opcode() network.PacketID {
	return network.MSG_MHF_KICK_EXPORT_FORCE
}

// Parse parses the packet from binary
func (m *MsgMhfKickExportForce) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfKickExportForce) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
