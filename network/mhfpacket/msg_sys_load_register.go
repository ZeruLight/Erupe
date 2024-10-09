package mhfpacket

import (
	"errors"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgSysLoadRegister represents the MSG_SYS_LOAD_REGISTER
type MsgSysLoadRegister struct {
	AckHandle  uint32
	RegisterID uint32
	Values     uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLoadRegister) Opcode() network.PacketID {
	return network.MSG_SYS_LOAD_REGISTER
}

// Parse parses the packet from binary
func (m *MsgSysLoadRegister) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.RegisterID = bf.ReadUint32()
	m.Values = bf.ReadUint8()
	bf.ReadUint8()  // Zeroed
	bf.ReadUint16() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysLoadRegister) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
