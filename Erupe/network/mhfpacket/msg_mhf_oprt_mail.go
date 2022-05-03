package mhfpacket

import (
 "errors"

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

type OperateMailOperation uint8

const (
	OPERATE_MAIL_DELETE = 0x01
  OPERATE_MAIL_ACQUIRE_ITEM = 0x05
)

// MsgMhfOprtMail represents the MSG_MHF_OPRT_MAIL
type MsgMhfOprtMail struct {
  AckHandle uint32
	AccIndex  uint8
	Index     uint8
	Operation OperateMailOperation
  Unk0      uint8
  Data      []byte
  Amount    uint16
  ItemID    uint16
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfOprtMail) Opcode() network.PacketID {
	return network.MSG_MHF_OPRT_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfOprtMail) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
  m.AckHandle = bf.ReadUint32()
	m.AccIndex = bf.ReadUint8()
	m.Index = bf.ReadUint8()
	m.Operation = OperateMailOperation(bf.ReadUint8())
  m.Unk0 = bf.ReadUint8()
  switch m.Operation {
  case OPERATE_MAIL_ACQUIRE_ITEM:
    m.Amount = bf.ReadUint16()
    m.ItemID = bf.ReadUint16()
  }
	return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfOprtMail) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
