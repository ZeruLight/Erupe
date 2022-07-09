package mhfpacket

import (
 "errors"

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysReserve205 represents the MSG_SYS_reserve205
type MsgSysReserve205 struct {
  AckHandle uint32
  Destination uint32
  Charge uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve205) Opcode() network.PacketID {
	return network.MSG_SYS_reserve205
}

// Parse parses the packet from binary
func (m *MsgSysReserve205) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
  m.AckHandle = bf.ReadUint32()
  m.Destination = bf.ReadUint32()
  m.Charge = bf.ReadUint32()
  _ = bf.ReadUint32() // CharID
  return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve205) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
