package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgSysCreateAcquireSemaphore represents the MSG_SYS_CREATE_ACQUIRE_SEMAPHORE
type MsgSysCreateAcquireSemaphore struct{
	AckHandle uint32
	Unk0 []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCreateAcquireSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_CREATE_ACQUIRE_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysCreateAcquireSemaphore) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadBytes(19)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysCreateAcquireSemaphore) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
