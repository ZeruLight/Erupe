package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfVoteFesta represents the MSG_MHF_VOTE_FESTA
type MsgMhfVoteFesta struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfVoteFesta) Opcode() network.PacketID {
	return network.MSG_MHF_VOTE_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfVoteFesta) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfVoteFesta) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
