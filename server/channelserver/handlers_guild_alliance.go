package channelserver

import (
	"fmt"
	"time"

	"erupe-ce/network/mhfpacket"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const allianceInfoSelectQuery = `
SELECT
ga.id,
ga.name,
created_at,
parent_id,
CASE
	WHEN sub1_id IS NULL THEN 0
	ELSE sub1_id
END,
CASE
	WHEN sub2_id IS NULL THEN 0
	ELSE sub2_id
END
FROM guild_alliances ga
`

type GuildAlliance struct {
	ID           uint32    `db:"id"`
	Name         string    `db:"name"`
	CreatedAt    time.Time `db:"created_at"`
	TotalMembers uint16

	ParentGuildID uint32 `db:"parent_id"`
	SubGuild1ID   uint32 `db:"sub1_id"`
	SubGuild2ID   uint32 `db:"sub2_id"`

	ParentGuild Guild
	SubGuild1   Guild
	SubGuild2   Guild
}

func GetAllianceData(s *Session, AllianceID uint32) (*GuildAlliance, error) {
	rows, err := s.server.db.Queryx(fmt.Sprintf(`
		%s
		WHERE ga.id = $1
	`, allianceInfoSelectQuery), AllianceID)
	if err != nil {
		s.logger.Error("Failed to retrieve alliance data from database", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	hasRow := rows.Next()
	if !hasRow {
		return nil, nil
	}

	return buildAllianceObjectFromDbResult(rows, err, s)
}

func buildAllianceObjectFromDbResult(result *sqlx.Rows, err error, s *Session) (*GuildAlliance, error) {
	alliance := &GuildAlliance{}

	err = result.StructScan(alliance)

	if err != nil {
		s.logger.Error("failed to retrieve alliance from database", zap.Error(err))
		return nil, err
	}

	parentGuild, err := GetGuildInfoByID(s, alliance.ParentGuildID)
	if err != nil {
		s.logger.Fatal("Failed to get parent guild info", zap.Error(err))
	} else {
		alliance.ParentGuild = *parentGuild
		alliance.TotalMembers += parentGuild.MemberCount
	}

	if alliance.SubGuild1ID > 0 {
		subGuild1, err := GetGuildInfoByID(s, alliance.SubGuild1ID)
		if err != nil {
			s.logger.Fatal("Failed to get sub guild 1 info", zap.Error(err))
		} else {
			alliance.SubGuild1 = *subGuild1
			alliance.TotalMembers += subGuild1.MemberCount
		}
	}

	if alliance.SubGuild2ID > 0 {
		subGuild2, err := GetGuildInfoByID(s, alliance.SubGuild2ID)
		if err != nil {
			s.logger.Fatal("Failed to get sub guild 2 info", zap.Error(err))
		} else {
			alliance.SubGuild2 = *subGuild2
			alliance.TotalMembers += subGuild2.MemberCount
		}
	}

	return alliance, nil
}

func handleMsgMhfCreateJoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateJoint)
	_, err := s.server.db.Exec("INSERT INTO guild_alliances (name, parent_id) VALUES ($1, $2)", pkt.Name, pkt.GuildID)
	if err != nil {
		s.logger.Fatal("Failed to create guild alliance in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x01, 0x01, 0x01, 0x01})
}

func handleMsgMhfOperateJoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateJoint)

	guild, err := GetGuildInfoByID(s, pkt.GuildID)
	if err != nil {
		s.logger.Fatal("Failed to get guild info", zap.Error(err))
	}
	alliance, err := GetAllianceData(s, pkt.AllianceID)
	if err != nil {
		s.logger.Fatal("Failed to get alliance info", zap.Error(err))
	}

	switch pkt.Action {
	case mhfpacket.OPERATE_JOINT_DISBAND:
		if guild.LeaderCharID == s.charID && alliance.ParentGuildID == guild.ID {
			_, err = s.server.db.Exec("DELETE FROM guild_alliances WHERE id=$1", alliance.ID)
			if err != nil {
				s.logger.Fatal("Failed to disband alliance", zap.Error(err))
			}
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		} else {
			s.logger.Warn(
				"Non-owner of alliance attempted disband",
				zap.Uint32("CharID", s.charID),
				zap.Uint32("AllyID", alliance.ID),
			)
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		}
	case mhfpacket.OPERATE_JOINT_LEAVE:
		if guild.LeaderCharID == s.charID {
			// delete alliance application
			// or leave alliance
		} else {
			s.logger.Warn(
				"Non-owner of guild attempted alliance leave",
				zap.Uint32("CharID", s.charID),
			)
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		}
	default:
		panic(fmt.Sprintf("Unhandled operate joint action '%d'", pkt.Action))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgMhfInfoJoint(s *Session, p mhfpacket.MHFPacket) {}
