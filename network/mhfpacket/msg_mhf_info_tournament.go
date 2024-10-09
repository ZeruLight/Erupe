package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfInfoTournament represents the MSG_MHF_INFO_TOURNAMENT
type MsgMhfInfoTournament struct {
	AckHandle uint32
	Unk0      uint8
	Unk1      uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfInfoTournament) Opcode() network.PacketID {
	return network.MSG_MHF_INFO_TOURNAMENT
}

// Parse parses the packet from binary
func (m *MsgMhfInfoTournament) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	m.Unk1 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfInfoTournament) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
