package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfStampcardStamp represents the MSG_MHF_STAMPCARD_STAMP
type MsgMhfStampcardStamp struct{
	// probably not actual format, just lined up neatly to an example packet
	AckHandle uint32
	Unk0 uint32
	Unk1 uint32
	Unk2 uint32
	Unk3 uint32
	Unk4 uint32
	Unk5 uint32
	Unk6 uint32
	Unk7 uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStampcardStamp) Opcode() network.PacketID {
	return network.MSG_MHF_STAMPCARD_STAMP
}

// Parse parses the packet from binary
func (m *MsgMhfStampcardStamp) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.Unk2 = bf.ReadUint32()
	m.Unk3 = bf.ReadUint32()
	m.Unk4 = bf.ReadUint32()
	m.Unk5 = bf.ReadUint32()
	m.Unk6 = bf.ReadUint32()
	m.Unk7 = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStampcardStamp) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
