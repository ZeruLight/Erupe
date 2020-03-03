package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetUdTacticsFollower represents the MSG_MHF_GET_UD_TACTICS_FOLLOWER
type MsgMhfGetUdTacticsFollower struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsFollower) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_FOLLOWER
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsFollower) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsFollower) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
