package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfReadMail represents the MSG_MHF_READ_MAIL
type MsgMhfReadMail struct {
	AckHandle uint32

	// AccIndex is incremented for each mail in the list
	// The index is persistent for game session, reopening the mail list
	// will continue from the last index + 1
	AccIndex uint8

	// This is the index within the current mail list
	Index uint8
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfReadMail) Opcode() network.PacketID {
	return network.MSG_MHF_READ_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfReadMail) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.AccIndex = bf.ReadUint8()
	m.Index = bf.ReadUint8()
	bf.ReadUint16() // Zeroed
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfReadMail) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
