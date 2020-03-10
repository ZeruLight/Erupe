package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysLoadRegister represents the MSG_SYS_LOAD_REGISTER
type MsgSysLoadRegister struct{
	AckHandle uint32
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysLoadRegister) Opcode() network.PacketID {
	return network.MSG_SYS_LOAD_REGISTER
}

// Parse parses the packet from binary
func (m *MsgSysLoadRegister) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	m.Unk2 = bf.ReadUint16()
	m.Unk3 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysLoadRegister) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
