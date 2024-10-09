package binpacket

import (
	"erupe-ce/network"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/stringsupport"
)

type MsgBinMailNotify struct {
	SenderName string
}

func (m MsgBinMailNotify) Parse(bf *byteframe.ByteFrame) error {
	panic("implement me")
}

func (m MsgBinMailNotify) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint8(0x01) // Unk
	bf.WriteBytes(stringsupport.PaddedString(m.SenderName, 21, true))
	return nil
}

func (m MsgBinMailNotify) Opcode() network.PacketID {
	return network.MSG_SYS_CASTED_BINARY
}
