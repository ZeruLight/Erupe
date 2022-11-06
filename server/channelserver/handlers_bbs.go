package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/common/token"
	"erupe-ce/network/mhfpacket"
)

func handleMsgMhfGetBbsUserStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBbsUserStatus)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(200)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetBbsSnsStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBbsSnsStatus)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(200)
	bf.WriteUint32(401)
	bf.WriteUint32(401)
	bf.WriteUint32(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfApplyBbsArticle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfApplyBbsArticle)
	bf := byteframe.NewByteFrame()
	articleToken := token.Generate(40)
	bf.WriteUint32(200)
	bf.WriteUint32(80)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteBytes(stringsupport.PaddedString(articleToken, 64, false))
	bf.WriteBytes(stringsupport.PaddedString(s.server.erupeConfig.ScreenshotAPIURL, 64, false))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
