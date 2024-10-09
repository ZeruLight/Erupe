package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfGetBbsUserStatus represents the MSG_MHF_GET_BBS_USER_STATUS
type MsgMhfGetBbsUserStatus struct {
	AckHandle uint32
	Unk       []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetBbsUserStatus) Opcode() network.PacketID {
	return network.MSG_MHF_GET_BBS_USER_STATUS
}

// Parse parses the packet from binary
func (m *MsgMhfGetBbsUserStatus) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk = bf.ReadBytes(12)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetBbsUserStatus) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
