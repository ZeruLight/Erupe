package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSaveDecoMyset represents the MSG_MHF_SAVE_DECO_MYSET
type MsgMhfSaveDecoMyset struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSaveDecoMyset) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_DECO_MYSET
}

// Parse parses the packet from binary
func (m *MsgMhfSaveDecoMyset) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSaveDecoMyset) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}