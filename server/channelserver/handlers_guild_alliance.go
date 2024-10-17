package channelserver

import (
	"erupe-ce/internal/service"
	"erupe-ce/utils/byteframe"
	ps "erupe-ce/utils/pascalstring"
	"fmt"

	"erupe-ce/network/mhfpacket"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func HandleMsgMhfCreateJoint(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateJoint)

	_, err := db.Exec("INSERT INTO guild_alliances (name, parent_id) VALUES ($1, $2)", pkt.Name, pkt.GuildID)
	if err != nil {
		s.Logger.Error("Failed to create guild alliance in db", zap.Error(err))
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x01, 0x01, 0x01, 0x01})
}

func HandleMsgMhfOperateJoint(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateJoint)

	guild, err := service.GetGuildInfoByID(pkt.GuildID)
	if err != nil {
		s.Logger.Error("Failed to get guild info", zap.Error(err))
	}
	alliance, err := service.GetAllianceData(pkt.AllianceID)
	if err != nil {
		s.Logger.Error("Failed to get alliance info", zap.Error(err))
	}

	switch pkt.Action {
	case mhfpacket.OPERATE_JOINT_DISBAND:
		if guild.LeaderCharID == s.CharID && alliance.ParentGuildID == guild.ID {
			_, err = db.Exec("DELETE FROM guild_alliances WHERE id=$1", alliance.ID)
			if err != nil {
				s.Logger.Error("Failed to disband alliance", zap.Error(err))
			}
			s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		} else {
			s.Logger.Warn(
				"Non-owner of alliance attempted disband",
				zap.Uint32("CharID", s.CharID),
				zap.Uint32("AllyID", alliance.ID),
			)
			s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		}
	case mhfpacket.OPERATE_JOINT_LEAVE:
		if guild.LeaderCharID == s.CharID {
			if guild.ID == alliance.SubGuild1ID && alliance.SubGuild2ID > 0 {
				db.Exec(`UPDATE guild_alliances SET sub1_id = sub2_id, sub2_id = NULL WHERE id = $1`, alliance.ID)
			} else if guild.ID == alliance.SubGuild1ID && alliance.SubGuild2ID == 0 {
				db.Exec(`UPDATE guild_alliances SET sub1_id = NULL WHERE id = $1`, alliance.ID)
			} else {
				db.Exec(`UPDATE guild_alliances SET sub2_id = NULL WHERE id = $1`, alliance.ID)
			}
			// TODO: Handle deleting Alliance applications
			s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		} else {
			s.Logger.Warn(
				"Non-owner of guild attempted alliance leave",
				zap.Uint32("CharID", s.CharID),
			)
			s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		}
	case mhfpacket.OPERATE_JOINT_KICK:
		if alliance.ParentGuild.LeaderCharID == s.CharID {
			kickedGuildID := pkt.Data1.ReadUint32()
			if kickedGuildID == alliance.SubGuild1ID && alliance.SubGuild2ID > 0 {
				db.Exec(`UPDATE guild_alliances SET sub1_id = sub2_id, sub2_id = NULL WHERE id = $1`, alliance.ID)
			} else if kickedGuildID == alliance.SubGuild1ID && alliance.SubGuild2ID == 0 {
				db.Exec(`UPDATE guild_alliances SET sub1_id = NULL WHERE id = $1`, alliance.ID)
			} else {
				db.Exec(`UPDATE guild_alliances SET sub2_id = NULL WHERE id = $1`, alliance.ID)
			}
			s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		} else {
			s.Logger.Warn(
				"Non-owner of alliance attempted kick",
				zap.Uint32("CharID", s.CharID),
				zap.Uint32("AllyID", alliance.ID),
			)
			s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
		}
	default:
		s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
		panic(fmt.Sprintf("Unhandled operate joint action '%d'", pkt.Action))
	}
}

func HandleMsgMhfInfoJoint(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoJoint)
	bf := byteframe.NewByteFrame()
	alliance, err := service.GetAllianceData(pkt.AllianceID)
	if err != nil {
		s.DoAckSimpleFail(pkt.AckHandle, make([]byte, 4))
	} else {
		bf.WriteUint32(alliance.ID)
		bf.WriteUint32(uint32(alliance.CreatedAt.Unix()))
		bf.WriteUint16(alliance.TotalMembers)
		bf.WriteUint16(0x0000) // Unk
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
		bf.WriteUint32(alliance.ParentGuild.LeaderCharID)
		bf.WriteUint16(alliance.ParentGuild.Rank())
		bf.WriteUint16(alliance.ParentGuild.MemberCount)
		ps.Uint16(bf, alliance.ParentGuild.Name, true)
		ps.Uint16(bf, alliance.ParentGuild.LeaderName, true)
		if alliance.SubGuild1ID > 0 {
			bf.WriteUint32(alliance.SubGuild1ID)
			bf.WriteUint32(alliance.SubGuild1.LeaderCharID)
			bf.WriteUint16(alliance.SubGuild1.Rank())
			bf.WriteUint16(alliance.SubGuild1.MemberCount)
			ps.Uint16(bf, alliance.SubGuild1.Name, true)
			ps.Uint16(bf, alliance.SubGuild1.LeaderName, true)
		}
		if alliance.SubGuild2ID > 0 {
			bf.WriteUint32(alliance.SubGuild2ID)
			bf.WriteUint32(alliance.SubGuild2.LeaderCharID)
			bf.WriteUint16(alliance.SubGuild2.Rank())
			bf.WriteUint16(alliance.SubGuild2.MemberCount)
			ps.Uint16(bf, alliance.SubGuild2.Name, true)
			ps.Uint16(bf, alliance.SubGuild2.LeaderName, true)
		}
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
	}
}
