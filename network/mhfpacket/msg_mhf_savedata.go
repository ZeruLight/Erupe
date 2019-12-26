package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSavedata represents the MSG_MHF_SAVEDATA
type MsgMhfSavedata struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSavedata) Opcode() network.PacketID {
	return network.MSG_MHF_SAVEDATA
}

// Parse parses the packet from binary
func (m *MsgMhfSavedata) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSavedata) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}