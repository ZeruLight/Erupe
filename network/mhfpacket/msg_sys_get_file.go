package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type scenarioFileIdentifer struct {
	Unk0 uint8
	Unk1 uint32
	Unk2 uint8
	Unk3 uint8
}

// MsgSysGetFile represents the MSG_SYS_GET_FILE
type MsgSysGetFile struct {
	AckHandle         uint32
	IsScenario        bool
	FilenameLength    uint8
	Filename          string
	ScenarioIdentifer scenarioFileIdentifer
}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysGetFile) Opcode() network.PacketID {
	return network.MSG_SYS_GET_FILE
}

// Parse parses the packet from binary
func (m *MsgSysGetFile) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.IsScenario = bf.ReadBool()
	if m.IsScenario {
		m.ScenarioIdentifer = scenarioFileIdentifer{
			bf.ReadUint8(),
			bf.ReadUint32(),
			bf.ReadUint8(),
			bf.ReadUint8(),
		}
	} else {
		m.FilenameLength = bf.ReadUint8()
		m.Filename = string(bf.ReadBytes(uint(m.FilenameLength)))
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgSysGetFile) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
