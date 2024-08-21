package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"fmt"
)

func handleMsgMhfPostGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostGuildScout)

	guild := GetGuildInfoByCharacterId(s, s.charID)
	if guild.ID == 0 {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	s.server.db.Exec(`INSERT INTO guild_invites (guild_id, character_id, actor_id) VALUES ($1, $2, $3)`, s.charID, pkt.CharID, guild.ID)

	mail := &Mail{
		SenderID:    s.charID,
		RecipientID: pkt.CharID,
		Subject:     s.server.i18n.guild.invite.title,
		Body: fmt.Sprintf(
			s.server.i18n.guild.invite.body,
			guild.Name,
		),
		IsGuildInvite: true,
	}

	mail.Send(s)
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfCancelGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCancelGuildScout)
	s.server.db.Exec(`DELETE FROM guild_invites WHERE id=$1`, pkt.InvitationID)
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAnswerGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAnswerGuildScout)

	guild := GetGuildInfoByCharacterId(s, pkt.LeaderID)
	if guild.ID == 0 {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	bf := byteframe.NewByteFrame()
	var mail []Mail
	var err error
	if pkt.Answer {
		err = guild.AcceptApplication(s, s.charID)
		mail = append(mail, Mail{
			RecipientID:     s.charID,
			Subject:         s.server.i18n.guild.invite.success.title,
			Body:            fmt.Sprintf(s.server.i18n.guild.invite.success.body, guild.Name),
			IsSystemMessage: true,
		})
		mail = append(mail, Mail{
			SenderID:        s.charID,
			RecipientID:     pkt.LeaderID,
			Subject:         s.server.i18n.guild.invite.accepted.title,
			Body:            fmt.Sprintf(s.server.i18n.guild.invite.accepted.body, guild.Name),
			IsSystemMessage: true,
		})
	} else {
		err = guild.RejectApplication(s, s.charID)
		mail = append(mail, Mail{
			RecipientID:     s.charID,
			Subject:         s.server.i18n.guild.invite.rejected.title,
			Body:            fmt.Sprintf(s.server.i18n.guild.invite.rejected.body, guild.Name),
			IsSystemMessage: true,
		})
		mail = append(mail, Mail{
			SenderID:        s.charID,
			RecipientID:     pkt.LeaderID,
			Subject:         s.server.i18n.guild.invite.declined.title,
			Body:            fmt.Sprintf(s.server.i18n.guild.invite.declined.body, guild.Name),
			IsSystemMessage: true,
		})
	}
	if err != nil {
		bf.WriteUint32(7)
		bf.WriteUint32(guild.ID)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		bf.WriteUint32(0)
		bf.WriteUint32(guild.ID)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		for _, m := range mail {
			m.Send(s)
		}
	}
}

func handleMsgMhfGetGuildScoutList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildScoutList)

	guild := GetGuildInfoByCharacterId(s, s.charID)
	if guild.ID == 0 && s.prevGuildID == 0 {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	} else {
		guild = GetGuildInfoByID(s, s.prevGuildID)
		if guild.ID == 0 {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	invites := GetGuildInvites(s, guild.ID)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(len(invites)))
	for _, invite := range invites {
		bf.WriteUint32(invite.ID)
		bf.WriteUint32(invite.ActorID)
		bf.WriteUint32(invite.CharID)
		bf.WriteUint32(uint32(invite.InvitedAt.Unix()))
		bf.WriteUint16(invite.HR)
		bf.WriteUint16(invite.GR)
		bf.WriteBytes(stringsupport.PaddedString(invite.Name, 32, true))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetRejectGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRejectGuildScout)
	var currentStatus bool
	s.server.db.QueryRow(`SELECT restrict_guild_scout FROM characters WHERE id=$1`, s.charID).Scan(&currentStatus)
	if currentStatus {
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
	} else {
		doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgMhfSetRejectGuildScout(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetRejectGuildScout)
	s.server.db.Exec(`UPDATE characters SET restrict_guild_scout=$1 WHERE id=$2`, pkt.Reject, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, nil)
}
