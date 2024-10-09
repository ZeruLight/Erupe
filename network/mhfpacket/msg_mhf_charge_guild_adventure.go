package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfChargeGuildAdventure represents the MSG_MHF_CHARGE_GUILD_ADVENTURE
type MsgMhfChargeGuildAdventure struct {
	AckHandle uint32
	ID        uint32
	Amount    uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfChargeGuildAdventure) Opcode() network.PacketID {
	return network.MSG_MHF_CHARGE_GUILD_ADVENTURE
}

// Parse parses the packet from binary
func (m *MsgMhfChargeGuildAdventure) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.ID = bf.ReadUint32()
	m.Amount = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfChargeGuildAdventure) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
