package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfEntryTournament represents the MSG_MHF_ENTRY_TOURNAMENT
type MsgMhfEntryTournament struct {
	AckHandle    uint32
	TournamentID uint32
	Unk0         uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEntryTournament) Opcode() network.PacketID {
	return network.MSG_MHF_ENTRY_TOURNAMENT
}

// Parse parses the packet from binary
func (m *MsgMhfEntryTournament) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.TournamentID = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEntryTournament) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
