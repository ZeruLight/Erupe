package mhfpacket

import ( 
 "errors" 

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
)

// MsgMhfUpdateMyhouseInfo represents the MSG_MHF_UPDATE_MYHOUSE_INFO
type MsgMhfUpdateMyhouseInfo struct {
	AckHandle uint32
	Unk0      []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfUpdateMyhouseInfo) Opcode() network.PacketID {
	return network.MSG_MHF_UPDATE_MYHOUSE_INFO
}

// Parse parses the packet from binary
func (m *MsgMhfUpdateMyhouseInfo) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	m.AckHandle = bf.ReadUint32()
	if _config.ErupeConfig.RealClientMode == _config.F5 {
		m.Unk0 = bf.ReadBytes(0x122)
	} else {
		m.Unk0 = bf.ReadBytes(0x16A)
	}
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfUpdateMyhouseInfo) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
