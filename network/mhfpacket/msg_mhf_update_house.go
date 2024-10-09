package mhfpacket

import (
	"errors"
	"erupe-ce/utils/stringsupport"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfUpdateHouse represents the MSG_MHF_UPDATE_HOUSE
type MsgMhfUpdateHouse struct {
	AckHandle uint32
	State     uint8
	Unk1      uint8 // Always 0x01
	Password  string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateHouse) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_HOUSE
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateHouse) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.State = bf.ReadUint8()
	m.Unk1 = bf.ReadUint8()
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Password length
	m.Password = stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateHouse) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
