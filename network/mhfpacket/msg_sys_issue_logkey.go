package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysIssueLogkey represents the MSG_SYS_ISSUE_LOGKEY
type MsgSysIssueLogkey struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysIssueLogkey) Opcode() network.PacketID {
	return network.MSG_SYS_ISSUE_LOGKEY
}

// Parse parses the packet from binary
func (m *MsgSysIssueLogkey) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysIssueLogkey) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}