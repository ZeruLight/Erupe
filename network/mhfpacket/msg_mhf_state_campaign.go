package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfStateCampaign represents the MSG_MHF_STATE_CAMPAIGN
type MsgMhfStateCampaign struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStateCampaign) Opcode() network.PacketID {
	return network.MSG_MHF_STATE_CAMPAIGN
}

// Parse parses the packet from binary
func (m *MsgMhfStateCampaign) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStateCampaign) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
