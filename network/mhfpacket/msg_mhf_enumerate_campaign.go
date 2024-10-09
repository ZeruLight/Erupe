package mhfpacket

import (
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEnumerateCampaign represents the MSG_MHF_ENUMERATE_CAMPAIGN
type MsgMhfEnumerateCampaign struct {
	AckHandle uint32
	Unk0      uint16
	Unk1      uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateCampaign) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_CAMPAIGN
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateCampaign) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateCampaign) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	bf.WriteUint32(m.AckHandle)
	bf.WriteUint16(m.Unk0)
	bf.WriteUint16(m.Unk1)
	return nil
}
