package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfReadMercenaryW represents the MSG_MHF_READ_MERCENARY_W
type MsgMhfReadMercenaryW struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadMercenaryW) Opcode() network.PacketID {
	return network.MSG_MHF_READ_MERCENARY_W
}

// Parse parses the packet from binary
func (m *MsgMhfReadMercenaryW) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadMercenaryW) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}