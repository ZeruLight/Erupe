package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAcquireGuildTresure represents the MSG_MHF_ACQUIRE_GUILD_TRESURE
type MsgMhfAcquireGuildTresure struct {
	AckHandle uint32
	HuntID    uint32
	Unk       bool
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireGuildTresure) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_GUILD_TRESURE
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireGuildTresure) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.HuntID = bf.ReadUint32()
	m.Unk = bf.ReadBool()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireGuildTresure) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
