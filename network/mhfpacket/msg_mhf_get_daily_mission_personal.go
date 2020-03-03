package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetDailyMissionPersonal represents the MSG_MHF_GET_DAILY_MISSION_PERSONAL
type MsgMhfGetDailyMissionPersonal struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetDailyMissionPersonal) Opcode() network.PacketID {
	return network.MSG_MHF_GET_DAILY_MISSION_PERSONAL
}

// Parse parses the packet from binary
func (m *MsgMhfGetDailyMissionPersonal) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetDailyMissionPersonal) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
