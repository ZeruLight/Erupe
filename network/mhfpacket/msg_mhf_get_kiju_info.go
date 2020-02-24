package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetKijuInfo represents the MSG_MHF_GET_KIJU_INFO
type MsgMhfGetKijuInfo struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetKijuInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_KIJU_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetKijuInfo) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetKijuInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
