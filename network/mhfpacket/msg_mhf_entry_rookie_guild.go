package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfEntryRookieGuild represents the MSG_MHF_ENTRY_ROOKIE_GUILD
type MsgMhfEntryRookieGuild struct {
	AckHandle uint32
	Unk       uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEntryRookieGuild) Opcode() network.PacketID {
	return network.MSG_MHF_ENTRY_ROOKIE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfEntryRookieGuild) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEntryRookieGuild) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
