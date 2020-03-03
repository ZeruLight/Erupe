package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetBbsUserStatus represents the MSG_MHF_GET_BBS_USER_STATUS
type MsgMhfGetBbsUserStatus struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetBbsUserStatus) Opcode() network.PacketID {
	return network.MSG_MHF_GET_BBS_USER_STATUS
}

// Parse parses the packet from binary
func (m *MsgMhfGetBbsUserStatus) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetBbsUserStatus) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
