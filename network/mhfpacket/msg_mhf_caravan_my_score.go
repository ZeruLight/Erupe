package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCaravanMyScore represents the MSG_MHF_CARAVAN_MY_SCORE
type MsgMhfCaravanMyScore struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCaravanMyScore) Opcode() network.PacketID {
	return network.MSG_MHF_CARAVAN_MY_SCORE
}

// Parse parses the packet from binary
func (m *MsgMhfCaravanMyScore) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCaravanMyScore) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
