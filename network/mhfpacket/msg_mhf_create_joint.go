package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/stringsupport"
)

// MsgMhfCreateJoint represents the MSG_MHF_CREATE_JOINT
type MsgMhfCreateJoint struct {
	AckHandle uint32
	GuildID   uint32
	Name      string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCreateJoint) Opcode() network.PacketID {
	return network.MSG_MHF_CREATE_JOINT
}

// Parse parses the packet from binary
func (m *MsgMhfCreateJoint) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	bf.ReadUint16() // Zeroed
	bf.ReadUint16() // Name length
	m.Name = stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCreateJoint) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
