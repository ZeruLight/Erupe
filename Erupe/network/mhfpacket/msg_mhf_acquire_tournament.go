package mhfpacket

import ( 
 "errors" 

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgMhfAcquireTournament represents the MSG_MHF_ACQUIRE_TOURNAMENT
type MsgMhfAcquireTournament struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfAcquireTournament) Opcode() network.PacketID {
	return network.MSG_MHF_ACQUIRE_TOURNAMENT
}

// Parse parses the packet from binary
func (m *MsgMhfAcquireTournament) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfAcquireTournament) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
