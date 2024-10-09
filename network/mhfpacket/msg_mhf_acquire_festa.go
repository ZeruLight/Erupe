package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAcquireFesta represents the MSG_MHF_ACQUIRE_FESTA
type MsgMhfAcquireFesta struct {
	AckHandle uint32
	FestaID   uint32
	GuildID   uint32
	Unk       uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireFesta) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireFesta) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.FestaID = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	m.Unk = bf.ReadUint8()
	bf.ReadUint8() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireFesta) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
