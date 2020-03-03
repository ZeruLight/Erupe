package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfExchangeWeeklyStamp represents the MSG_MHF_EXCHANGE_WEEKLY_STAMP
type MsgMhfExchangeWeeklyStamp struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfExchangeWeeklyStamp) Opcode() network.PacketID {
	return network.MSG_MHF_EXCHANGE_WEEKLY_STAMP
}

// Parse parses the packet from binary
func (m *MsgMhfExchangeWeeklyStamp) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfExchangeWeeklyStamp) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
