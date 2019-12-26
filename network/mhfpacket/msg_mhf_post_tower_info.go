package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPostTowerInfo represents the MSG_MHF_POST_TOWER_INFO
type MsgMhfPostTowerInfo struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostTowerInfo) Opcode() network.PacketID {
	return network.MSG_MHF_POST_TOWER_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfPostTowerInfo) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostTowerInfo) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}