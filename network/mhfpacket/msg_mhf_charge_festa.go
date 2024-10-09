package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfChargeFesta represents the MSG_MHF_CHARGE_FESTA
type MsgMhfChargeFesta struct {
	AckHandle uint32
	FestaID   uint32
	GuildID   uint32
	Souls     []uint16
	Auto      bool
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfChargeFesta) Opcode() network.PacketID {
	return network.MSG_MHF_CHARGE_FESTA
}

// Parse parses the packet from binary
func (m *MsgMhfChargeFesta) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.FestaID = bf.ReadUint32()
	m.GuildID = bf.ReadUint32()
	for i := bf.ReadUint16(); i > 0; i-- {
		m.Souls = append(m.Souls, bf.ReadUint16())
	}
	m.Auto = bf.ReadBool()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfChargeFesta) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
