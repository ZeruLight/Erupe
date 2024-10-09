package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfLoadScenarioData represents the MSG_MHF_LOAD_SCENARIO_DATA
type MsgMhfLoadScenarioData struct {
	AckHandle uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfLoadScenarioData) Opcode() network.PacketID {
	return network.MSG_MHF_LOAD_SCENARIO_DATA
}

// Parse parses the packet from binary
func (m *MsgMhfLoadScenarioData) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfLoadScenarioData) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
