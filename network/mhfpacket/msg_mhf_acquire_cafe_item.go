package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireCafeItem represents the MSG_MHF_ACQUIRE_CAFE_ITEM
type MsgMhfAcquireCafeItem struct {
	AckHandle uint32

	// Valid sizes, not sure if [un]signed.
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
	Unk3 uint32
	Unk4 uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireCafeItem) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_CAFE_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireCafeItem) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Unk1 = bf.ReadUint16()
	m.Unk2 = bf.ReadUint16()
	m.Unk3 = bf.ReadUint32()
	m.Unk4 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireCafeItem) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
