package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
)

// MsgMhfRegistGuildTresure represents the MSG_MHF_REGIST_GUILD_TRESURE
type MsgMhfRegistGuildTresure struct {
	AckHandle uint32
	Data      []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegistGuildTresure) Opcode() network.PacketID {
	return network.MSG_MHF_REGIST_GUILD_TRESURE
}

// Parse parses the packet from binary
func (m *MsgMhfRegistGuildTresure) Parse(bf *byteframe.ByteFrame) error {
	m.AckHandle = bf.ReadUint32()
	m.Data = bf.ReadBytes(uint(bf.ReadUint16()))
	_ = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegistGuildTresure) Build(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}
