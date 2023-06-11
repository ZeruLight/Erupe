package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfPostTowerInfo represents the MSG_MHF_POST_TOWER_INFO
type MsgMhfPostTowerInfo struct {
	AckHandle uint32
	InfoType  uint32
	Unk1      uint32
	Unk2      int32
	Unk3      int32
	Unk4      int32
	Unk5      int32
	Unk6      int32
	Unk7      int32
	Unk8      int32
	Unk9      int64
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostTowerInfo) Opcode() network.PacketID {
	return network.MSG_MHF_POST_TOWER_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfPostTowerInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.InfoType = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadInt32()
	m.Unk3 = bf.ReadInt32()
	m.Unk4 = bf.ReadInt32()
	m.Unk5 = bf.ReadInt32()
	m.Unk6 = bf.ReadInt32()
	m.Unk7 = bf.ReadInt32()
	m.Unk8 = bf.ReadInt32()
	m.Unk9 = bf.ReadInt64()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostTowerInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
