package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfCaravanMyRank represents the MSG_MHF_CARAVAN_MY_RANK
type MsgMhfCaravanMyRank struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCaravanMyRank) Opcode() network.PacketID {
	return network.MSG_MHF_CARAVAN_MY_RANK
}

// Parse parses the packet from binary
func (m *MsgMhfCaravanMyRank) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCaravanMyRank) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
