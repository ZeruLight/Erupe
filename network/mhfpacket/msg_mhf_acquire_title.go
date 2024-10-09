package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfAcquireTitle represents the MSG_MHF_ACQUIRE_TITLE
type MsgMhfAcquireTitle struct {
	AckHandle uint32
	TitleIDs  []uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireTitle) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_TITLE
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireTitle) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	titles := int(bf.ReadUint16())
	bf.ReadUint16() // Zeroed
	for i := 0; i < titles; i++ {
		m.TitleIDs = append(m.TitleIDs, bf.ReadUint16())
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireTitle) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
