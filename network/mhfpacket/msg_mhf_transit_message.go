package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfTransitMessage represents the MSG_MHF_TRANSIT_MESSAGE
type MsgMhfTransitMessage struct {
	AckHandle   uint32
	Unk0        uint8
	SearchType  uint16
	MessageData []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfTransitMessage) Opcode() network.PacketID {
	return network.MSG_MHF_TRANSIT_MESSAGE
}

// Parse parses the packet from binary
func (m *MsgMhfTransitMessage) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint8()
	bf.ReadUint8() // Zeroed
	m.SearchType = bf.ReadUint16()
	m.MessageData = bf.ReadBytes(uint(bf.ReadUint16()))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfTransitMessage) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
