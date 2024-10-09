package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfUnreserveSrg represents the MSG_MHF_UNRESERVE_SRG
type MsgMhfUnreserveSrg struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUnreserveSrg) Opcode() network.PacketID {
	return network.MSG_MHF_UNRESERVE_SRG
}

// Parse parses the packet from binary
func (m *MsgMhfUnreserveSrg) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUnreserveSrg) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
