package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/common/token"
	"erupe-ce/network/mhfpacket"
)

func handleMsgMhfGetBbsUserStatus(s *Session, p mhfpacket.MHFPacket) {
	//Post Screenshot pauses till this succeedes
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
	bf.WriteUint32(s.server.erupeConfig.Screenshots.Port)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteBytes(stringsupport.PaddedString(articleToken, 64, false))
	bf.WriteBytes(stringsupport.PaddedString(s.server.erupeConfig.Screenshots.Host, 64, false))

	if s.server.erupeConfig.SaveDumps.Enabled && s.server.erupeConfig.Discord.Enabled {
		messageId := s.server.DiscordScreenShotSend(pkt.Name, pkt.Title, pkt.Description) // TODO: send and get back message id store in db

		_, err := s.server.db.Exec("INSERT INTO public.screenshots (article_id,discord_message_id,char_id,title,description) VALUES ($1,$2,$3,$4,$5)", articleToken, messageId, s.charID, pkt.Title, pkt.Description)
		if err != nil {
			doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		} else {

			doAckBufSucceed(s, pkt.AckHandle, bf.Data())
			s.server.BroadcastChatMessage("Screenshot has been sent to discord")

		}

	} else if s.server.erupeConfig.SaveDumps.Enabled {
		_, err := s.server.db.Exec("INSERT INTO public.screenshots (article_id,char_id,title,description) VALUES ($1,$2,$3,$4)", articleToken, s.charID, pkt.Title, pkt.Description)
		if err != nil {
			doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		} else {
			s.server.BroadcastChatMessage("Screenshot has been sent to server")
			doAckBufSucceed(s, pkt.AckHandle, bf.Data())

		}
	} else {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		s.server.BroadcastChatMessage("No destination for screenshots have been configured by the host")
	}
}
