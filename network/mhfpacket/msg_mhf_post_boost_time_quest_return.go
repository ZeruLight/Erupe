package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfPostBoostTimeQuestReturn represents the MSG_MHF_POST_BOOST_TIME_QUEST_RETURN
type MsgMhfPostBoostTimeQuestReturn struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostBoostTimeQuestReturn) Opcode() network.PacketID {
	return network.MSG_MHF_POST_BOOST_TIME_QUEST_RETURN
}

// Parse parses the packet from binary
func (m *MsgMhfPostBoostTimeQuestReturn) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostBoostTimeQuestReturn) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
