package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfAcquireMonthlyReward represents the MSG_MHF_ACQUIRE_MONTHLY_REWARD
type MsgMhfAcquireMonthlyReward struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireMonthlyReward) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_MONTHLY_REWARD
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireMonthlyReward) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireMonthlyReward) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}