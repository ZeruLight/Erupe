package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEnumerateInvGuild represents the MSG_MHF_ENUMERATE_INV_GUILD
type MsgMhfEnumerateInvGuild struct {
	AckHandle    uint32
	Unk          uint32
	Operation    uint8
	ActiveHours  uint8
	DaysActive   uint8
	PlayStyle    uint8
	GuildRequest uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateInvGuild) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_INV_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateInvGuild) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk = bf.ReadUint32()
	m.Operation = bf.ReadUint8()
	m.ActiveHours = bf.ReadUint8()
	m.DaysActive = bf.ReadUint8()
	m.PlayStyle = bf.ReadUint8()
	m.GuildRequest = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateInvGuild) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
