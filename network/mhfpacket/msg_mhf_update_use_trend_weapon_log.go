package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfUpdateUseTrendWeaponLog represents the MSG_MHF_UPDATE_USE_TREND_WEAPON_LOG
type MsgMhfUpdateUseTrendWeaponLog struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateUseTrendWeaponLog) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_USE_TREND_WEAPON_LOG
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateUseTrendWeaponLog) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateUseTrendWeaponLog) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}