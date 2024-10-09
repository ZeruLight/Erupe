package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfLoadHunterNavi represents the MSG_MHF_LOAD_HUNTER_NAVI
type MsgMhfLoadHunterNavi struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadHunterNavi) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_HUNTER_NAVI
}

// Parse parses the packet from binary
func (m *MsgMhfLoadHunterNavi) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadHunterNavi) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
