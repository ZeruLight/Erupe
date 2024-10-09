package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysSetStagePass represents the MSG_SYS_SET_STAGE_PASS
type MsgSysSetStagePass struct {
	Password string // NULL-terminated string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysSetStagePass) Opcode() network.PacketID {
	return network.MSG_SYS_SET_STAGE_PASS
}

// Parse parses the packet from binary
func (m *MsgSysSetStagePass) Parse(bf *byteframe.ByteFrame) error {
	bf.ReadUint8() // Zeroed
	bf.ReadUint8() // Password length
	m.Password = string(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysSetStagePass) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
