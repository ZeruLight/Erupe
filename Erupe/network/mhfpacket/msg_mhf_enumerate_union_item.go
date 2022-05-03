package mhfpacket

import ( 
 "errors" 

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfEnumerateUnionItem represents the MSG_MHF_ENUMERATE_UNION_ITEM
type MsgMhfEnumerateUnionItem struct {
	AckHandle uint32
	Unk0      uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfEnumerateUnionItem) Opcode() network.PacketID {
	return network.MSG_MHF_ENUMERATE_UNION_ITEM
}

// Parse parses the packet from binary
func (m *MsgMhfEnumerateUnionItem) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	m.Unk0 = bf.ReadUint16()

	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfEnumerateUnionItem) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
