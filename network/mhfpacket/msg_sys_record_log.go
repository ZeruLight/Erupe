package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysRecordLog represents the MSG_SYS_RECORD_LOG
type MsgSysRecordLog struct {
	AckHandle uint32
	Unk0      uint32
	Unk1      uint32
	Data      []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysRecordLog) Opcode() network.PacketID {
	return network.MSG_SYS_RECORD_LOG
}

// Parse parses the packet from binary
func (m *MsgSysRecordLog) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	bf.ReadUint16() // Zeroed
	size := bf.ReadUint16()
	m.Unk1 = bf.ReadUint32()
	m.Data = bf.ReadBytes(uint(size))
	return nil

}

// Build builds a binary packet from the current data.
func (m *MsgSysRecordLog) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
