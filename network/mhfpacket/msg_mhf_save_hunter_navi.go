package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSaveHunterNavi represents the MSG_MHF_SAVE_HUNTER_NAVI
type MsgMhfSaveHunterNavi struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSaveHunterNavi) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_HUNTER_NAVI
}

// Parse parses the packet from binary
func (m *MsgMhfSaveHunterNavi) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSaveHunterNavi) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}