package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfUpdateInterior represents the MSG_MHF_UPDATE_INTERIOR
type MsgMhfUpdateInterior struct {
	AckHandle    uint32
	InteriorData []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateInterior) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_INTERIOR
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateInterior) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.InteriorData = bf.ReadBytes(20)
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateInterior) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
