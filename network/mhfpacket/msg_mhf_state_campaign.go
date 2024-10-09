package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfStateCampaign represents the MSG_MHF_STATE_CAMPAIGN
type MsgMhfStateCampaign struct {
	AckHandle  uint32
	CampaignID uint32
	Unk1       uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStateCampaign) Opcode() network.PacketID {
	return network.MSG_MHF_STATE_CAMPAIGN
}

// Parse parses the packet from binary
func (m *MsgMhfStateCampaign) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.CampaignID = bf.ReadUint32()
	m.Unk1 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStateCampaign) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
