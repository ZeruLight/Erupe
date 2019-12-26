package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfGetGuildTresureSouvenir represents the MSG_MHF_GET_GUILD_TRESURE_SOUVENIR
type MsgMhfGetGuildTresureSouvenir struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetGuildTresureSouvenir) Opcode() network.PacketID {
	return network.MSG_MHF_GET_GUILD_TRESURE_SOUVENIR
}

// Parse parses the packet from binary
func (m *MsgMhfGetGuildTresureSouvenir) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetGuildTresureSouvenir) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}