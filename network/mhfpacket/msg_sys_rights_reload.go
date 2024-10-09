package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysRightsReload represents the MSG_SYS_RIGHTS_RELOAD
type MsgSysRightsReload struct {
	AckHandle uint32
	Unk0      []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysRightsReload) Opcode() network.PacketID {
	return network.MSG_SYS_RIGHTS_RELOAD
}

// Parse parses the packet from binary
func (m *MsgSysRightsReload) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadBytes(uint(bf.ReadUint8()))
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysRightsReload) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
