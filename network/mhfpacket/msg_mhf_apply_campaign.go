package mhfpacket

import (
	"errors"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfApplyCampaign represents the MSG_MHF_APPLY_CAMPAIGN
type MsgMhfApplyCampaign struct {
	AckHandle   uint32
	CampaignID  uint32
	NullPadding uint16 // set as 0 in z2
	CodeString  string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfApplyCampaign) Opcode() network.PacketID {
	return network.MSG_MHF_APPLY_CAMPAIGN
}

// Parse parses the packet from binary
func (m *MsgMhfApplyCampaign) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.CampaignID = bf.ReadUint32()
	m.NullPadding = bf.ReadUint16()
	m.CodeString = stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	bf.ReadInt8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfApplyCampaign) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
