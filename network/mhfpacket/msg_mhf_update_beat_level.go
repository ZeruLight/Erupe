package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfUpdateBeatLevel represents the MSG_MHF_UPDATE_BEAT_LEVEL
type MsgMhfUpdateBeatLevel struct {
	AckHandle uint32
	Unk1      uint32
	Unk2      uint32
	Data1     []int32
	Data2     []int32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateBeatLevel) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_BEAT_LEVEL
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateBeatLevel) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	for i := 0; i < 16; i++ {
		m.Data1 = append(m.Data1, bf.ReadInt32())
	}
	for i := 0; i < 16; i++ {
		m.Data2 = append(m.Data2, bf.ReadInt32())
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateBeatLevel) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
