package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSaveMercenary represents the MSG_MHF_SAVE_MERCENARY
type MsgMhfSaveMercenary struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSaveMercenary) Opcode() network.PacketID {
	return network.MSG_MHF_SAVE_MERCENARY
}

// Parse parses the packet from binary
func (m *MsgMhfSaveMercenary) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSaveMercenary) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}