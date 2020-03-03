package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfApplyCampaign represents the MSG_MHF_APPLY_CAMPAIGN
type MsgMhfApplyCampaign struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfApplyCampaign) Opcode() network.PacketID {
	return network.MSG_MHF_APPLY_CAMPAIGN
}

// Parse parses the packet from binary
func (m *MsgMhfApplyCampaign) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfApplyCampaign) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
