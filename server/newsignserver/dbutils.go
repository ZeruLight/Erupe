package newsignserver

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Server) createNewUser(ctx context.Context, username string, password string) (int, error) {
	// Create salted hash of user password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	var id int
	err = s.db.QueryRowContext(
		ctx, `
		INSERT INTO users (username, password, return_expires)
		VALUES ($1, $2, $3)
		RETURNING id
		`,
		username, string(passwordHash), time.Now().Add(time.Hour*24*30),
	).Scan(&id)
	return id, err
}

func (s *Server) createLoginToken(ctx context.Context, uid int) (string, error) {
	token := randSeq(16)
	_, err := s.db.ExecContext(ctx, "INSERT INTO sign_sessions (user_id, token) VALUES ($1, $2)", uid, token)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Server) userIDFromToken(ctx context.Context, token string) (int, error) {
	var userID int
	err := s.db.QueryRowContext(ctx, "SELECT user_id FROM sign_sessions WHERE token = $1", token).Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invalid login token")
	} else if err != nil {
		return 0, err
	}
	return userID, nil
}

func (s *Server) createCharacter(ctx context.Context, userID int) (int, error) {
	var charID int
	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM characters WHERE is_new_character = true AND user_id = $1",
		userID,
	).Scan(&charID)
	if err == sql.ErrNoRows {
		err = s.db.QueryRowContext(ctx, `
			INSERT INTO characters (
				user_id, is_female, is_new_character, name, unk_desc_string,
				hrp, gr, weapon_type, last_login
			)
			VALUES ($1, false, true, '', '', 0, 0, 0, $2)
			RETURNING id`,
			userID, uint32(time.Now().Unix()),
		).Scan(&charID)
	}
	return charID, err
}

func (s *Server) deleteCharacter(ctx context.Context, userID int, charID int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(
		ctx, `
		DELETE FROM login_boost_state
		WHERE char_id = $1`,
		charID,
	)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(
		ctx, `
		DELETE FROM guild_characters
		WHERE character_id = $1`,
		charID,
	)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(
		ctx, `
		DELETE FROM characters
		WHERE user_id = $1 AND id = $2`,
		userID, charID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Server) getCharactersForUser(ctx context.Context, uid int) ([]Character, error) {
	characters := make([]Character, 0)
	err := s.db.SelectContext(
		ctx, &characters, `
		SELECT id, name, is_female, weapon_type, hrp, gr, last_login
		FROM characters
		WHERE user_id = $1 AND deleted = false AND is_new_character = false ORDER BY id ASC`,
		uid,
	)
	if err != nil {
		return nil, err
	}
	return characters, nil
}
