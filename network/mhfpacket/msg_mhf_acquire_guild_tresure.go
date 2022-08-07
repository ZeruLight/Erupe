package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfAcquireGuildTresure represents the MSG_MHF_ACQUIRE_GUILD_TRESURE
type MsgMhfAcquireGuildTresure struct {
	AckHandle uint32
	HuntID    uint32
	Unk       uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireGuildTresure) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_GUILD_TRESURE
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireGuildTresure) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.HuntID = bf.ReadUint32()
	m.Unk = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireGuildTresure) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
