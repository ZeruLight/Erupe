package channelserver

import (
	"erupe-ce/config"
	"erupe-ce/internal/constant"
	"erupe-ce/internal/model"
	"erupe-ce/internal/service"

	"erupe-ce/utils/database"
	"erupe-ce/utils/gametime"
	"erupe-ce/utils/mhfitem"

	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	ps "erupe-ce/utils/pascalstring"
	"erupe-ce/utils/stringsupport"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func HandleMsgMhfCreateGuild(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateGuild)

	guildId, err := service.CreateGuild(pkt.Name, s.CharID)

	if err != nil {
		bf := byteframe.NewByteFrame()

		// No reasoning behind these values other than they cause a 'failed to create'
		// style message, it's better than nothing for now.
		bf.WriteUint32(0x01010101)

		s.DoAckSimpleFail(pkt.AckHandle, bf.Data())
		return
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint32(uint32(guildId))

	s.DoAckSimpleSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfOperateGuild(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuild)

	guild, err := service.GetGuildInfoByID(pkt.GuildID)
	characterGuildInfo, err := service.GetCharacterGuildData(s.CharID)
	if err != nil {
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	bf := byteframe.NewByteFrame()

	switch pkt.Action {
	case mhfpacket.OperateGuildDisband:
		response := 1
		if guild.LeaderCharID != s.CharID {
			s.Logger.Warn(fmt.Sprintf("character '%d' is attempting to manage guild '%d' without permission", s.CharID, guild.ID))
			response = 0
		} else {
			err = guild.Disband(s.CharID)
			if err != nil {
				response = 0
			}
		}
		bf.WriteUint32(uint32(response))
	case mhfpacket.OperateGuildResign:
		guildMembers, err := service.GetGuildMembers(guild.ID, false)
		if err == nil {
			sort.Slice(guildMembers[:], func(i, j int) bool {
				return guildMembers[i].OrderIndex < guildMembers[j].OrderIndex
			})
			for i := 1; i < len(guildMembers); i++ {
				if !guildMembers[i].AvoidLeadership {
					guild.LeaderCharID = guildMembers[i].CharID
					guildMembers[0].OrderIndex = guildMembers[i].OrderIndex
					guildMembers[i].OrderIndex = 1
					guildMembers[0].Save()
					guildMembers[i].Save()
					bf.WriteUint32(guildMembers[i].CharID)
					break
				}
			}
			guild.Save()
		}
	case mhfpacket.OperateGuildApply:
		err = guild.CreateApplication(s.CharID, constant.GuildApplicationTypeApplied, nil, s.CharID)
		if err == nil {
			bf.WriteUint32(guild.LeaderCharID)
		} else {
			bf.WriteUint32(0)
		}
	case mhfpacket.OperateGuildLeave:
		if characterGuildInfo.IsApplicant {
			err = guild.RejectApplication(s.CharID)
		} else {
			err = guild.RemoveCharacter(s.CharID)
		}
		response := 1
		if err != nil {
			response = 0
		} else {
			mail := service.Mail{
				RecipientID:     s.CharID,
				Subject:         "Withdrawal",
				Body:            fmt.Sprintf("You have withdrawn from 「%s」.", guild.Name),
				IsSystemMessage: true,
			}
			mail.Send(nil)
		}
		bf.WriteUint32(uint32(response))
	case mhfpacket.OperateGuildDonateRank:
		bf.WriteBytes(handleDonateRP(s, uint16(pkt.Data1.ReadUint32()), guild, 0))
	case mhfpacket.OperateGuildSetApplicationDeny:
		db.Exec("UPDATE guilds SET recruiting=false WHERE id=$1", guild.ID)
	case mhfpacket.OperateGuildSetApplicationAllow:
		db.Exec("UPDATE guilds SET recruiting=true WHERE id=$1", guild.ID)
	case mhfpacket.OperateGuildSetAvoidLeadershipTrue:
		handleAvoidLeadershipUpdate(s, pkt, true)
	case mhfpacket.OperateGuildSetAvoidLeadershipFalse:
		handleAvoidLeadershipUpdate(s, pkt, false)
	case mhfpacket.OperateGuildUpdateComment:
		if !characterGuildInfo.IsLeader && !characterGuildInfo.IsSubLeader() {
			s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
			return
		}
		guild.Comment = stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())
		guild.Save()
	case mhfpacket.OperateGuildUpdateMotto:
		if !characterGuildInfo.IsLeader && !characterGuildInfo.IsSubLeader() {
			s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
			return
		}
		_ = pkt.Data1.ReadUint16()
		guild.SubMotto = pkt.Data1.ReadUint8()
		guild.MainMotto = pkt.Data1.ReadUint8()
		guild.Save()
	case mhfpacket.OperateGuildRenamePugi1:
		handleRenamePugi(s, pkt.Data2, guild, 1)
	case mhfpacket.OperateGuildRenamePugi2:
		handleRenamePugi(s, pkt.Data2, guild, 2)
	case mhfpacket.OperateGuildRenamePugi3:
		handleRenamePugi(s, pkt.Data2, guild, 3)
	case mhfpacket.OperateGuildChangePugi1:
		handleChangePugi(s, uint8(pkt.Data1.ReadUint32()), guild, 1)
	case mhfpacket.OperateGuildChangePugi2:
		handleChangePugi(s, uint8(pkt.Data1.ReadUint32()), guild, 2)
	case mhfpacket.OperateGuildChangePugi3:
		handleChangePugi(s, uint8(pkt.Data1.ReadUint32()), guild, 3)
	case mhfpacket.OperateGuildUnlockOutfit:
		// TODO: This doesn't implement blocking, if someone unlocked the same outfit at the same time
		db.Exec(`UPDATE guilds SET pugi_outfits=pugi_outfits+$1 WHERE id=$2`, int(math.Pow(float64(pkt.Data1.ReadUint32()), 2)), guild.ID)
	case mhfpacket.OperateGuildDonateRoom:
		quantity := uint16(pkt.Data1.ReadUint32())
		bf.WriteBytes(handleDonateRP(s, quantity, guild, 2))
	case mhfpacket.OperateGuildDonateEvent:
		quantity := uint16(pkt.Data1.ReadUint32())
		bf.WriteBytes(handleDonateRP(s, quantity, guild, 1))
		// TODO: Move this value onto rp_yesterday and reset to 0... daily?
		db.Exec(`UPDATE guild_characters SET rp_today=rp_today+$1 WHERE character_id=$2`, quantity, s.CharID)
	case mhfpacket.OperateGuildEventExchange:
		rp := uint16(pkt.Data1.ReadUint32())
		var balance uint32
		db.QueryRow(`UPDATE guilds SET event_rp=event_rp-$1 WHERE id=$2 RETURNING event_rp`, rp, guild.ID).Scan(&balance)
		bf.WriteUint32(balance)
	default:
		panic(fmt.Sprintf("unhandled operate guild action '%d'", pkt.Action))
	}

	if len(bf.Data()) > 0 {
		s.DoAckSimpleSucceed(pkt.AckHandle, bf.Data())
	} else {
		s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
	}
}

func handleRenamePugi(s *Session, bf *byteframe.ByteFrame, guild *service.Guild, num int) {
	name := stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	switch num {
	case 1:
		guild.PugiName1 = name
	case 2:
		guild.PugiName2 = name
	default:
		guild.PugiName3 = name
	}
	guild.Save()
}

func handleChangePugi(s *Session, outfit uint8, guild *service.Guild, num int) {
	switch num {
	case 1:
		guild.PugiOutfit1 = outfit
	case 2:
		guild.PugiOutfit2 = outfit
	case 3:
		guild.PugiOutfit3 = outfit
	}
	guild.Save()
}

func handleDonateRP(s *Session, amount uint16, guild *service.Guild, _type int) []byte {
	db, err := database.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	saveData, err := service.GetCharacterSaveData(s.CharID)
	if err != nil {
		return bf.Data()
	}
	var resetRoom bool
	if _type == 2 {
		var currentRP uint16
		db.QueryRow(`SELECT room_rp FROM guilds WHERE id = $1`, guild.ID).Scan(&currentRP)
		if currentRP+amount >= 30 {
			amount = 30 - currentRP
			resetRoom = true
		}
	}
	saveData.RP -= amount
	saveData.Save(s)

	switch _type {
	case 0:
		db.Exec(`UPDATE guilds SET rank_rp = rank_rp + $1 WHERE id = $2`, amount, guild.ID)
	case 1:
		db.Exec(`UPDATE guilds SET event_rp = event_rp + $1 WHERE id = $2`, amount, guild.ID)
	case 2:
		if resetRoom {
			db.Exec(`UPDATE guilds SET room_rp = 0 WHERE id = $1`, guild.ID)
			db.Exec(`UPDATE guilds SET room_expiry = $1 WHERE id = $2`, gametime.TimeAdjusted().Add(time.Hour*24*7), guild.ID)
		} else {
			db.Exec(`UPDATE guilds SET room_rp = room_rp + $1 WHERE id = $2`, amount, guild.ID)
		}
	}
	bf.Seek(0, 0)
	bf.WriteUint32(uint32(saveData.RP))
	return bf.Data()
}

func handleAvoidLeadershipUpdate(s *Session, pkt *mhfpacket.MsgMhfOperateGuild, avoidLeadership bool) {
	characterGuildData, err := service.GetCharacterGuildData(s.CharID)

	if err != nil {
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	characterGuildData.AvoidLeadership = avoidLeadership

	err = characterGuildData.Save()

	if err != nil {
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfOperateGuildMember(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuildMember)

	guild, err := service.GetGuildInfoByCharacterId(pkt.CharID)

	if err != nil || guild == nil {
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	actorCharacter, err := service.GetCharacterGuildData(s.CharID)

	if err != nil || (!actorCharacter.IsSubLeader() && guild.LeaderCharID != s.CharID) {
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	var mail service.Mail
	switch pkt.Action {
	case mhfpacket.OPERATE_GUILD_MEMBER_ACTION_ACCEPT:
		err = guild.AcceptApplication(pkt.CharID)
		mail = service.Mail{
			RecipientID:     pkt.CharID,
			Subject:         "Accepted!",
			Body:            fmt.Sprintf("Your application to join 「%s」 was accepted.", guild.Name),
			IsSystemMessage: true,
		}
	case mhfpacket.OPERATE_GUILD_MEMBER_ACTION_REJECT:
		err = guild.RejectApplication(pkt.CharID)
		mail = service.Mail{
			RecipientID:     pkt.CharID,
			Subject:         "Rejected",
			Body:            fmt.Sprintf("Your application to join 「%s」 was rejected.", guild.Name),
			IsSystemMessage: true,
		}
	case mhfpacket.OPERATE_GUILD_MEMBER_ACTION_KICK:
		err = guild.RemoveCharacter(pkt.CharID)
		mail = service.Mail{
			RecipientID:     pkt.CharID,
			Subject:         "Kicked",
			Body:            fmt.Sprintf("You were kicked from 「%s」.", guild.Name),
			IsSystemMessage: true,
		}
	default:
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		s.Logger.Warn(fmt.Sprintf("unhandled operateGuildMember action '%d'", pkt.Action))
	}

	if err != nil {
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
	} else {
		mail.Send(nil)
		for _, channel := range s.Server.Channels {
			for _, session := range channel.sessions {
				if session.CharID == pkt.CharID {
					service.SendMailNotification(&mail, session)
				}
			}
		}
		s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
	}
}

func HandleMsgMhfInfoGuild(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoGuild)

	var guild *service.Guild
	var err error

	if pkt.GuildID > 0 {
		guild, err = service.GetGuildInfoByID(pkt.GuildID)
	} else {
		guild, err = service.GetGuildInfoByCharacterId(s.CharID)
	}

	if err == nil && guild != nil {
		s.prevGuildID = guild.ID

		guildName := stringsupport.UTF8ToSJIS(guild.Name)
		guildComment := stringsupport.UTF8ToSJIS(guild.Comment)
		guildLeaderName := stringsupport.UTF8ToSJIS(guild.LeaderName)

		characterGuildData, err := service.GetCharacterGuildData(s.CharID)
		characterJoinedAt := uint32(0xFFFFFFFF)

		if characterGuildData != nil && characterGuildData.JoinedAt != nil {
			characterJoinedAt = uint32(characterGuildData.JoinedAt.Unix())
		}

		if err != nil {
			resp := byteframe.NewByteFrame()
			resp.WriteUint32(0) // Count
			resp.WriteUint8(0)  // Unk, read if count == 0.

			s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
			return
		}

		bf := byteframe.NewByteFrame()

		bf.WriteUint32(guild.ID)
		bf.WriteUint32(guild.LeaderCharID)
		bf.WriteUint16(guild.Rank())
		bf.WriteUint16(guild.MemberCount)

		bf.WriteUint8(guild.MainMotto)
		bf.WriteUint8(guild.SubMotto)

		// Unk appears to be static
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)

		bf.WriteBool(!guild.Recruiting)

		if characterGuildData == nil || characterGuildData.IsApplicant {
			bf.WriteUint16(0x00)
		} else if guild.LeaderCharID == s.CharID {
			bf.WriteUint16(0x01)
		} else {
			bf.WriteUint16(0x02)
		}

		bf.WriteUint32(uint32(guild.CreatedAt.Unix()))
		bf.WriteUint32(characterJoinedAt)
		bf.WriteUint8(uint8(len(guildName)))
		bf.WriteUint8(uint8(len(guildComment)))
		bf.WriteUint8(uint8(5)) // Length of unknown string below
		bf.WriteUint8(uint8(len(guildLeaderName)))
		bf.WriteBytes(guildName)
		bf.WriteBytes(guildComment)
		bf.WriteInt8(int8(constant.FestivalColorCodes[guild.FestivalColor]))
		bf.WriteUint32(guild.RankRP)
		bf.WriteBytes(guildLeaderName)
		bf.WriteUint32(0)   // Unk
		bf.WriteBool(false) // isReturnGuild
		bf.WriteBool(false) // earnedSpecialHall
		bf.WriteUint8(2)
		bf.WriteUint8(2)
		bf.WriteUint32(guild.EventRP) // Skipped if last byte is <2?
		ps.Uint8(bf, guild.PugiName1, true)
		ps.Uint8(bf, guild.PugiName2, true)
		ps.Uint8(bf, guild.PugiName3, true)
		bf.WriteUint8(guild.PugiOutfit1)
		bf.WriteUint8(guild.PugiOutfit2)
		bf.WriteUint8(guild.PugiOutfit3)
		if config.GetConfig().ClientID >= config.Z1 {
			bf.WriteUint8(guild.PugiOutfit1)
			bf.WriteUint8(guild.PugiOutfit2)
			bf.WriteUint8(guild.PugiOutfit3)
		}
		bf.WriteUint32(guild.PugiOutfits)

		limit := config.GetConfig().GameplayOptions.ClanMemberLimits[0][1]
		for _, j := range config.GetConfig().GameplayOptions.ClanMemberLimits {
			if guild.Rank() >= uint16(j[0]) {
				limit = j[1]
			}
		}
		if limit > 100 {
			limit = 100
		}
		bf.WriteUint8(limit)

		bf.WriteUint32(55000)
		bf.WriteUint32(uint32(guild.RoomExpiry.Unix()))
		bf.WriteUint16(guild.RoomRP)
		bf.WriteUint16(0) // Ignored

		if guild.AllianceID > 0 {
			alliance, err := service.GetAllianceData(guild.AllianceID)
			if err != nil {
				bf.WriteUint32(0) // Error, no alliance
			} else {
				bf.WriteUint32(alliance.ID)
				bf.WriteUint32(uint32(alliance.CreatedAt.Unix()))
				bf.WriteUint16(alliance.TotalMembers)
				bf.WriteUint8(0) // Ignored
				bf.WriteUint8(0)
				ps.Uint16(bf, alliance.Name, true)
				if alliance.SubGuild1ID > 0 {
					if alliance.SubGuild2ID > 0 {
						bf.WriteUint8(3)
					} else {
						bf.WriteUint8(2)
					}
				} else {
					bf.WriteUint8(1)
				}
				bf.WriteUint32(alliance.ParentGuildID)
				bf.WriteUint32(0) // Unk1
				if alliance.ParentGuildID == guild.ID {
					bf.WriteUint16(1)
				} else {
					bf.WriteUint16(0)
				}
				bf.WriteUint16(alliance.ParentGuild.Rank())
				bf.WriteUint16(alliance.ParentGuild.MemberCount)
				ps.Uint16(bf, alliance.ParentGuild.Name, true)
				ps.Uint16(bf, alliance.ParentGuild.LeaderName, true)
				if alliance.SubGuild1ID > 0 {
					bf.WriteUint32(alliance.SubGuild1ID)
					bf.WriteUint32(0) // Unk1
					if alliance.SubGuild1ID == guild.ID {
						bf.WriteUint16(1)
					} else {
						bf.WriteUint16(0)
					}
					bf.WriteUint16(alliance.SubGuild1.Rank())
					bf.WriteUint16(alliance.SubGuild1.MemberCount)
					ps.Uint16(bf, alliance.SubGuild1.Name, true)
					ps.Uint16(bf, alliance.SubGuild1.LeaderName, true)
				}
				if alliance.SubGuild2ID > 0 {
					bf.WriteUint32(alliance.SubGuild2ID)
					bf.WriteUint32(0) // Unk1
					if alliance.SubGuild2ID == guild.ID {
						bf.WriteUint16(1)
					} else {
						bf.WriteUint16(0)
					}
					bf.WriteUint16(alliance.SubGuild2.Rank())
					bf.WriteUint16(alliance.SubGuild2.MemberCount)
					ps.Uint16(bf, alliance.SubGuild2.Name, true)
					ps.Uint16(bf, alliance.SubGuild2.LeaderName, true)
				}
			}
		} else {
			bf.WriteUint32(0) // No alliance
		}

		applicants, err := service.GetGuildMembers(guild.ID, true)
		if err != nil || (characterGuildData != nil && !characterGuildData.CanRecruit()) {
			bf.WriteUint16(0)
		} else {
			bf.WriteUint16(uint16(len(applicants)))
			for _, applicant := range applicants {
				bf.WriteUint32(applicant.CharID)
				bf.WriteUint32(0)
				bf.WriteUint16(applicant.HR)
				bf.WriteUint16(applicant.GR)
				ps.Uint8(bf, applicant.Name, true)
			}
		}

		unkGuildInfo := []model.UnkGuildInfo{}
		bf.WriteUint8(uint8(len(unkGuildInfo)))
		for _, info := range unkGuildInfo {
			bf.WriteUint8(info.Unk0)
			bf.WriteUint8(info.Unk1)
			bf.WriteUint8(info.Unk2)
		}

		allianceInvites := []model.GuildAllianceInvite{}
		bf.WriteUint8(uint8(len(allianceInvites)))
		for _, invite := range allianceInvites {
			bf.WriteUint32(invite.GuildID)
			bf.WriteUint32(invite.LeaderID)
			bf.WriteUint16(invite.Unk0)
			bf.WriteUint16(invite.Unk1)
			bf.WriteUint16(invite.Members)
			ps.Uint16(bf, invite.GuildName, true)
			ps.Uint16(bf, invite.LeaderName, true)
		}

		if guild.Icon != nil {
			bf.WriteUint8(uint8(len(guild.Icon.Parts)))

			for _, p := range guild.Icon.Parts {
				bf.WriteUint16(p.Index)
				bf.WriteUint16(p.ID)
				bf.WriteUint8(p.Page)
				bf.WriteUint8(p.Size)
				bf.WriteUint8(p.Rotation)
				bf.WriteUint8(p.Red)
				bf.WriteUint8(p.Green)
				bf.WriteUint8(p.Blue)
				bf.WriteUint16(p.PosX)
				bf.WriteUint16(p.PosY)
			}
		} else {
			bf.WriteUint8(0)
		}
		bf.WriteUint8(0) // Unk

		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
	} else {
		s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 5))
	}
}

func HandleMsgMhfEnumerateGuild(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuild)

	var guilds []*service.Guild
	var alliances []*service.GuildAlliance
	var rows *sqlx.Rows
	var err error

	if pkt.Type <= 8 {
		var tempGuilds []*service.Guild
		rows, err = db.Queryx(service.GuildInfoSelectQuery)
		if err == nil {
			for rows.Next() {
				guild, err := service.BuildGuildObjectFromDbResult(rows, err)
				if err != nil {
					continue
				}
				tempGuilds = append(tempGuilds, guild)
			}
		}
		switch pkt.Type {
		case mhfpacket.ENUMERATE_GUILD_TYPE_GUILD_NAME:
			for _, guild := range tempGuilds {
				if strings.Contains(guild.Name, stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())) {
					guilds = append(guilds, guild)
				}
			}
		case mhfpacket.ENUMERATE_GUILD_TYPE_LEADER_NAME:
			for _, guild := range tempGuilds {
				if strings.Contains(guild.LeaderName, stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())) {
					guilds = append(guilds, guild)
				}
			}
		case mhfpacket.ENUMERATE_GUILD_TYPE_LEADER_ID:
			CID := pkt.Data1.ReadUint32()
			for _, guild := range tempGuilds {
				if guild.LeaderCharID == CID {
					guilds = append(guilds, guild)
				}
			}
		case mhfpacket.ENUMERATE_GUILD_TYPE_ORDER_MEMBERS:
			if pkt.Sorting {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].MemberCount > tempGuilds[j].MemberCount
				})
			} else {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].MemberCount < tempGuilds[j].MemberCount
				})
			}
			guilds = tempGuilds
		case mhfpacket.ENUMERATE_GUILD_TYPE_ORDER_REGISTRATION:
			if pkt.Sorting {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].CreatedAt.Unix() > tempGuilds[j].CreatedAt.Unix()
				})
			} else {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].CreatedAt.Unix() < tempGuilds[j].CreatedAt.Unix()
				})
			}
			guilds = tempGuilds
		case mhfpacket.ENUMERATE_GUILD_TYPE_ORDER_RANK:
			if pkt.Sorting {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].RankRP > tempGuilds[j].RankRP
				})
			} else {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].RankRP < tempGuilds[j].RankRP
				})
			}
			guilds = tempGuilds
		case mhfpacket.ENUMERATE_GUILD_TYPE_MOTTO:
			mainMotto := uint8(pkt.Data1.ReadUint16())
			subMotto := uint8(pkt.Data1.ReadUint16())
			for _, guild := range tempGuilds {
				if guild.MainMotto == mainMotto && guild.SubMotto == subMotto {
					guilds = append(guilds, guild)
				}
			}
		case mhfpacket.ENUMERATE_GUILD_TYPE_RECRUITING:
			recruitingMotto := uint8(pkt.Data1.ReadUint16())
			for _, guild := range tempGuilds {
				if guild.MainMotto == recruitingMotto {
					guilds = append(guilds, guild)
				}
			}
		}
	}

	if pkt.Type > 8 {
		var tempAlliances []*service.GuildAlliance
		rows, err = db.Queryx(service.AllianceInfoSelectQuery)
		if err == nil {
			for rows.Next() {
				alliance, _ := service.BuildAllianceObjectFromDbResult(rows, err)
				tempAlliances = append(tempAlliances, alliance)
			}
		}
		switch pkt.Type {
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_ALLIANCE_NAME:
			for _, alliance := range tempAlliances {
				if strings.Contains(alliance.Name, stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())) {
					alliances = append(alliances, alliance)
				}
			}
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_LEADER_NAME:
			for _, alliance := range tempAlliances {
				if strings.Contains(alliance.ParentGuild.LeaderName, stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())) {
					alliances = append(alliances, alliance)
				}
			}
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_LEADER_ID:
			CID := pkt.Data1.ReadUint32()
			for _, alliance := range tempAlliances {
				if alliance.ParentGuild.LeaderCharID == CID {
					alliances = append(alliances, alliance)
				}
			}
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_ORDER_MEMBERS:
			if pkt.Sorting {
				sort.Slice(tempAlliances, func(i, j int) bool {
					return tempAlliances[i].TotalMembers > tempAlliances[j].TotalMembers
				})
			} else {
				sort.Slice(tempAlliances, func(i, j int) bool {
					return tempAlliances[i].TotalMembers < tempAlliances[j].TotalMembers
				})
			}
			alliances = tempAlliances
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_ORDER_REGISTRATION:
			if pkt.Sorting {
				sort.Slice(tempAlliances, func(i, j int) bool {
					return tempAlliances[i].CreatedAt.Unix() > tempAlliances[j].CreatedAt.Unix()
				})
			} else {
				sort.Slice(tempAlliances, func(i, j int) bool {
					return tempAlliances[i].CreatedAt.Unix() < tempAlliances[j].CreatedAt.Unix()
				})
			}
			alliances = tempAlliances
		}
	}

	if err != nil || (guilds == nil && alliances == nil) {
		stubEnumerateNoResults(s, pkt.AckHandle)
		return
	}

	bf := byteframe.NewByteFrame()

	if pkt.Type > 8 {
		hasNextPage := false
		if len(alliances) > 10 {
			hasNextPage = true
			alliances = alliances[:10]
		}
		bf.WriteUint16(uint16(len(alliances)))
		bf.WriteBool(hasNextPage)
		for _, alliance := range alliances {
			bf.WriteUint32(alliance.ID)
			bf.WriteUint32(alliance.ParentGuild.LeaderCharID)
			bf.WriteUint16(alliance.TotalMembers)
			bf.WriteUint16(0x0000)
			if alliance.SubGuild1ID == 0 && alliance.SubGuild2ID == 0 {
				bf.WriteUint16(1)
			} else if alliance.SubGuild1ID > 0 && alliance.SubGuild2ID == 0 || alliance.SubGuild1ID == 0 && alliance.SubGuild2ID > 0 {
				bf.WriteUint16(2)
			} else {
				bf.WriteUint16(3)
			}
			bf.WriteUint32(uint32(alliance.CreatedAt.Unix()))
			ps.Uint8(bf, alliance.Name, true)
			ps.Uint8(bf, alliance.ParentGuild.LeaderName, true)
			bf.WriteUint8(0x01) // Unk
			bf.WriteBool(true)  // TODO: Enable GuildAlliance applications
		}
	} else {
		hasNextPage := false
		if len(guilds) > 10 {
			hasNextPage = true
			guilds = guilds[:10]
		}
		bf.WriteUint16(uint16(len(guilds)))
		bf.WriteBool(hasNextPage)
		for _, guild := range guilds {
			bf.WriteUint32(guild.ID)
			bf.WriteUint32(guild.LeaderCharID)
			bf.WriteUint16(guild.MemberCount)
			bf.WriteUint16(0x0000) // Unk
			bf.WriteUint16(guild.Rank())
			bf.WriteUint32(uint32(guild.CreatedAt.Unix()))
			ps.Uint8(bf, guild.Name, true)
			ps.Uint8(bf, guild.LeaderName, true)
			bf.WriteUint8(0x01) // Unk
			bf.WriteBool(!guild.Recruiting)
		}
	}

	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfArrangeGuildMember(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfArrangeGuildMember)

	guild, err := service.GetGuildInfoByID(pkt.GuildID)

	if err != nil {
		s.Logger.Error(
			"failed to respond to ArrangeGuildMember message",
			zap.Uint32("charID", s.CharID),
		)
		return
	}

	if guild.LeaderCharID != s.CharID {
		s.Logger.Error("non leader attempting to rearrange guild members!",
			zap.Uint32("charID", s.CharID),
			zap.Uint32("guildID", guild.ID),
		)
		return
	}

	err = guild.ArrangeCharacters(pkt.CharIDs)

	if err != nil {
		s.Logger.Error(
			"failed to respond to ArrangeGuildMember message",
			zap.Uint32("charID", s.CharID),
			zap.Uint32("guildID", guild.ID),
		)
		return
	}

	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfEnumerateGuildMember(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildMember)

	var guild *service.Guild
	var err error

	if pkt.GuildID > 0 {
		guild, err = service.GetGuildInfoByID(pkt.GuildID)
	} else {
		guild, err = service.GetGuildInfoByCharacterId(s.CharID)
	}

	if guild != nil {
		isApplicant, _ := guild.HasApplicationForCharID(s.CharID)
		if isApplicant {
			s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 2))
			return
		}
	}

	if guild == nil && s.prevGuildID > 0 {
		guild, err = service.GetGuildInfoByID(s.prevGuildID)
	}

	if err != nil {
		s.Logger.Warn("failed to retrieve guild sending no result message")
		s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 2))
		return
	} else if guild == nil {
		s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 2))
		return
	}

	guildMembers, err := service.GetGuildMembers(guild.ID, false)

	if err != nil {
		s.Logger.Error("failed to retrieve guild")
		return
	}

	alliance, err := service.GetAllianceData(guild.AllianceID)
	if err != nil {
		s.Logger.Error("Failed to get alliance data")
		return
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint16(uint16(len(guildMembers)))

	sort.Slice(guildMembers[:], func(i, j int) bool {
		return guildMembers[i].OrderIndex < guildMembers[j].OrderIndex
	})

	for _, member := range guildMembers {
		bf.WriteUint32(member.CharID)
		bf.WriteUint16(member.HR)
		if config.GetConfig().ClientID >= config.G10 {
			bf.WriteUint16(member.GR)
		}
		if config.GetConfig().ClientID < config.ZZ {
			// Magnet Spike crash workaround
			bf.WriteUint16(0)
		} else {
			bf.WriteUint16(member.WeaponID)
		}
		if member.WeaponType == 1 || member.WeaponType == 5 || member.WeaponType == 10 { // If weapon is ranged
			bf.WriteUint8(7)
		} else {
			bf.WriteUint8(6)
		}
		bf.WriteUint16(member.OrderIndex)
		bf.WriteBool(member.AvoidLeadership)
		ps.Uint8(bf, member.Name, true)
	}

	for _, member := range guildMembers {
		bf.WriteUint32(member.LastLogin)
	}

	if guild.AllianceID > 0 {
		bf.WriteUint16(alliance.TotalMembers - uint16(len(guildMembers)))
		if guild.ID != alliance.ParentGuildID {
			mems, err := service.GetGuildMembers(alliance.ParentGuildID, false)
			if err != nil {
				panic(err)
			}
			for _, m := range mems {
				bf.WriteUint32(m.CharID)
			}
		}
		if guild.ID != alliance.SubGuild1ID {
			mems, err := service.GetGuildMembers(alliance.SubGuild1ID, false)
			if err != nil {
				panic(err)
			}
			for _, m := range mems {
				bf.WriteUint32(m.CharID)
			}
		}
		if guild.ID != alliance.SubGuild2ID {
			mems, err := service.GetGuildMembers(alliance.SubGuild2ID, false)
			if err != nil {
				panic(err)
			}
			for _, m := range mems {
				bf.WriteUint32(m.CharID)
			}
		}
	} else {
		bf.WriteUint16(0)
	}

	for _, member := range guildMembers {
		bf.WriteUint16(member.RPToday)
		bf.WriteUint16(member.RPYesterday)
	}

	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfGetGuildManageRight(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildManageRight)

	guild, err := service.GetGuildInfoByCharacterId(s.CharID)
	if guild == nil && s.prevGuildID != 0 {
		guild, err = service.GetGuildInfoByID(s.prevGuildID)
		s.prevGuildID = 0
		if guild == nil || err != nil {
			s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(guild.MemberCount))
	members, _ := service.GetGuildMembers(guild.ID, false)
	for _, member := range members {
		bf.WriteUint32(member.CharID)
		bf.WriteBool(member.Recruiter)
		bf.WriteBytes(make([]byte, 3))
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfGetUdGuildMapInfo(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdGuildMapInfo)
	s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfGetGuildTargetMemberNum(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildTargetMemberNum)

	var guild *service.Guild
	var err error

	if pkt.GuildID == 0x0 {
		guild, err = service.GetGuildInfoByCharacterId(s.CharID)
	} else {
		guild, err = service.GetGuildInfoByID(pkt.GuildID)
	}

	if err != nil || guild == nil {
		s.DoAckBufSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x02})
		return
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint16(0x0)
	bf.WriteUint16(guild.MemberCount - 1)

	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func guildGetItems(s *Session, guildID uint32) []mhfitem.MHFItemStack {
	db, err := database.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var data []byte
	var items []mhfitem.MHFItemStack

	db.QueryRow(`SELECT item_box FROM guilds WHERE id=$1`, guildID).Scan(&data)
	if len(data) > 0 {
		box := byteframe.NewByteFrameFromBytes(data)
		numStacks := box.ReadUint16()
		box.ReadUint16() // Unused
		for i := 0; i < int(numStacks); i++ {
			items = append(items, mhfitem.ReadWarehouseItem(box))
		}
	}
	return items
}

func HandleMsgMhfEnumerateGuildItem(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildItem)
	items := guildGetItems(s, pkt.GuildID)
	bf := byteframe.NewByteFrame()
	bf.WriteBytes(mhfitem.SerializeWarehouseItems(items))
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfUpdateGuildItem(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuildItem)
	newStacks := mhfitem.DiffItemStacks(guildGetItems(s, pkt.GuildID), pkt.UpdatedItems)

	db.Exec(`UPDATE guilds SET item_box=$1 WHERE id=$2`, mhfitem.SerializeWarehouseItems(newStacks), pkt.GuildID)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfUpdateGuildIcon(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuildIcon)

	guild, err := service.GetGuildInfoByID(pkt.GuildID)

	if err != nil {
		panic(err)
	}

	characterInfo, err := service.GetCharacterGuildData(s.CharID)

	if err != nil {
		panic(err)
	}

	if !characterInfo.IsSubLeader() && !characterInfo.IsLeader {
		s.Logger.Warn(
			"character without leadership attempting to update guild icon",
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", s.CharID),
		)
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	icon := &service.GuildIcon{}

	icon.Parts = make([]model.GuildIconPart, len(pkt.IconParts))

	for i, p := range pkt.IconParts {
		icon.Parts[i] = model.GuildIconPart{
			Index:    p.Index,
			ID:       p.ID,
			Page:     p.Page,
			Size:     p.Size,
			Rotation: p.Rotation,
			Red:      p.Red,
			Green:    p.Green,
			Blue:     p.Blue,
			PosX:     p.PosX,
			PosY:     p.PosY,
		}
	}

	guild.Icon = icon

	err = guild.Save()

	if err != nil {
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		return
	}

	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfReadGuildcard(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadGuildcard)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
}

func HandleMsgMhfGetGuildMissionList(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildMissionList)
	bf := byteframe.NewByteFrame()
	missions := []model.GuildMission{
		{431201, 574, 1, 4761, 35, 1, false, 2, 1},
		{431202, 755, 0, 95, 12, 2, false, 3, 2},
		{431203, 746, 0, 95, 6, 1, false, 1, 1},
		{431204, 581, 0, 83, 16, 2, false, 4, 2},
		{431205, 694, 1, 4763, 25, 1, false, 2, 1},
		{431206, 988, 0, 27, 16, 1, false, 6, 1},
		{431207, 730, 1, 4768, 25, 1, false, 4, 1},
		{431208, 680, 1, 3567, 50, 2, false, 2, 2},
		{431209, 1109, 0, 34, 60, 2, false, 6, 2},
		{431210, 128, 1, 8921, 70, 2, false, 3, 2},
		{431211, 406, 0, 59, 10, 1, false, 1, 1},
		{431212, 1170, 0, 70, 90, 3, false, 6, 3},
		{431213, 164, 0, 38, 24, 2, false, 6, 2},
		{431214, 378, 1, 3556, 150, 3, false, 1, 3},
		{431215, 446, 0, 94, 20, 2, false, 4, 2},
	}
	for _, mission := range missions {
		bf.WriteUint32(mission.ID)
		bf.WriteUint32(mission.Unk)
		bf.WriteUint16(mission.Type)
		bf.WriteUint16(mission.Goal)
		bf.WriteUint16(mission.Quantity)
		bf.WriteUint16(mission.SkipTickets)
		bf.WriteBool(mission.GR)
		bf.WriteUint16(mission.RewardType)
		bf.WriteUint16(mission.RewardLevel)
		bf.WriteUint32(uint32(gametime.TimeAdjusted().Unix()))
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfGetGuildMissionRecord(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildMissionRecord)

	// No guild mission records = 0x190 empty bytes
	s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 0x190))
}

func HandleMsgMhfAddGuildMissionCount(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddGuildMissionCount)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfSetGuildMissionTarget(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetGuildMissionTarget)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfCancelGuildMissionTarget(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCancelGuildMissionTarget)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfLoadGuildCooking(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadGuildCooking)

	guild, _ := service.GetGuildInfoByCharacterId(s.CharID)
	data, err := db.Queryx("SELECT id, meal_id, level, created_at FROM guild_meals WHERE guild_id = $1", guild.ID)
	if err != nil {
		s.Logger.Error("Failed to get guild meals from db", zap.Error(err))
		s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 2))
		return
	}
	var meals []model.GuildMeal
	var temp model.GuildMeal
	for data.Next() {
		err = data.StructScan(&temp)
		if err != nil {
			continue
		}
		if temp.CreatedAt.Add(60 * time.Minute).After(gametime.TimeAdjusted()) {
			meals = append(meals, temp)
		}
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(meals)))
	for _, meal := range meals {
		bf.WriteUint32(meal.ID)
		bf.WriteUint32(meal.MealID)
		bf.WriteUint32(meal.Level)
		bf.WriteUint32(uint32(meal.CreatedAt.Unix()))
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfRegistGuildCooking(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildCooking)

	guild, _ := service.GetGuildInfoByCharacterId(s.CharID)
	startTime := gametime.TimeAdjusted().Add(time.Duration(config.GetConfig().GameplayOptions.ClanMealDuration-3600) * time.Second)
	if pkt.OverwriteID != 0 {
		db.Exec("UPDATE guild_meals SET meal_id = $1, level = $2, created_at = $3 WHERE id = $4", pkt.MealID, pkt.Success, startTime, pkt.OverwriteID)
	} else {
		db.QueryRow("INSERT INTO guild_meals (guild_id, meal_id, level, created_at) VALUES ($1, $2, $3, $4) RETURNING id", guild.ID, pkt.MealID, pkt.Success, startTime).Scan(&pkt.OverwriteID)
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(1)
	bf.WriteUint32(pkt.OverwriteID)
	bf.WriteUint32(uint32(pkt.MealID))
	bf.WriteUint32(uint32(pkt.Success))
	bf.WriteUint32(uint32(startTime.Unix()))
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfGetGuildWeeklyBonusMaster(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildWeeklyBonusMaster)

	// Values taken from brand new guild capture
	s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 40))
}
func HandleMsgMhfGetGuildWeeklyBonusActiveCount(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildWeeklyBonusActiveCount)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(60) // Active count
	bf.WriteUint8(60) // Current active count
	bf.WriteUint8(0)  // New active count
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfGuildHuntdata(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGuildHuntdata)

	bf := byteframe.NewByteFrame()
	switch pkt.Operation {
	case 0: // Acquire
		db.Exec(`UPDATE guild_characters SET box_claimed=$1 WHERE character_id=$2`, gametime.TimeAdjusted(), s.CharID)
	case 1: // Enumerate
		bf.WriteUint8(0) // Entries
		rows, err := db.Query(`SELECT kl.id, kl.monster FROM kill_logs kl
			INNER JOIN guild_characters gc ON kl.character_id = gc.character_id
			WHERE gc.guild_id=$1
			AND kl.timestamp >= (SELECT box_claimed FROM guild_characters WHERE character_id=$2)
		`, pkt.GuildID, s.CharID)
		if err == nil {
			var count uint8
			var huntID, monID uint32
			for rows.Next() {
				err = rows.Scan(&huntID, &monID)
				if err != nil {
					continue
				}
				count++
				if count > 255 {
					count = 255
					rows.Close()
					break
				}
				bf.WriteUint32(huntID)
				bf.WriteUint32(monID)
			}
			bf.Seek(0, 0)
			bf.WriteUint8(count)
		}
	case 2: // Check
		guild, err := service.GetGuildInfoByCharacterId(s.CharID)
		if err == nil {
			var count uint8
			err = db.QueryRow(`SELECT COUNT(*) FROM kill_logs kl
				INNER JOIN guild_characters gc ON kl.character_id = gc.character_id
				WHERE gc.guild_id=$1
				AND kl.timestamp >= (SELECT box_claimed FROM guild_characters WHERE character_id=$2)
			`, guild.ID, s.CharID).Scan(&count)
			if err == nil && count > 0 {
				bf.WriteBool(true)
			} else {
				bf.WriteBool(false)
			}
		} else {
			bf.WriteBool(false)
		}
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func HandleMsgMhfEnumerateGuildMessageBoard(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildMessageBoard)

	guild, _ := service.GetGuildInfoByCharacterId(s.CharID)
	if pkt.BoardType == 1 {
		pkt.MaxPosts = 4
	}
	msgs, err := db.Queryx("SELECT id, stamp_id, title, body, author_id, created_at, liked_by FROM guild_posts WHERE guild_id = $1 AND post_type = $2 ORDER BY created_at DESC", guild.ID, int(pkt.BoardType))
	if err != nil {
		s.Logger.Error("Failed to get guild messages from db", zap.Error(err))
		s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 4))
		return
	}
	db.Exec("UPDATE characters SET guild_post_checked = now() WHERE id = $1", s.CharID)
	bf := byteframe.NewByteFrame()
	var postCount uint32
	for msgs.Next() {
		postData := &model.MessageBoardPost{}
		err = msgs.StructScan(&postData)
		if err != nil {
			continue
		}
		postCount++
		bf.WriteUint32(postData.ID)
		bf.WriteUint32(postData.AuthorID)
		bf.WriteUint32(0)
		bf.WriteUint32(uint32(postData.Timestamp.Unix()))
		bf.WriteUint32(uint32(stringsupport.CSVLength(postData.LikedBy)))
		bf.WriteBool(stringsupport.CSVContains(postData.LikedBy, int(s.CharID)))
		bf.WriteUint32(postData.StampID)
		ps.Uint32(bf, postData.Title, true)
		ps.Uint32(bf, postData.Body, true)
	}
	data := byteframe.NewByteFrame()
	data.WriteUint32(postCount)
	data.WriteBytes(bf.Data())
	s.DoAckBufSucceed(pkt.AckHandle, data.Data())
}

func HandleMsgMhfUpdateGuildMessageBoard(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuildMessageBoard)

	guild, err := service.GetGuildInfoByCharacterId(s.CharID)
	applicant := false
	if guild != nil {
		applicant, _ = guild.HasApplicationForCharID(s.CharID)
	}
	if err != nil || guild == nil || applicant {
		s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		return
	}
	switch pkt.MessageOp {
	case 0: // Create message
		db.Exec("INSERT INTO guild_posts (guild_id, author_id, stamp_id, post_type, title, body) VALUES ($1, $2, $3, $4, $5, $6)", guild.ID, s.CharID, pkt.StampID, pkt.PostType, pkt.Title, pkt.Body)
		// TODO: if there are too many messages, purge excess
	case 1: // Delete message
		db.Exec("DELETE FROM guild_posts WHERE id = $1", pkt.PostID)
	case 2: // Update message
		db.Exec("UPDATE guild_posts SET title = $1, body = $2 WHERE id = $3", pkt.Title, pkt.Body, pkt.PostID)
	case 3: // Update stamp
		db.Exec("UPDATE guild_posts SET stamp_id = $1 WHERE id = $2", pkt.StampID, pkt.PostID)
	case 4: // Like message
		var likedBy string
		err := db.QueryRow("SELECT liked_by FROM guild_posts WHERE id = $1", pkt.PostID).Scan(&likedBy)
		if err != nil {
			s.Logger.Error("Failed to get guild message like data from db", zap.Error(err))
		} else {
			if pkt.LikeState {
				likedBy = stringsupport.CSVAdd(likedBy, int(s.CharID))
				db.Exec("UPDATE guild_posts SET liked_by = $1 WHERE id = $2", likedBy, pkt.PostID)
			} else {
				likedBy = stringsupport.CSVRemove(likedBy, int(s.CharID))
				db.Exec("UPDATE guild_posts SET liked_by = $1 WHERE id = $2", likedBy, pkt.PostID)
			}
		}
	case 5: // Check for new messages
		var timeChecked time.Time
		var newPosts int
		err := db.QueryRow("SELECT guild_post_checked FROM characters WHERE id = $1", s.CharID).Scan(&timeChecked)
		if err == nil {
			db.QueryRow("SELECT COUNT(*) FROM guild_posts WHERE guild_id = $1 AND (EXTRACT(epoch FROM created_at)::int) > $2", guild.ID, timeChecked.Unix()).Scan(&newPosts)
			if newPosts > 0 {
				s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
				return
			}
		}
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfEntryRookieGuild(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryRookieGuild)
	s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfUpdateForceGuildRank(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func HandleMsgMhfAddGuildWeeklyBonusExceptionalUser(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddGuildWeeklyBonusExceptionalUser)
	// TODO: record pkt.NumUsers to DB
	// must use addition
	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func HandleMsgMhfGenerateUdGuildMap(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGenerateUdGuildMap)
	s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfUpdateGuild(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func HandleMsgMhfSetGuildManageRight(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetGuildManageRight)

	db.Exec("UPDATE guild_characters SET recruiter=$1 WHERE character_id=$2", pkt.Allowed, pkt.CharID)
	s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfCheckMonthlyItem(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckMonthlyItem)
	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
	// TODO: Implement month-by-month tracker, 0 = Not claimed, 1 = Claimed
	// Also handles HLC and EXC items, IDs = 064D, 076B
}

func HandleMsgMhfAcquireMonthlyItem(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireMonthlyItem)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfEnumerateInvGuild(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateInvGuild)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func HandleMsgMhfOperationInvGuild(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperationInvGuild)
	s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfUpdateGuildcard(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}
