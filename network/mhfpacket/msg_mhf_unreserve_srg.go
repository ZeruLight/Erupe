package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUnreserveSrg represents the MSG_MHF_UNRESERVE_SRG
type MsgMhfUnreserveSrg struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUnreserveSrg) Opcode() network.PacketID {
	return network.MSG_MHF_UNRESERVE_SRG
}

// Parse parses the packet from binary
func (m *MsgMhfUnreserveSrg) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUnreserveSrg) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
