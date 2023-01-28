package mhfpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

// MsgMhfContractMercenary represents the MSG_MHF_CONTRACT_MERCENARY
type MsgMhfContractMercenary struct {
	AckHandle  uint32
	PactMercID uint32
	CID        uint32
	Unk        bool
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfContractMercenary) Opcode() network.PacketID {
	return network.MSG_MHF_CONTRACT_MERCENARY
}

// Parse parses the packet from binary
func (m *MsgMhfContractMercenary) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.PactMercID = bf.ReadUint32()
	m.CID = bf.ReadUint32()
	m.Unk = bf.ReadBool()
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfContractMercenary) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
