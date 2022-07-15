package mhfpacket

import ( 
 "errors" 

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgSysReserve20A represents the MSG_SYS_reserve20A
type MsgSysReserve20A struct{}

// Opcode returns the ID associated with this packet type.
func (m *MsgSysReserve20A) Opcode() network.PacketID {
	return network.MSG_SYS_reserve20A
}

// Parse parses the packet from binary
func (m *MsgSysReserve20A) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}

// Build builds a binary packet from the current data.
func (m *MsgSysReserve20A) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
