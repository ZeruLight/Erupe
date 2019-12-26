package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfStampcardStamp represents the MSG_MHF_STAMPCARD_STAMP
type MsgMhfStampcardStamp struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfStampcardStamp) Opcode() network.PacketID {
	return network.MSG_MHF_STAMPCARD_STAMP
}

// Parse parses the packet from binary
func (m *MsgMhfStampcardStamp) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfStampcardStamp) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}