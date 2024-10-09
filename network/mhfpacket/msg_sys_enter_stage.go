package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgSysEnterStage represents the MSG_SYS_ENTER_STAGE
type MsgSysEnterStage struct {
	AckHandle uint32
	Unk       bool
	StageID   string
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysEnterStage) Opcode() network.PacketID {
	return network.MSG_SYS_ENTER_STAGE
}

// Parse parses the packet from binary
func (m *MsgSysEnterStage) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk = bf.ReadBool() // IsQuest?
	bf.ReadUint8()        // Length StageID
	m.StageID = string(bf.ReadNullTerminatedBytes())
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysEnterStage) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
