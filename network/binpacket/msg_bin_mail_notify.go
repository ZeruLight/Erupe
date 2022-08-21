package binpacket

import (
	"errors"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network"
)

type MsgBinMailNotify struct {
	SenderName string
}

func (m MsgBinMailNotify) Parse(bf *byteframe.ByteFrame) error {
	return errors.New("NOT IMPLEMENTED")
}

func (m MsgBinMailNotify) Build(bf *byteframe.ByteFrame) error {
	bf.WriteUint8(0x01) // Unk
	byteName, _ := stringsupport.ConvertUTF8ToShiftJIS(m.SenderName)

	bf.WriteBytes(byteName)
	bf.WriteBytes(make([]byte, 21-len(byteName)))

	return nil
}

func (m MsgBinMailNotify) Opcode() network.PacketID {
	return network.MSG_SYS_CASTED_BINARY
}
