package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfLoadLegendDispatch represents the MSG_MHF_LOAD_LEGEND_DISPATCH
type MsgMhfLoadLegendDispatch struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadLegendDispatch) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_LEGEND_DISPATCH
}

// Parse parses the packet from binary
func (m *MsgMhfLoadLegendDispatch) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadLegendDispatch) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}