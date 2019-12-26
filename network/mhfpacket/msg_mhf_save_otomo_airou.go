package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSaveOtomoAirou represents the MSG_MHF_SAVE_OTOMO_AIROU
type MsgMhfSaveOtomoAirou struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSaveOtomoAirou) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_OTOMO_AIROU
}

// Parse parses the packet from binary
func (m *MsgMhfSaveOtomoAirou) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSaveOtomoAirou) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}