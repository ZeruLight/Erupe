package mhfpacket

import ( 
 "errors" 

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgMhfRegistGuildTresure represents the MSG_MHF_REGIST_GUILD_TRESURE
type MsgMhfRegistGuildTresure struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfRegistGuildTresure) Opcode() network.PacketID {
	return network.MSG_MHF_REGIST_GUILD_TRESURE
}

// Parse parses the packet from binary
func (m *MsgMhfRegistGuildTresure) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfRegistGuildTresure) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
