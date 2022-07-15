package mhfpacket

import ( 
 "errors" 

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgMhfGetUdTacticsRemainingPoint represents the MSG_MHF_GET_UD_TACTICS_REMAINING_POINT
type MsgMhfGetUdTacticsRemainingPoint struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfGetUdTacticsRemainingPoint) Opcode() network.PacketID {
	return network.MSG_MHF_GET_UD_TACTICS_REMAINING_POINT
}

// Parse parses the packet from binary
func (m *MsgMhfGetUdTacticsRemainingPoint) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgMhfGetUdTacticsRemainingPoint) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
