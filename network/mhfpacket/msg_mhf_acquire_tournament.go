package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAcquireTournament represents the MSG_MHF_ACQUIRE_TOURNAMENT
type MsgMhfAcquireTournament struct {
	AckHandle    uint32
	TournamentID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireTournament) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_TOURNAMENT
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireTournament) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.TournamentID = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireTournament) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
