package mhfpacket

import (
 "errors"

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgMhfTransitMessage represents the MSG_MHF_TRANSIT_MESSAGE
type MsgMhfTransitMessage struct {
  AckHandle uint32
  Unk0 uint8
  Unk1 uint8
  Unk2 uint16
  Unk3 uint16
  TargetID uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfTransitMessage) Opcode() network.PacketID {
	return network.MSG_MHF_TRANSIT_MESSAGE
}

// Parse parses the packet from binary
func (m *MsgMhfTransitMessage) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
  m.AckHandle = bf.ReadUint32()
  m.Unk0 = bf.ReadUint8()
  m.Unk1 = bf.ReadUint8()
  m.Unk2 = bf.ReadUint16()
  m.Unk3 = bf.ReadUint16()
  m.TargetID = bf.ReadUint32()
  return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfTransitMessage) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
