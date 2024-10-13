package sign

import (
	"database/sql"
	"errors"
	"erupe-ce/config"
	"erupe-ce/utils/db"
	"erupe-ce/utils/mhfcourse"
	"erupe-ce/utils/token"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (server *SignServer) newUserChara(uid uint32) error {
	var numNewChars int
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("SELECT COUNT(*) FROM characters WHERE user_id = $1 AND is_new_character = true", uid).Scan(&numNewChars)
	if err != nil {
		return err
	}

	// prevent users with an uninitialised character from creating more
	if numNewChars >= 1 {
		return err
	}

	_, err = database.Exec(`
		INSERT INTO characters (
			user_id, is_female, is_new_character, name, unk_desc_string,
			hr, gr, weapon_type, last_login)
		VALUES($1, False, True, '', '', 0, 0, 0, $2)`,
		uid,
		uint32(time.Now().Unix()),
	)
	if err != nil {
		return err
	}

	return nil
}

func (server *SignServer) registerDBAccount(username string, password string) (uint32, error) {
	var uid uint32
	server.logger.Info("Creating user", zap.String("User", username))
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	// Create salted hash of user password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	err = database.QueryRow("INSERT INTO users (username, password, return_expires) VALUES ($1, $2, $3) RETURNING id", username, string(passwordHash), time.Now().Add(time.Hour*24*30)).Scan(&uid)
	if err != nil {
		return 0, err
	}

	return uid, nil
}

type character struct {
	ID             uint32 `db:"id"`
	IsFemale       bool   `db:"is_female"`
	IsNewCharacter bool   `db:"is_new_character"`
	Name           string `db:"name"`
	UnkDescString  string `db:"unk_desc_string"`
	HR             uint16 `db:"hr"`
	GR             uint16 `db:"gr"`
	WeaponType     uint16 `db:"weapon_type"`
	LastLogin      uint32 `db:"last_login"`
}

func (server *SignServer) getCharactersForUser(uid uint32) ([]character, error) {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	characters := make([]character, 0)
	err = database.Select(&characters, "SELECT id, is_female, is_new_character, name, unk_desc_string, hr, gr, weapon_type, last_login FROM characters WHERE user_id = $1 AND deleted = false ORDER BY id", uid)
	if err != nil {
		return nil, err
	}
	return characters, nil
}

func (server *SignServer) getReturnExpiry(uid uint32) time.Time {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var returnExpiry, lastLogin time.Time
	database.Get(&lastLogin, "SELECT COALESCE(last_login, now()) FROM users WHERE id=$1", uid)
	if time.Now().Add((time.Hour * 24) * -90).After(lastLogin) {
		returnExpiry = time.Now().Add(time.Hour * 24 * 30)
		database.Exec("UPDATE users SET return_expires=$1 WHERE id=$2", returnExpiry, uid)
	} else {
		err := database.Get(&returnExpiry, "SELECT return_expires FROM users WHERE id=$1", uid)
		if err != nil {
			returnExpiry = time.Now()
			database.Exec("UPDATE users SET return_expires=$1 WHERE id=$2", returnExpiry, uid)
		}
	}
	database.Exec("UPDATE users SET last_login=$1 WHERE id=$2", time.Now(), uid)
	return returnExpiry
}

func (server *SignServer) getLastCID(uid uint32) uint32 {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var lastPlayed uint32
	_ = database.QueryRow("SELECT last_character FROM users WHERE id=$1", uid).Scan(&lastPlayed)
	return lastPlayed
}

func (server *SignServer) getUserRights(uid uint32) uint32 {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var rights uint32
	if uid != 0 {
		_ = database.QueryRow("SELECT rights FROM users WHERE id=$1", uid).Scan(&rights)
		_, rights = mhfcourse.GetCourseStruct(rights)
	}
	return rights
}

type members struct {
	CID  uint32 // Local character ID
	ID   uint32 `db:"id"`
	Name string `db:"name"`
}

func (server *SignServer) getFriendsForCharacters(chars []character) []members {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	friends := make([]members, 0)
	for _, char := range chars {
		friendsCSV := ""
		err := database.QueryRow("SELECT friends FROM characters WHERE id=$1", char.ID).Scan(&friendsCSV)
		friendsSlice := strings.Split(friendsCSV, ",")
		friendQuery := "SELECT id, name FROM characters WHERE id="
		for i := 0; i < len(friendsSlice); i++ {
			friendQuery += friendsSlice[i]
			if i+1 != len(friendsSlice) {
				friendQuery += " OR id="
			}
		}
		charFriends := make([]members, 0)
		err = database.Select(&charFriends, friendQuery)
		if err != nil {
			continue
		}
		for i := range charFriends {
			charFriends[i].CID = char.ID
		}
		friends = append(friends, charFriends...)
	}
	return friends
}

func (server *SignServer) getGuildmatesForCharacters(chars []character) []members {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	guildmates := make([]members, 0)
	for _, char := range chars {
		var inGuild int
		_ = database.QueryRow("SELECT count(*) FROM guild_characters WHERE character_id=$1", char.ID).Scan(&inGuild)
		if inGuild > 0 {
			var guildID int
			err := database.QueryRow("SELECT guild_id FROM guild_characters WHERE character_id=$1", char.ID).Scan(&guildID)
			if err != nil {
				continue
			}
			charGuildmates := make([]members, 0)
			err = database.Select(&charGuildmates, "SELECT character_id AS id, c.name FROM guild_characters gc JOIN characters c ON c.id = gc.character_id WHERE guild_id=$1 AND character_id!=$2", guildID, char.ID)
			if err != nil {
				continue
			}
			for i := range charGuildmates {
				charGuildmates[i].CID = char.ID
			}
			guildmates = append(guildmates, charGuildmates...)
		}
	}
	return guildmates
}

func (server *SignServer) deleteCharacter(cid int, token string, tokenID uint32) error {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	if !server.validateToken(token, tokenID) {
		return errors.New("invalid token")
	}
	var isNew bool
	err = database.QueryRow("SELECT is_new_character FROM characters WHERE id = $1", cid).Scan(&isNew)
	if isNew {
		_, err = database.Exec("DELETE FROM characters WHERE id = $1", cid)
	} else {
		_, err = database.Exec("UPDATE characters SET deleted = true WHERE id = $1", cid)
	}
	if err != nil {
		return err
	}
	return nil
}

// Unused
func (server *SignServer) checkToken(uid uint32) (bool, error) {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var exists int
	err = database.QueryRow("SELECT count(*) FROM sign_sessions WHERE user_id = $1", uid).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists > 0 {
		return true, nil
	}
	return false, nil
}

func (server *SignServer) registerUidToken(uid uint32) (uint32, string, error) {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	_token := token.Generate(16)
	var tid uint32
	err = database.QueryRow(`INSERT INTO sign_sessions (user_id, token) VALUES ($1, $2) RETURNING id`, uid, _token).Scan(&tid)
	return tid, _token, err
}

func (server *SignServer) registerPsnToken(psn string) (uint32, string, error) {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	_token := token.Generate(16)
	var tid uint32
	err = database.QueryRow(`INSERT INTO sign_sessions (psn_id, token) VALUES ($1, $2) RETURNING id`, psn, _token).Scan(&tid)
	return tid, _token, err
}

func (server *SignServer) validateToken(token string, tokenID uint32) bool {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	query := `SELECT count(*) FROM sign_sessions WHERE token = $1`
	if tokenID > 0 {
		query += ` AND id = $2`
	}
	var exists int
	err = database.QueryRow(query, token, tokenID).Scan(&exists)
	if err != nil || exists == 0 {
		return false
	}
	return true
}

func (server *SignServer) validateLogin(user string, pass string) (uint32, RespID) {
	database, err := db.GetDB() // Capture both return values
	if err != nil {
		server.logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var uid uint32
	var passDB string
	err = database.QueryRow(`SELECT id, password FROM users WHERE username = $1`, user).Scan(&uid, &passDB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			server.logger.Info("User not found", zap.String("User", user))
			if config.GetConfig().AutoCreateAccount {
				uid, err = server.registerDBAccount(user, pass)
				if err == nil {
					return uid, SIGN_SUCCESS
				} else {
					return 0, SIGN_EABORT
				}
			}
			return 0, SIGN_EAUTH
		}
		return 0, SIGN_EABORT
	} else {
		if bcrypt.CompareHashAndPassword([]byte(passDB), []byte(pass)) == nil {
			var bans int
			err = database.QueryRow(`SELECT count(*) FROM bans WHERE user_id=$1 AND expires IS NULL`, uid).Scan(&bans)
			if err == nil && bans > 0 {
				return uid, SIGN_EELIMINATE
			}
			err = database.QueryRow(`SELECT count(*) FROM bans WHERE user_id=$1 AND expires > now()`, uid).Scan(&bans)
			if err == nil && bans > 0 {
				return uid, SIGN_ESUSPEND
			}
			return uid, SIGN_SUCCESS
		}
		return 0, SIGN_EPASS
	}
}