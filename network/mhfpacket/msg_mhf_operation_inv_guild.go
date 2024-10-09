package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfOperationInvGuild represents the MSG_MHF_OPERATION_INV_GUILD
type MsgMhfOperationInvGuild struct {
	AckHandle    uint32
	Operation    uint8
	ActiveHours  uint8
	DaysActive   uint8
	PlayStyle    uint8
	GuildRequest uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOperationInvGuild) Opcode() network.PacketID {
	return network.MSG_MHF_OPERATION_INV_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfOperationInvGuild) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Operation = bf.ReadUint8()
	m.ActiveHours = bf.ReadUint8()
	m.DaysActive = bf.ReadUint8()
	m.PlayStyle = bf.ReadUint8()
	m.GuildRequest = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOperationInvGuild) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
