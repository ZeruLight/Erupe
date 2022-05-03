package mhfpacket

import (
 "errors"

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve205 represents the MSG_SYS_reserve205
type MsgSysReserve205 struct {
  AckHandle uint32
  Unk0 uint32
  Unk1 uint32
  Unk2 uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve205) Opcode() network.PacketID {
	return network.MSG_SYS_reserve205
}

// Parse parses the packet from binary
func (m *MsgSysReserve205) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
  m.AckHandle = bf.ReadUint32()
  m.Unk0 = bf.ReadUint32()
  m.Unk1 = bf.ReadUint32()
  m.Unk2 = bf.ReadUint32()
  return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve205) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
