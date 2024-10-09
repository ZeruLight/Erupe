package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfReceiveGachaItem represents the MSG_MHF_RECEIVE_GACHA_ITEM
type MsgMhfReceiveGachaItem struct {
	AckHandle uint32
	Max       uint8
	Freeze    bool
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReceiveGachaItem) Opcode() network.PacketID {
	return network.MSG_MHF_RECEIVE_GACHA_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfReceiveGachaItem) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Max = bf.ReadUint8()
	m.Freeze = bf.ReadBool()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReceiveGachaItem) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
