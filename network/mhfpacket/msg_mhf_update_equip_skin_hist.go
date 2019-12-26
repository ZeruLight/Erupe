package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateEquipSkinHist represents the MSG_MHF_UPDATE_EQUIP_SKIN_HIST
type MsgMhfUpdateEquipSkinHist struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateEquipSkinHist) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_EQUIP_SKIN_HIST
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateEquipSkinHist) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateEquipSkinHist) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}