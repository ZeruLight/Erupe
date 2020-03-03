package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysOperateRegister represents the MSG_SYS_OPERATE_REGISTER
type MsgSysOperateRegister struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysOperateRegister) Opcode() network.PacketID {
	return network.MSG_SYS_OPERATE_REGISTER
}

// Parse parses the packet from binary
func (m *MsgSysOperateRegister) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgSysOperateRegister) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
