package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfSavePlateMyset represents the MSG_MHF_SAVE_PLATE_MYSET
type MsgMhfSavePlateMyset struct {
	AckHandle      uint32
	DataSize       uint32
	RawDataPayload []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSavePlateMyset) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_PLATE_MYSET
}

// Parse parses the packet from binary
func (m *MsgMhfSavePlateMyset) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.DataSize = bf.ReadUint32()
	m.RawDataPayload = bf.ReadBytes(uint(m.DataSize))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSavePlateMyset) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
