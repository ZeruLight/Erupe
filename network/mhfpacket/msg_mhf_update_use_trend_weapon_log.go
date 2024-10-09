package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfUpdateUseTrendWeaponLog represents the MSG_MHF_UPDATE_USE_TREND_WEAPON_LOG
type MsgMhfUpdateUseTrendWeaponLog struct {
	AckHandle  uint32
	WeaponType uint8
	WeaponID   uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateUseTrendWeaponLog) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_USE_TREND_WEAPON_LOG
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateUseTrendWeaponLog) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.WeaponType = bf.ReadUint8()
	m.WeaponID = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateUseTrendWeaponLog) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
