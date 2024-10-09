package mhfpacket

import (
	"errors"

	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/utils/byteframe"
)

// MsgMhfPostBoostTimeLimit represents the MSG_MHF_POST_BOOST_TIME_LIMIT
type MsgMhfPostBoostTimeLimit struct {
	AckHandle  uint32
	Expiration uint32
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfPostBoostTimeLimit) Opcode() network.PacketID {
	return network.MSG_MHF_POST_BOOST_TIME_LIMIT
}

// Parse parses the packet from binary
func (m *MsgMhfPostBoostTimeLimit) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Expiration = bf.ReadUint32()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfPostBoostTimeLimit) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
