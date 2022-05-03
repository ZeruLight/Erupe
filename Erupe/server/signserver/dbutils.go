package signserver

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) newUserChara(username string) error {
	var id int
	err := s.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return err
	}

	var numNewChars int
	err = s.db.QueryRow("SELECT COUNT(*) FROM characters WHERE user_id = $1 AND is_new_character = true", id).Scan(&numNewChars)
	if err != nil {
		return err
	}

	// prevent users with an uninitialised character from creating more
	if numNewChars >= 1 {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO characters (
			user_id, is_female, is_new_character, name, unk_desc_string,
			hrp, gr, weapon_type, last_login)
		VALUES($1, False, True, '', '', 1, 0, 0, $2)`,
		id,
		uint32(time.Now().Unix()),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) registerDBAccount(username string, password string) error {
	// Create salted hash of user password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, string(passwordHash))
	if err != nil {
		return err
	}

	var id int
	err = s.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&id)
	if err != nil {
		return err
	}

	// Create a base new character.
	_, err = s.db.Exec(`
		INSERT INTO characters (
			user_id, is_female, is_new_character, name, unk_desc_string,
			hrp, gr, weapon_type, last_login)
		VALUES($1, False, True, '', '', 1, 0, 0, $2)`,
		id,
		uint32(time.Now().Unix()),
	)
	if err != nil {
		return err
	}

	return nil
}

type character struct {
	ID              uint32 `db:"id"`
	IsFemale        bool   `db:"is_female"`
	IsNewCharacter  bool   `db:"is_new_character"`
	Name            string `db:"name"`
	UnkDescString   string `db:"unk_desc_string"`
	HRP             uint16 `db:"hrp"`
	GR              uint16 `db:"gr"`
	WeaponType      uint16 `db:"weapon_type"`
	LastLogin       uint32 `db:"last_login"`
}

func (s *Server) getCharactersForUser(uid int) ([]character, error) {
	characters := []character{}
	err := s.db.Select(&characters, "SELECT id, is_female, is_new_character, name, unk_desc_string, hrp, gr, weapon_type, last_login FROM characters WHERE user_id = $1", uid)
	if err != nil {
		return nil, err
	}
	return characters, nil
}