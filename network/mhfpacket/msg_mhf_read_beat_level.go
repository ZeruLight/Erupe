package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadBeatLevel represents the MSG_MHF_READ_BEAT_LEVEL
type MsgMhfReadBeatLevel struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadBeatLevel) Opcode() network.PacketID {
	return network.MSG_MHF_READ_BEAT_LEVEL
}

// Parse parses the packet from binary
func (m *MsgMhfReadBeatLevel) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadBeatLevel) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}