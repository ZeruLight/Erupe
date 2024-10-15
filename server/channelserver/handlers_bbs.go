package channelserver

import (
	"erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/stringsupport"
	"erupe-ce/utils/token"
)

// Handler BBS handles all the interactions with the for the screenshot sending to bulitin board functionality. For it to work it requires the API to be hosted somehwere. This implementation supports discord.

// Checks the status of the user to see if they can use Bulitin Board yet
func handleMsgMhfGetBbsUserStatus(s *Session, p mhfpacket.MHFPacket) {
	//Post Screenshot pauses till this succeedes
	pkt := p.(*mhfpacket.MsgMhfGetBbsUserStatus)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(200) //HTTP Status Codes //200 Success //404 You wont be able to post for a certain amount of time after creating your character //401/500 A error occured server side
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

// Checks the status of Bultin Board Server to see if authenticated
func handleMsgMhfGetBbsSnsStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBbsSnsStatus)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(200) //200 Success //4XX Authentication has expired Please re-authenticate //5XX
	bf.WriteUint32(401) //unk http status?
	bf.WriteUint32(401) //unk http status?
	bf.WriteUint32(0)
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

// Tells the game client what host port and gives the bultin board article a token
func handleMsgMhfApplyBbsArticle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfApplyBbsArticle)
	bf := byteframe.NewByteFrame()
	articleToken := token.Generate(40)

	bf.WriteUint32(200) //http status //200 success //4XX An error occured server side
	bf.WriteUint32(config.GetConfig().Screenshots.Port)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteBytes(stringsupport.PaddedString(articleToken, 64, false))
	bf.WriteBytes(stringsupport.PaddedString(config.GetConfig().Screenshots.Host, 64, false))
	//pkt.unk1[3] ==  Changes sometimes?
	if config.GetConfig().Screenshots.Enabled && config.GetConfig().Discord.Enabled {
		s.Server.DiscordScreenShotSend(pkt.Name, pkt.Title, pkt.Description, articleToken)
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())

}
