package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfGuildHuntdata represents the MSG_MHF_GUILD_HUNTDATA
type MsgMhfGuildHuntdata struct {
	AckHandle uint32
	Operation uint8
	GuildID   uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGuildHuntdata) Opcode() network.PacketID {
	return network.MSG_MHF_GUILD_HUNTDATA
}

// Parse parses the packet from binary
func (m *MsgMhfGuildHuntdata) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Operation = bf.ReadUint8()
	if m.Operation == 1 {
		m.GuildID = bf.ReadUint32()
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGuildHuntdata) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
