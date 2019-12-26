package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadOtomoAirou represents the MSG_MHF_LOAD_OTOMO_AIROU
type MsgMhfLoadOtomoAirou struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadOtomoAirou) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_OTOMO_AIROU
}

// Parse parses the packet from binary
func (m *MsgMhfLoadOtomoAirou) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadOtomoAirou) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}