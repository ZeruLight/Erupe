package mhfpacket

import (
	"errors"
	_config "erupe-ce/config"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysCreateSemaphore represents the MSG_SYS_CREATE_SEMAPHORE
type MsgSysCreateSemaphore struct {
	AckHandle   uint32
	Unk0        uint16
	PlayerCount uint8
	SemaphoreID string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysCreateSemaphore) Opcode() network.PacketID {
	return network.MSG_SYS_CREATE_SEMAPHORE
}

// Parse parses the packet from binary
func (m *MsgSysCreateSemaphore) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()
	if _config.ErupeConfig.ClientID >= _config.S7 { // Assuming this was added with Ravi?
		m.PlayerCount = bf.ReadUint8()
	}
	bf.ReadUint8() // SemaphoreID length
	m.SemaphoreID = string(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysCreateSemaphore) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
