package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAcquireItem represents the MSG_MHF_ACQUIRE_ITEM
type MsgMhfAcquireItem struct {
	AckHandle uint32
	Unk0      uint16
	Length    uint16
	Unk1      []uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireItem) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireItem) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	m.Length = bf.ReadUint16()
	for i := 0; i < int(m.Length); i++ {
		m.Unk1 = append(m.Unk1, bf.ReadUint32())
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireItem) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
