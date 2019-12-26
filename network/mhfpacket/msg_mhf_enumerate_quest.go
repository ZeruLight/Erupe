package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateQuest represents the MSG_MHF_ENUMERATE_QUEST
type MsgMhfEnumerateQuest struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateQuest) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_QUEST
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateQuest) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateQuest) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}