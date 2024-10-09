package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfTransferItem represents the MSG_MHF_TRANSFER_ITEM
type MsgMhfTransferItem struct {
	AckHandle uint32
	// looking at packets, these were static across sessions and did not actually
	// correlate with any item IDs that would make sense to get after quests so
	// I have no idea what this actually does
	Unk0 uint32
	Unk1 uint8
	Unk2 uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfTransferItem) Opcode() network.PacketID {
	return network.MSG_MHF_TRANSFER_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfTransferItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Unk1 = bf.ReadUint8()
	bf.ReadUint8() // Zeroed
	m.Unk2 = bf.ReadUint16()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfTransferItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
