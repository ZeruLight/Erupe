package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireItem represents the MSG_MHF_ACQUIRE_ITEM
type MsgMhfAcquireItem struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireItem) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireItem) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireItem) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
