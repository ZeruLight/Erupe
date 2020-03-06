package binpacket

import (
	"github.com/Andoryuuta/Erupe/network"
	"github.com/Andoryuuta/Erupe/network/mhfpacket"
	"github.com/Andoryuuta/byteframe"
)

type PartyJoinType uint8

const (
	JoinedLocalParty PartyJoinType = 0x01
	JoinedYourParty  PartyJoinType = 0x04
)

type MsgBinPlayerJoinedParty struct {
	CharID        uint32
	PartyJoinType PartyJoinType
}

// Opcode returns the ID associated with this packet type.
func (m *MsgBinPlayerJoinedParty) Opcode() network.PacketID {
	return network.MSG_SYS_CASTED_BINARY
}

func (m *MsgBinPlayerJoinedParty) Parse(bf *byteframe.ByteFrame) error {
	panic("Not implemented")
}

// Build builds a binary packet from the current data.
func (m *MsgBinPlayerJoinedParty) Build(bf *byteframe.ByteFrame) error {
	payload := byteframe.NewByteFrame()

	payload.WriteUint16(0x2)
	payload.WriteUint8(uint8(m.PartyJoinType))
	payload.WriteBytes([]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	pkt := &mhfpacket.MsgSysCastedBinary{
		CharID:         m.CharID,
		Type0:          0x03,
		Type1:          0x03,
		RawDataPayload: payload.Data(),
	}

	return pkt.Build(bf)
}
