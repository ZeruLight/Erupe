package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

type Gook struct {
	Exists    bool
	Index     uint32
	Type      uint16
	Data      []byte
	Birthday1 uint32
	Birthday2 uint32
	NameLen   uint8
	Name      []byte
}

// MsgMhfUpdateGuacot represents the MSG_MHF_UPDATE_GUACOT
type MsgMhfUpdateGuacot struct {
	AckHandle  uint32
	EntryCount uint16
	Unk0       uint16 // Hardcoded 0 in binary
	Gooks      []Gook
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateGuacot) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_GUACOT
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateGuacot) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.EntryCount = bf.ReadUint16()
	m.Unk0 = bf.ReadUint16()
	for i := 0; i < int(m.EntryCount); i++ {
		e := Gook{}
		e.Index = bf.ReadUint32()
		e.Type = bf.ReadUint16()
		e.Data = bf.ReadBytes(42)
		e.Birthday1 = bf.ReadUint32()
		e.Birthday2 = bf.ReadUint32()
		e.NameLen = bf.ReadUint8()
		e.Name = bf.ReadBytes(uint(e.NameLen))
		if e.Type > 0 {
			e.Exists = true
		} else {
			e.Exists = false
		}
		m.Gooks = append(m.Gooks, e)
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateGuacot) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
