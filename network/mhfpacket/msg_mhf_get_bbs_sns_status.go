package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfGetBbsSnsStatus represents the MSG_MHF_GET_BBS_SNS_STATUS
type MsgMhfGetBbsSnsStatus struct {
	AckHandle uint32
	Unk       []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetBbsSnsStatus) Opcode() network.PacketID {
	return network.MSG_MHF_GET_BBS_SNS_STATUS
}

// Parse parses the packet from binary
func (m *MsgMhfGetBbsSnsStatus) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk = bf.ReadBytes(12)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetBbsSnsStatus) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
