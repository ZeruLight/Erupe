package mhfpacket

import (
 "errors"

 	"github.com/Solenataris/Erupe/network/clientctx"
	"github.com/Solenataris/Erupe/network"
	"github.com/Andoryuuta/byteframe"
)

// MsgMhfSendMail represents the MSG_MHF_SEND_MAIL
type MsgMhfSendMail struct {
  AckHandle     uint32
  RecipientID   uint32
  SubjectLength uint16
  BodyLength    uint16
  Quantity      uint32
  ItemID        uint16
  Subject       []byte
  Body          []byte
}

// Opcode returns the ID associated with this packet type.
func (m *MsgMhfSendMail) Opcode() network.PacketID {
	return network.MSG_MHF_SEND_MAIL
}

// Parse parses the packet from binary
func (m *MsgMhfSendMail) Parse(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
  m.AckHandle = bf.ReadUint32()
  m.RecipientID = bf.ReadUint32()
  m.SubjectLength = bf.ReadUint16()
  m.BodyLength = bf.ReadUint16()
  m.Quantity = bf.ReadUint32()
  m.ItemID = bf.ReadUint16()
  m.Subject = bf.ReadNullTerminatedBytes()
  m.Body = bf.ReadNullTerminatedBytes()
  return nil
}

// Build builds a binary packet from the current data.
func (m *MsgMhfSendMail) Build(bf *byteframe.ByteFrame, ctx *clientctx.ClientContext) error {
	return errors.New("NOT IMPLEMENTED")
}
