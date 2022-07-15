package mhfpacket

import ( 
 "errors" 

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgMhfEntryRookieGuild represents the MSG_MHF_ENTRY_ROOKIE_GUILD
type MsgMhfEntryRookieGuild struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEntryRookieGuild) Opcode() network.PacketID {
	return network.MSG_MHF_ENTRY_ROOKIE_GUILD
}

// Parse parses the packet from binary
func (m *MsgMhfEntryRookieGuild) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEntryRookieGuild) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
