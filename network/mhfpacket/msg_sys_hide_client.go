package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysHideClient represents the MSG_SYS_HIDE_CLIENT
type MsgSysHideClient struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysHideClient) Opcode() network.PacketID {
	return network.MSG_SYS_HIDE_CLIENT
}

// Parse parses the packet from binary
func (m *MsgSysHideClient) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysHideClient) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}