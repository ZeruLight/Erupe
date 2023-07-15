package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func handleMsgMhfEnumerateCampaign(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateCampaign)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	bf.WriteUint8(0)
	bf.WriteUint8(0)
	bf.WriteUint8(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfStateCampaign(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateCampaign)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfApplyCampaign(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfApplyCampaign)
	doAckSimpleFail(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfEnumerateItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateItem)
	items := []struct {
		Unk0 uint32
		Unk1 uint16
		Unk2 uint16
		Unk3 uint16
		Unk4 uint32
		Unk5 uint32
	}{}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(items)))
	for _, item := range items {
		bf.WriteUint32(item.Unk0)
		bf.WriteUint16(item.Unk1)
		bf.WriteUint16(item.Unk2)
		bf.WriteUint16(item.Unk3)
		bf.WriteUint32(item.Unk4)
		bf.WriteUint32(item.Unk5)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireItem)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
