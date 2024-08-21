package channelserver

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type GuildMember struct {
	GuildID         uint32    `db:"guild_id"`
	CharID          uint32    `db:"character_id"`
	JoinedAt        time.Time `db:"joined_at"`
	Souls           uint32    `db:"souls"`
	RPToday         uint16    `db:"rp_today"`
	RPYesterday     uint16    `db:"rp_yesterday"`
	Name            string    `db:"name"`
	OrderIndex      uint16    `db:"order_index"`
	LastLogin       uint32    `db:"last_login"`
	Recruiter       bool      `db:"recruiter"`
	AvoidLeadership bool      `db:"avoid_leadership"`
	HR              uint16    `db:"hr"`
	GR              uint16    `db:"gr"`
	WeaponID        uint16    `db:"weapon_id"`
	WeaponType      uint8     `db:"weapon_type"`
}

func (gm GuildMember) CanRecruit() bool {
	if gm.Recruiter {
		return true
	}
	return gm.IsSubLeader()
}

func (gm GuildMember) IsSubLeader() bool {
	return gm.OrderIndex <= 3
}

func (gm GuildMember) IsLeader() bool {
	return gm.OrderIndex == 1
}

func (gm GuildMember) Save(s *Session) error {
	_, err := s.server.db.Exec("UPDATE guild_characters SET order_index=$1 WHERE character_id=$2", gm.OrderIndex, gm.CharID)

	if err != nil {
		s.logger.Error(
			"failed to update guild member data",
			zap.Error(err),
			zap.Uint32("charID", gm.CharID),
			zap.Uint32("guildID", gm.GuildID),
		)
		return err
	}
	return nil
}

const guildMembersSelectSQL = `
	SELECT
		g.id AS guild_id,
		joined_at,
		COALESCE((SELECT SUM(souls) FROM festa_submissions fs WHERE fs.character_id=c.id), 0) AS souls,
		COALESCE(rp_today, 0) AS rp_today,
		COALESCE(rp_yesterday, 0) AS rp_yesterday,
		c.name,
		c.id AS character_id,
		COALESCE(order_index, 0) AS order_index,
		c.last_login,
		COALESCE(recruiter, false) AS recruiter,
		COALESCE(avoid_leadership, false) AS avoid_leadership,
		c.hr,
		c.gr,
		c.weapon_id,
		c.weapon_type
		FROM guild_characters gc
		LEFT JOIN characters c ON c.id = gc.character_id
		LEFT JOIN guilds g ON g.id = gc.guild_id
`

func GetGuildApplications(s *Session, guildID uint32) []GuildApplication {
	var applications []GuildApplication
	var application GuildApplication
	rows, err := s.server.db.Queryx(`SELECT ga.character_id, ga.created_at, c.hr, c.gr, c.name FROM guild_applications ga LEFT JOIN characters c ON c.id = ga.character_id WHERE ga.guild_id=$1`, guildID)
	if err == nil {
		for rows.Next() {
			err = rows.StructScan(&application)
			if err != nil {
				continue
			}
			applications = append(applications, application)
		}
	}
	return applications
}

func GetGuildInvites(s *Session, guildID uint32) []GuildInvite {
	var invites []GuildInvite
	var invite GuildInvite
	rows, err := s.server.db.Queryx(`SELECT gi.character_id, gi.id, gi.created_at, gi.actor_id, c.hr, c.gr, c.name FROM guild_invites gi LEFT JOIN characters c ON c.id = gi.character_id WHERE gi.guild_id=$1`, guildID)
	if err == nil {
		for rows.Next() {
			err = rows.StructScan(&invite)
			if err != nil {
				continue
			}
			invites = append(invites, invite)
		}
	}
	return invites
}

func GetGuildMembers(s *Session, guildID uint32) []GuildMember {
	var members []GuildMember
	var member GuildMember
	rows, err := s.server.db.Queryx(fmt.Sprintf(`
			%s
			WHERE guild_id = $1
	`, guildMembersSelectSQL), guildID)
	if err != nil {
		s.logger.Error("Failed to retrieve membership data for guild", zap.Error(err), zap.Uint32("guildID", guildID))
		return members
	}

	for rows.Next() {
		err = rows.StructScan(&member)
		if err != nil {
			continue
		}
		members = append(members, member)
	}
	return members
}

func GetCharacterGuildData(s *Session, charID uint32) GuildMember {
	var member GuildMember
	err := s.server.db.QueryRowx(fmt.Sprintf("%s WHERE character_id=$1", guildMembersSelectSQL), charID).StructScan(&member)
	if err != nil {
		s.logger.Error("Failed to retrieve membership data for character", zap.Error(err), zap.Uint32("charID", charID))
	}
	return member
}
