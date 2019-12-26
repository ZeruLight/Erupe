package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateMercenaryLog represents the MSG_MHF_ENUMERATE_MERCENARY_LOG
type MsgMhfEnumerateMercenaryLog struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateMercenaryLog) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_MERCENARY_LOG
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateMercenaryLog) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateMercenaryLog) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}