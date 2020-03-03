package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfChargeFesta represents the MSG_MHF_CHARGE_FESTA
type MsgMhfChargeFesta struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfChargeFesta) Opcode() network.PacketID {
	return network.MSG_MHF_CHARGE_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfChargeFesta) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfChargeFesta) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
