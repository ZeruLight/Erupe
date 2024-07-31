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
	Skill     int32
	TR        int32
	TRP       int32
	Cost      int32
	Unk6      int32
	Unk7      int32
	Block1    int32
	TimeTaken int32
	CID       int32
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
	m.Skill = bf.ReadInt32()
	m.TR = bf.ReadInt32()
	m.TRP = bf.ReadInt32()
	m.Cost = bf.ReadInt32()
	m.Unk6 = bf.ReadInt32()
	m.Unk7 = bf.ReadInt32()
	m.Block1 = bf.ReadInt32()
	m.TimeTaken = bf.ReadInt32()
	m.CID = bf.ReadInt32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostTowerInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
