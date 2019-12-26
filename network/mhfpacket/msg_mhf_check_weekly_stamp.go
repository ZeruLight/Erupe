package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCheckWeeklyStamp represents the MSG_MHF_CHECK_WEEKLY_STAMP
type MsgMhfCheckWeeklyStamp struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCheckWeeklyStamp) Opcode() network.PacketID {
	return network.MSG_MHF_CHECK_WEEKLY_STAMP
}

// Parse parses the packet from binary
func (m *MsgMhfCheckWeeklyStamp) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCheckWeeklyStamp) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}