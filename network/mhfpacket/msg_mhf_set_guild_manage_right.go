package mhfpacket

import ( 
 "errors" 

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgMhfSetGuildManageRight represents the MSG_MHF_SET_GUILD_MANAGE_RIGHT
type MsgMhfSetGuildManageRight struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSetGuildManageRight) Opcode() network.PacketID {
	return network.MSG_MHF_SET_GUILD_MANAGE_RIGHT
}

// Parse parses the packet from binary
func (m *MsgMhfSetGuildManageRight) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSetGuildManageRight) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
