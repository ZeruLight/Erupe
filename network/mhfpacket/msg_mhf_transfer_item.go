package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfTransferItem represents the MSG_MHF_TRANSFER_ITEM
type MsgMhfTransferItem struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfTransferItem) Opcode() network.PacketID {
	return network.MSG_MHF_TRANSFER_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfTransferItem) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfTransferItem) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}