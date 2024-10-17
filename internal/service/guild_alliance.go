package service

import (
	"erupe-ce/utils/db"
	"erupe-ce/utils/logger"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

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

const AllianceInfoSelectQuery = `
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

func GetAllianceData(AllianceID uint32) (*GuildAlliance, error) {
	db, err := db.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	rows, err := db.Queryx(fmt.Sprintf(`
		%s
		WHERE ga.id = $1
	`, AllianceInfoSelectQuery), AllianceID)
	if err != nil {
		logger.Error("Failed to retrieve alliance data from database", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	hasRow := rows.Next()
	if !hasRow {
		return nil, nil
	}

	return BuildAllianceObjectFromDbResult(rows, err)
}

func BuildAllianceObjectFromDbResult(result *sqlx.Rows, err error) (*GuildAlliance, error) {
	alliance := &GuildAlliance{}
	logger := logger.Get()

	err = result.StructScan(alliance)

	if err != nil {
		logger.Error("failed to retrieve alliance from database", zap.Error(err))
		return nil, err
	}

	parentGuild, err := GetGuildInfoByID(alliance.ParentGuildID)
	if err != nil {
		logger.Error("Failed to get parent guild info", zap.Error(err))
		return nil, err
	} else {
		alliance.ParentGuild = *parentGuild
		alliance.TotalMembers += parentGuild.MemberCount
	}

	if alliance.SubGuild1ID > 0 {
		subGuild1, err := GetGuildInfoByID(alliance.SubGuild1ID)
		if err != nil {
			logger.Error("Failed to get sub guild 1 info", zap.Error(err))
			return nil, err
		} else {
			alliance.SubGuild1 = *subGuild1
			alliance.TotalMembers += subGuild1.MemberCount
		}
	}

	if alliance.SubGuild2ID > 0 {
		subGuild2, err := GetGuildInfoByID(alliance.SubGuild2ID)
		if err != nil {
			logger.Error("Failed to get sub guild 2 info", zap.Error(err))
			return nil, err
		} else {
			alliance.SubGuild2 = *subGuild2
			alliance.TotalMembers += subGuild2.MemberCount
		}
	}

	return alliance, nil
}
