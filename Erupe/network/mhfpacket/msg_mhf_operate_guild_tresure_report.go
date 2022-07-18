package mhfpacket

import ( 
 "errors" 

 	"erupe-ce/network/clientctx"
	"erupe-ce/network"
	"erupe-ce/common/byteframe"
)

// MsgMhfOperateGuildTresureReport represents the MSG_MHF_OPERATE_GUILD_TRESURE_REPORT
type MsgMhfOperateGuildTresureReport struct{
  AckHandle uint32
  HuntID uint32
  Unk uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOperateGuildTresureReport) Opcode() network.PacketID {
	return network.MSG_MHF_OPERATE_GUILD_TRESURE_REPORT
}

// Parse parses the packet from binary
func (m *MsgMhfOperateGuildTresureReport) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
  m.AckHandle = bf.ReadUint32()
  m.HuntID = bf.ReadUint32()
  m.Unk = bf.ReadUint16()
  return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOperateGuildTresureReport) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
