package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAcquireGuildAdventure represents the MSG_MHF_ACQUIRE_GUILD_ADVENTURE
type MsgMhfAcquireGuildAdventure struct {
	AckHandle uint32
	ID        uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireGuildAdventure) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_GUILD_ADVENTURE
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireGuildAdventure) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.ID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireGuildAdventure) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
