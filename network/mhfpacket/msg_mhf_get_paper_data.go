package mhfpacket

import (
	"errors"
	"fmt"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfGetPaperData represents the MSG_MHF_GET_PAPER_DATA
type MsgMhfGetPaperData struct {
	// Communicator type, multi-format. This might be valid for only one type.
	AckHandle uint32
	Type      uint32
	Unk1      uint32
	ID        uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetPaperData) Opcode() network.PacketID {
	return network.MSG_MHF_GET_PAPER_DATA
}

// Parse parses the packet from binary
func (m *MsgMhfGetPaperData) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Type = bf.ReadUint32()
	m.Unk1 = bf.ReadUint32()
	m.ID = bf.ReadUint32()
	fmt.Printf("MsgMhfGetPaperData: Type:[%d] Unk1:[%d] ID:[%d] \n\n", m.Type, m.Unk1, m.ID)

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetPaperData) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
