package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfLoadGuildAdventure represents the MSG_MHF_LOAD_GUILD_ADVENTURE
type MsgMhfLoadGuildAdventure struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadGuildAdventure) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_GUILD_ADVENTURE
}

// Parse parses the packet from binary
func (m *MsgMhfLoadGuildAdventure) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadGuildAdventure) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
