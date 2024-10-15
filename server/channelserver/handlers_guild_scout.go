package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/gametime"
	"erupe-ce/utils/stringsupport"
	"fmt"
	"io"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func HandleMsgMhfPostGuildScout(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostGuildScout)

	actorCharGuildData, err := GetCharacterGuildData(s, s.CharID)

	if err != nil {
		s.DoAckBufFail(pkt.AckHandle, make([]byte, 4))
		panic(err)
	}

	if actorCharGuildData == nil || !actorCharGuildData.CanRecruit() {
		s.DoAckBufFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	guildInfo, err := GetGuildInfoByID(s, actorCharGuildData.GuildID)

	if err != nil {
		s.DoAckBufFail(pkt.AckHandle, make([]byte, 4))
		panic(err)
	}

	hasApplication, err := guildInfo.HasApplicationForCharID(s, pkt.CharID)

	if err != nil {
		s.DoAckBufFail(pkt.AckHandle, make([]byte, 4))
		panic(err)
	}

	if hasApplication {
		s.DoAckBufSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x04})
		return
	}

	transaction, err := db.Begin()

	if err != nil {
		panic(err)
	}

	err = guildInfo.CreateApplication(s, pkt.CharID, GuildApplicationTypeInvited, transaction)

	if err != nil {
		rollbackTransaction(s, transaction)
		s.DoAckBufFail(pkt.AckHandle, nil)
		panic(err)
	}

	mail := &Mail{
		SenderID:    s.CharID,
		RecipientID: pkt.CharID,
		Subject:     s.Server.i18n.guild.invite.title,
		Body: fmt.Sprintf(
			s.Server.i18n.guild.invite.body,
			guildInfo.Name,
		),
		IsGuildInvite: true,
	}

	err = mail.Send(s, transaction)

	if err != nil {
		rollbackTransaction(s, transaction)
		s.DoAckBufFail(pkt.AckHandle, nil)
		return
	}

	err = transaction.Commit()

	if err != nil {
		s.DoAckBufFail(pkt.AckHandle, nil)
		panic(err)
	}

	s.DoAckBufSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func HandleMsgMhfCancelGuildScout(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCancelGuildScout)

	guildCharData, err := GetCharacterGuildData(s, s.CharID)

	if err != nil {
		panic(err)
	}

	if guildCharData == nil || !guildCharData.CanRecruit() {
		s.DoAckBufFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	guild, err := GetGuildInfoByID(s, guildCharData.GuildID)

	if err != nil {
		s.DoAckBufFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	err = guild.CancelInvitation(s, pkt.InvitationID)

	if err != nil {
		s.DoAckBufFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfAnswerGuildScout(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAnswerGuildScout)
	bf := byteframe.NewByteFrame()
	guild, err := GetGuildInfoByCharacterId(s, pkt.LeaderID)

	if err != nil {
		panic(err)
	}

	app, err := guild.GetApplicationForCharID(s, s.CharID, GuildApplicationTypeInvited)

	if app == nil || err != nil {
		s.Logger.Warn(
			"Guild invite missing, deleted?",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", s.CharID),
		)
		bf.WriteUint32(7)
		bf.WriteUint32(guild.ID)
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
		return
	}

	var mail []Mail
	if pkt.Answer {
		err = guild.AcceptApplication(s, s.CharID)
		mail = append(mail, Mail{
			RecipientID:     s.CharID,
			Subject:         s.Server.i18n.guild.invite.success.title,
			Body:            fmt.Sprintf(s.Server.i18n.guild.invite.success.body, guild.Name),
			IsSystemMessage: true,
		})
		mail = append(mail, Mail{
			SenderID:        s.CharID,
			RecipientID:     pkt.LeaderID,
			Subject:         s.Server.i18n.guild.invite.accepted.title,
			Body:            fmt.Sprintf(s.Server.i18n.guild.invite.accepted.body, guild.Name),
			IsSystemMessage: true,
		})
	} else {
		err = guild.RejectApplication(s, s.CharID)
		mail = append(mail, Mail{
			RecipientID:     s.CharID,
			Subject:         s.Server.i18n.guild.invite.rejected.title,
			Body:            fmt.Sprintf(s.Server.i18n.guild.invite.rejected.body, guild.Name),
			IsSystemMessage: true,
		})
		mail = append(mail, Mail{
			SenderID:        s.CharID,
			RecipientID:     pkt.LeaderID,
			Subject:         s.Server.i18n.guild.invite.declined.title,
			Body:            fmt.Sprintf(s.Server.i18n.guild.invite.declined.body, guild.Name),
			IsSystemMessage: true,
		})
	}
	if err != nil {
		bf.WriteUint32(7)
		bf.WriteUint32(guild.ID)
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
	} else {
		bf.WriteUint32(0)
		bf.WriteUint32(guild.ID)
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
		for _, m := range mail {
			m.Send(s, nil)
		}
	}
}

func HandleMsgMhfGetGuildScoutList(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildScoutList)

	guildInfo, err := GetGuildInfoByCharacterId(s, s.CharID)

	if guildInfo == nil && s.prevGuildID == 0 {
		s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		return
	} else {
		guildInfo, err = GetGuildInfoByID(s, s.prevGuildID)
		if guildInfo == nil || err != nil {
			s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	rows, err := db.Queryx(`
		SELECT c.id, c.name, c.hr, c.gr, ga.actor_id
			FROM guild_applications ga 
			JOIN characters c ON c.id = ga.character_id
		WHERE ga.guild_id = $1 AND ga.application_type = 'invited'
	`, guildInfo.ID)

	if err != nil {
		s.Logger.Error("failed to retrieve scouted characters", zap.Error(err))
		s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		return
	}

	defer rows.Close()

	bf := byteframe.NewByteFrame()

	bf.SetBE()

	// Result count, we will overwrite this later
	bf.WriteUint32(0x00)

	count := uint32(0)

	for rows.Next() {
		var charName string
		var charID, actorID uint32
		var HR, GR uint16

		err = rows.Scan(&charID, &charName, &HR, &GR, &actorID)

		if err != nil {
			s.DoAckSimpleFail(pkt.AckHandle, nil)
			continue
		}

		// This seems to be used as a unique ID for the invitation sent
		// we can just use the charID and then filter on guild_id+charID when performing operations
		// this might be a problem later with mails sent referencing IDs but we'll see.
		bf.WriteUint32(charID)
		bf.WriteUint32(actorID)
		bf.WriteUint32(charID)
		bf.WriteUint32(uint32(gametime.TimeAdjusted().Unix()))
		bf.WriteUint16(HR) // HR?
		bf.WriteUint16(GR) // GR?
		bf.WriteBytes(stringsupport.PaddedString(charName, 32, true))
		count++
	}

	_, err = bf.Seek(0, io.SeekStart)

	if err != nil {
		panic(err)
	}

	bf.WriteUint32(count)

	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfGetRejectGuildScout(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRejectGuildScout)

	row := db.QueryRow("SELECT restrict_guild_scout FROM characters WHERE id=$1", s.CharID)

	var currentStatus bool

	err := row.Scan(&currentStatus)

	if err != nil {
		s.Logger.Error(
			"failed to retrieve character guild scout status",
			zap.Error(err),
			zap.Uint32("charID", s.CharID),
		)
		s.DoAckSimpleFail(pkt.AckHandle, nil)
		return
	}

	response := uint8(0x00)

	if currentStatus {
		response = 0x01
	}

	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, response})
}

func HandleMsgMhfSetRejectGuildScout(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetRejectGuildScout)

	_, err := db.Exec("UPDATE characters SET restrict_guild_scout=$1 WHERE id=$2", pkt.Reject, s.CharID)

	if err != nil {
		s.Logger.Error(
			"failed to update character guild scout status",
			zap.Error(err),
			zap.Uint32("charID", s.CharID),
		)
		s.DoAckSimpleFail(pkt.AckHandle, nil)
		return
	}

	s.DoAckSimpleSucceed(pkt.AckHandle, nil)
}
