package mhfpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfApplyBbsArticle represents the MSG_MHF_APPLY_BBS_ARTICLE
type MsgMhfApplyBbsArticle struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfApplyBbsArticle) Opcode() network.PacketID {
	return network.MSG_MHF_APPLY_BBS_ARTICLE
}

// Parse parses the packet from binary
func (m *MsgMhfApplyBbsArticle) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfApplyBbsArticle) Build(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}
