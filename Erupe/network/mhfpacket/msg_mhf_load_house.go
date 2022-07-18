package mhfpacket

import (
 "errors"

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgMhfLoadHouse represents the MSG_MHF_LOAD_HOUSE
type MsgMhfLoadHouse struct {
	AckHandle uint32
	CharID uint32
  // dest?
  // 0x3 = house
  // 0x4 = bookshelf
  // 0x5 = gallery
  // 0x8 = tore
  // 0x9 = own house
  // 0xA = garden
	Unk1 uint8
  // bool inMezSquare?
	Unk2 uint8
	Unk3 uint16 // Hardcoded 0 in binary
	Password []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadHouse) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_HOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfLoadHouse) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.CharID = bf.ReadUint32()
	m.Unk1 = bf.ReadUint8()
	m.Unk2 = bf.ReadUint8()
	_ = bf.ReadUint16()
  _ = bf.ReadUint8() // Password length
	m.Password = bf.ReadNullTerminatedBytes()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadHouse) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
