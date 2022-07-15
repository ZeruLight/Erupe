package mhfpacket

import (
 "errors"

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgMhfVoteFesta represents the MSG_MHF_VOTE_FESTA
type MsgMhfVoteFesta struct {
  AckHandle uint32
  Unk uint32
  GuildID uint32
  TrialID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfVoteFesta) Opcode() network.PacketID {
	return network.MSG_MHF_VOTE_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfVoteFesta) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
  m.AckHandle = bf.ReadUint32()
  m.Unk = bf.ReadUint32()
  m.GuildID = bf.ReadUint32()
  m.TrialID = bf.ReadUint32()
  return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfVoteFesta) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
