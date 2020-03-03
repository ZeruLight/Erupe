package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfTransitMessage represents the MSG_MHF_TRANSIT_MESSAGE
type MsgMhfTransitMessage struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfTransitMessage) Opcode() network.PacketID {
	return network.MSG_MHF_TRANSIT_MESSAGE
}

// Parse parses the packet from binary
func (m *MsgMhfTransitMessage) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfTransitMessage) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
