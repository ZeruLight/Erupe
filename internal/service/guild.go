package service

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"erupe-ce/config"
	"erupe-ce/internal/model"
	"erupe-ce/utils/database"
	"erupe-ce/utils/logger"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Guild struct {
	ID            uint32              `db:"id"`
	Name          string              `db:"name"`
	MainMotto     uint8               `db:"main_motto"`
	SubMotto      uint8               `db:"sub_motto"`
	CreatedAt     time.Time           `db:"created_at"`
	MemberCount   uint16              `db:"member_count"`
	RankRP        uint32              `db:"rank_rp"`
	EventRP       uint32              `db:"event_rp"`
	RoomRP        uint16              `db:"room_rp"`
	RoomExpiry    time.Time           `db:"room_expiry"`
	Comment       string              `db:"comment"`
	PugiName1     string              `db:"pugi_name_1"`
	PugiName2     string              `db:"pugi_name_2"`
	PugiName3     string              `db:"pugi_name_3"`
	PugiOutfit1   uint8               `db:"pugi_outfit_1"`
	PugiOutfit2   uint8               `db:"pugi_outfit_2"`
	PugiOutfit3   uint8               `db:"pugi_outfit_3"`
	PugiOutfits   uint32              `db:"pugi_outfits"`
	Recruiting    bool                `db:"recruiting"`
	FestivalColor model.FestivalColor `db:"festival_color"`
	Souls         uint32              `db:"souls"`
	AllianceID    uint32              `db:"alliance_id"`
	Icon          *GuildIcon          `db:"icon"`

	model.GuildLeader
}

func RollbackTransaction(transaction *sql.Tx) {
	err := transaction.Rollback()
	logger := logger.Get()

	if err != nil {
		logger.Error("failed to rollback transaction", zap.Error(err))
	}
}

type GuildIcon struct {
	Parts []model.GuildIconPart
}

func (gi *GuildIcon) Scan(val interface{}) (err error) {
	switch v := val.(type) {
	case []byte:
		err = json.Unmarshal(v, &gi)
	case string:
		err = json.Unmarshal([]byte(v), &gi)
	}

	return
}

func (gi *GuildIcon) Value() (valuer driver.Value, err error) {
	return json.Marshal(gi)
}

func (g *Guild) Rank() uint16 {
	rpMap := []uint32{
		24, 48, 96, 144, 192, 240, 288, 360, 432,
		504, 600, 696, 792, 888, 984, 1080, 1200,
	}
	if config.GetConfig().ClientID <= config.Z2 {
		rpMap = []uint32{
			3500, 6000, 8500, 11000, 13500, 16000, 20000, 24000, 28000,
			33000, 38000, 43000, 48000, 55000, 70000, 90000, 120000,
		}
	}
	for i, u := range rpMap {
		if g.RankRP < u {
			if config.GetConfig().ClientID <= config.S6 && i >= 12 {
				return 12
			} else if config.GetConfig().ClientID <= config.F5 && i >= 13 {
				return 13
			} else if config.GetConfig().ClientID <= config.G32 && i >= 14 {
				return 14
			}
			return uint16(i)
		}
	}
	if config.GetConfig().ClientID <= config.S6 {
		return 12
	} else if config.GetConfig().ClientID <= config.F5 {
		return 13
	} else if config.GetConfig().ClientID <= config.G32 {
		return 14
	}
	return 17
}

func (guild *Guild) Save() error {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	_, err = db.Exec(`
		UPDATE guilds SET main_motto=$2, sub_motto=$3, comment=$4, pugi_name_1=$5, pugi_name_2=$6, pugi_name_3=$7,
		pugi_outfit_1=$8, pugi_outfit_2=$9, pugi_outfit_3=$10, pugi_outfits=$11, icon=$12, leader_id=$13 WHERE id=$1
	`, guild.ID, guild.MainMotto, guild.SubMotto, guild.Comment, guild.PugiName1, guild.PugiName2, guild.PugiName3,
		guild.PugiOutfit1, guild.PugiOutfit2, guild.PugiOutfit3, guild.PugiOutfits, guild.Icon, guild.GuildLeader.LeaderCharID)

	if err != nil {
		logger.Error("failed to update guild data", zap.Error(err), zap.Uint32("guildID", guild.ID))
		return err
	}

	return nil
}

func (guild *Guild) CreateApplication(charID uint32, applicationType model.GuildApplicationType, transaction *sql.Tx, actorId uint32) error {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	query := `
		INSERT INTO guild_applications (guild_id, character_id, actor_id, application_type)
		VALUES ($1, $2, $3, $4)
	`
	// Actor ID is the s.CharID
	if transaction == nil {
		_, err = db.Exec(query, guild.ID, charID, actorId, applicationType)
	} else {
		_, err = transaction.Exec(query, guild.ID, charID, actorId, applicationType)
	}

	if err != nil {
		logger.Error(
			"failed to add guild application",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", charID),
		)
		return err
	}

	return nil
}

func (guild *Guild) Disband(charID uint32) error {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	transaction, err := db.Begin()

	if err != nil {
		logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}

	_, err = transaction.Exec("DELETE FROM guild_characters WHERE guild_id = $1", guild.ID)

	if err != nil {
		logger.Error("failed to remove guild characters", zap.Error(err), zap.Uint32("guildId", guild.ID))
		RollbackTransaction(transaction)
		return err
	}

	_, err = transaction.Exec("DELETE FROM guilds WHERE id = $1", guild.ID)

	if err != nil {
		logger.Error("failed to remove guild", zap.Error(err), zap.Uint32("guildID", guild.ID))
		RollbackTransaction(transaction)
		return err
	}

	_, err = transaction.Exec("DELETE FROM guild_alliances WHERE parent_id=$1", guild.ID)

	if err != nil {
		logger.Error("failed to remove guild alliance", zap.Error(err), zap.Uint32("guildID", guild.ID))
		RollbackTransaction(transaction)
		return err
	}

	_, err = transaction.Exec("UPDATE guild_alliances SET sub1_id=sub2_id, sub2_id=NULL WHERE sub1_id=$1", guild.ID)

	if err != nil {
		logger.Error("failed to remove guild from alliance", zap.Error(err), zap.Uint32("guildID", guild.ID))
		RollbackTransaction(transaction)
		return err
	}

	_, err = transaction.Exec("UPDATE guild_alliances SET sub2_id=NULL WHERE sub2_id=$1", guild.ID)

	if err != nil {
		logger.Error("failed to remove guild from alliance", zap.Error(err), zap.Uint32("guildID", guild.ID))
		RollbackTransaction(transaction)
		return err
	}

	err = transaction.Commit()

	if err != nil {
		logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	logger.Info("Character disbanded guild", zap.Uint32("charID", charID), zap.Uint32("guildID", guild.ID))

	return nil
}

func (guild *Guild) RemoveCharacter(charID uint32) error {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	_, err = db.Exec("DELETE FROM guild_characters WHERE character_id=$1", charID)

	if err != nil {
		logger.Error(
			"failed to remove character from guild",
			zap.Error(err),
			zap.Uint32("charID", charID),
			zap.Uint32("guildID", guild.ID),
		)

		return err
	}

	return nil
}

func (guild *Guild) AcceptApplication(charID uint32) error {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	transaction, err := db.Begin()

	if err != nil {
		logger.Error("failed to start db transaction", zap.Error(err))
		return err
	}

	_, err = transaction.Exec(`DELETE FROM guild_applications WHERE character_id = $1`, charID)

	if err != nil {
		logger.Error("failed to accept character's guild application", zap.Error(err))
		RollbackTransaction(transaction)
		return err
	}

	_, err = transaction.Exec(`
		INSERT INTO guild_characters (guild_id, character_id, order_index)
		VALUES ($1, $2, (SELECT MAX(order_index) + 1 FROM guild_characters WHERE guild_id = $1))
	`, guild.ID, charID)

	if err != nil {
		logger.Error(
			"failed to add applicant to guild",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", charID),
		)
		RollbackTransaction(transaction)
		return err
	}

	err = transaction.Commit()

	if err != nil {
		logger.Error("failed to commit db transaction", zap.Error(err))
		RollbackTransaction(transaction)
		return err
	}

	return nil
}

// This is relying on the fact that invitation ID is also character ID right now
// if invitation ID changes, this will break.
func (guild *Guild) CancelInvitation(charID uint32) error {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	_, err = db.Exec(
		`DELETE FROM guild_applications WHERE character_id = $1 AND guild_id = $2 AND application_type = 'invited'`,
		charID, guild.ID,
	)

	if err != nil {
		logger.Error(
			"failed to cancel guild invitation",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", charID),
		)
		return err
	}

	return nil
}

func (guild *Guild) RejectApplication(charID uint32) error {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	_, err = db.Exec(
		`DELETE FROM guild_applications WHERE character_id = $1 AND guild_id = $2 AND application_type = 'applied'`,
		charID, guild.ID,
	)

	if err != nil {
		logger.Error(
			"failed to reject guild application",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", charID),
		)
		return err
	}

	return nil
}

func (guild *Guild) ArrangeCharacters(charIDs []uint32) error {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	transaction, err := db.Begin()

	if err != nil {
		logger.Error("failed to start db transaction", zap.Error(err))
		return err
	}

	for i, id := range charIDs {
		_, err := transaction.Exec("UPDATE guild_characters SET order_index = $1 WHERE character_id = $2", 2+i, id)

		if err != nil {
			err = transaction.Rollback()

			if err != nil {
				logger.Error("failed to rollback db transaction", zap.Error(err))
			}

			return err
		}
	}

	err = transaction.Commit()

	if err != nil {
		logger.Error("failed to commit db transaction", zap.Error(err))
		return err
	}

	return nil
}

func (guild *Guild) GetApplicationForCharID(charID uint32, applicationType model.GuildApplicationType) (*model.GuildApplication, error) {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	row := db.QueryRowx(`
		SELECT * from guild_applications WHERE character_id = $1 AND guild_id = $2 AND application_type = $3
	`, charID, guild.ID, applicationType)

	application := &model.GuildApplication{}

	err = row.StructScan(application)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		logger.Error(
			"failed to retrieve guild application for character",
			zap.Error(err),
			zap.Uint32("charID", charID),
			zap.Uint32("guildID", guild.ID),
		)
		return nil, err
	}

	return application, nil
}

func (guild *Guild) HasApplicationForCharID(charID uint32) (bool, error) {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	row := db.QueryRowx(`
		SELECT 1 from guild_applications WHERE character_id = $1 AND guild_id = $2
	`, charID, guild.ID)

	num := 0

	err = row.Scan(&num)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		logger.Error(
			"failed to retrieve guild applications for character",
			zap.Error(err),
			zap.Uint32("charID", charID),
			zap.Uint32("guildID", guild.ID),
		)
		return false, err
	}

	return true, nil
}

const GuildInfoSelectQuery = `
SELECT
	g.id,
	g.name,
	rank_rp,
	event_rp,
	room_rp,
	COALESCE(room_expiry, '1970-01-01') AS room_expiry,
	main_motto,
	sub_motto,
	created_at,
	leader_id,
	c.name AS leader_name,
	comment,
	COALESCE(pugi_name_1, '') AS pugi_name_1,
	COALESCE(pugi_name_2, '') AS pugi_name_2,
	COALESCE(pugi_name_3, '') AS pugi_name_3,
	pugi_outfit_1,
	pugi_outfit_2,
	pugi_outfit_3,
	pugi_outfits,
	recruiting,
	COALESCE((SELECT team FROM festa_registrations fr WHERE fr.guild_id = g.id), 'none') AS festival_color,
	COALESCE((SELECT SUM(fs.souls) FROM festa_submissions fs WHERE fs.guild_id=g.id), 0) AS souls,
	COALESCE((
		SELECT id FROM guild_alliances ga WHERE
	 	ga.parent_id = g.id OR
	 	ga.sub1_id = g.id OR
	 	ga.sub2_id = g.id
	), 0) AS alliance_id,
	icon,
	(SELECT count(1) FROM guild_characters gc WHERE gc.guild_id = g.id) AS member_count
	FROM guilds g
	JOIN guild_characters gc ON gc.character_id = leader_id
	JOIN characters c on leader_id = c.id
`

func CreateGuild(guildName string, charID uint32) (int32, error) {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	transaction, err := db.Begin()

	if err != nil {
		logger.Error("failed to start db transaction", zap.Error(err))
		return 0, err
	}

	if err != nil {
		panic(err)
	}

	guildResult, err := transaction.Query(
		"INSERT INTO guilds (name, leader_id) VALUES ($1, $2) RETURNING id",
		guildName, charID,
	)

	if err != nil {
		logger.Error("failed to create guild", zap.Error(err))
		RollbackTransaction(transaction)
		return 0, err
	}

	var guildId int32

	guildResult.Next()

	err = guildResult.Scan(&guildId)

	if err != nil {
		logger.Error("failed to retrieve guild ID", zap.Error(err))
		RollbackTransaction(transaction)
		return 0, err
	}

	err = guildResult.Close()

	if err != nil {
		logger.Error("failed to finalise query", zap.Error(err))
		RollbackTransaction(transaction)
		return 0, err
	}

	_, err = transaction.Exec(`
		INSERT INTO guild_characters (guild_id, character_id)
		VALUES ($1, $2)
	`, guildId, charID)

	if err != nil {
		logger.Error("failed to add character to guild", zap.Error(err))
		RollbackTransaction(transaction)
		return 0, err
	}

	err = transaction.Commit()

	if err != nil {
		logger.Error("failed to commit guild creation", zap.Error(err))
		return 0, err
	}

	return guildId, nil
}

func GetGuildInfoByID(guildID uint32) (*Guild, error) {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	rows, err := db.Queryx(fmt.Sprintf(`
		%s
		WHERE g.id = $1
		LIMIT 1
	`, GuildInfoSelectQuery), guildID)

	if err != nil {
		logger.Error("failed to retrieve guild", zap.Error(err), zap.Uint32("guildID", guildID))
		return nil, err
	}

	defer rows.Close()

	hasRow := rows.Next()

	if !hasRow {
		return nil, nil
	}

	return BuildGuildObjectFromDbResult(rows, err)
}

func GetGuildInfoByCharacterId(charID uint32) (*Guild, error) {
	db, err := database.GetDB()
	logger := logger.Get()

	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	rows, err := db.Queryx(fmt.Sprintf(`
		%s
		WHERE EXISTS(
				SELECT 1
				FROM guild_characters gc1
				WHERE gc1.character_id = $1
				  AND gc1.guild_id = g.id
			)
		   OR EXISTS(
				SELECT 1
				FROM guild_applications ga
				WHERE ga.character_id = $1
				  AND ga.guild_id = g.id
				  AND ga.application_type = 'applied'
			)
		LIMIT 1
	`, GuildInfoSelectQuery), charID)

	if err != nil {
		logger.Error("failed to retrieve guild for character", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}

	defer rows.Close()

	hasRow := rows.Next()

	if !hasRow {
		return nil, nil
	}

	return BuildGuildObjectFromDbResult(rows, err)
}

func BuildGuildObjectFromDbResult(result *sqlx.Rows, err error) (*Guild, error) {
	guild := &Guild{}
	logger := logger.Get()

	err = result.StructScan(guild)

	if err != nil {
		logger.Error("failed to retrieve guild data from database", zap.Error(err))
		return nil, err
	}

	return guild, nil
}
