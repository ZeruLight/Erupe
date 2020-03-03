package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetMyhouseInfo represents the MSG_MHF_GET_MYHOUSE_INFO
type MsgMhfGetMyhouseInfo struct{
	AckHandle      uint32
	Unk0       uint32
	Unk1       uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetMyhouseInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_MYHOUSE_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetMyhouseInfo) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetMyhouseInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
