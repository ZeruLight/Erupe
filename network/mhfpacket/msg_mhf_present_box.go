package mhfpacket

import (
	"errors"
	"fmt"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfPresentBox represents the MSG_MHF_PRESENT_BOX
type MsgMhfPresentBox struct {
	AckHandle    uint32
	Unk0         uint32
	Operation    uint32
	PresentCount uint32
	Unk3         uint32
	Unk4         uint32
	Unk5         uint32
	Unk6         uint32
	PresentType  []uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPresentBox) Opcode() network.PacketID {
	return network.MSG_MHF_PRESENT_BOX
}

// Parse parses the packet from binary
func (m *MsgMhfPresentBox) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint32()
	m.Operation = bf.ReadUint32()
	m.PresentCount = bf.ReadUint32()
	m.Unk3 = bf.ReadUint32()
	m.Unk4 = bf.ReadUint32()
	m.Unk5 = bf.ReadUint32()
	m.Unk6 = bf.ReadUint32()
	for i := uint32(0); i < m.PresentCount; i++ {
		m.PresentType = append(m.PresentType, bf.ReadUint32())
	}
	fmt.Printf("MsgMhfPresentBox: Unk0:[%d] Unk1:[%d] Unk2:[%d] Unk3:[%d] Unk4:[%d] Unk5:[%d] Unk6:[%d] \n\n", m.Unk0, m.Operation, m.PresentCount, m.Unk3, m.Unk4, m.Unk5, m.Unk6)
	for _, mdata := range m.PresentType {
		fmt.Printf("MsgMhfPresentBox: PresentType: [%d] \n", mdata)

	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPresentBox) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
