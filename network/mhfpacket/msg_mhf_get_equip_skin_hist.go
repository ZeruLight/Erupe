package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetEquipSkinHist represents the MSG_MHF_GET_EQUIP_SKIN_HIST
type MsgMhfGetEquipSkinHist struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetEquipSkinHist) Opcode() network.PacketID {
	return network.MSG_MHF_GET_EQUIP_SKIN_HIST
}

// Parse parses the packet from binary
func (m *MsgMhfGetEquipSkinHist) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetEquipSkinHist) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}