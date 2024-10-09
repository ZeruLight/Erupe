package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfCreateMercenary represents the MSG_MHF_CREATE_MERCENARY
type MsgMhfCreateMercenary struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfCreateMercenary) Opcode() network.PacketID {
	return network.MSG_MHF_CREATE_MERCENARY
}

// Parse parses the packet from binary
func (m *MsgMhfCreateMercenary) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfCreateMercenary) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
