package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetTowerInfo represents the MSG_MHF_GET_TOWER_INFO
type MsgMhfGetTowerInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetTowerInfo) Opcode() network.PacketID {
	return network.MSG_MHF_GET_TOWER_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfGetTowerInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetTowerInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}