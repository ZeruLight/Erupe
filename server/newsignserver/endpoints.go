package newsignserver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func randSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type LauncherMessage struct {
	Message string `json:"message"`
	Date    int64  `json:"date"`
	Link    string `json:"link"`
}

type Character struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsFemale  bool   `json:"isFemale" db:"is_female"`
	Weapon    int    `json:"weapon" db:"weapon_type"`
	HR        int    `json:"hr" db:"hrp"`
	GR        int    `json:"gr"`
	LastLogin int64  `json:"lastLogin" db:"last_login"`
}

func (s *Server) Launcher(w http.ResponseWriter, r *http.Request) {
	var respData struct {
		Important []LauncherMessage `json:"important"`
		Normal    []LauncherMessage `json:"normal"`
	}
	respData.Important = []LauncherMessage{
		{
			Message: "Server Update 9 Released!",
			Date:    time.Date(2022, 8, 2, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.com/channels/368424389416583169/929509970624532511/1003985850255818762",
		},
		{
			Message: "Eng 2.0 & Ravi Patch Released!",
			Date:    time.Date(2022, 5, 3, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.com/channels/368424389416583169/929509970624532511/969305400795078656",
		},
		{
			Message: "Launcher Patch V1.0 Released!",
			Date:    time.Date(2022, 4, 24, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.com/channels/368424389416583169/929509970624532511/969286397301248050",
		},
	}
	respData.Normal = []LauncherMessage{
		{
			Message: "Join the community Discord for updates!",
			Date:    time.Date(2022, 4, 24, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.gg/CFnzbhQ",
		},
	}
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respData)
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		w.Write([]byte("Invalid data received"))
		return
	}
	var (
		userID   int
		password string
	)
	err := s.db.QueryRow("SELECT id, password FROM users WHERE username = $1", reqData.Username).Scan(&userID, &password)
	if err == sql.ErrNoRows {
		w.WriteHeader(400)
		w.Write([]byte("Username does not exist"))
		return
	} else if err != nil {
		s.logger.Warn("SQL query error", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(password), []byte(reqData.Password)) != nil {
		w.WriteHeader(400)
		w.Write([]byte("Your password is incorrect"))
		return
	}

	var respData struct {
		Token      string      `json:"token"`
		Characters []Character `json:"characters"`
	}
	respData.Token, err = s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Warn("Error registering login token", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	respData.Characters, err = s.getCharactersForUser(ctx, userID)
	if err != nil {
		s.logger.Warn("Error getting characters from DB", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respData)
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		w.Write([]byte("Invalid data received"))
		return
	}
	s.logger.Info("Creating account", zap.String("username", reqData.Username))
	userID, err := s.createNewUser(ctx, reqData.Username, reqData.Password)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Constraint == "users_username_key" {
			w.WriteHeader(400)
			w.Write([]byte("User already exists"))
			return
		}
		s.logger.Error("Error checking user", zap.Error(err), zap.String("username", reqData.Username))
		w.WriteHeader(500)
		return
	}

	var respData struct {
		Token string `json:"token"`
	}
	respData.Token, err = s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Error("Error registering login token", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	json.NewEncoder(w).Encode(respData)
}

func (s *Server) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		w.Write([]byte("Invalid data received"))
		return
	}

	var respData struct {
		CharID int `json:"id"`
	}
	userID, err := s.userIDFromToken(ctx, reqData.Token)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	respData.CharID, err = s.createCharacter(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to create character", zap.Error(err), zap.String("token", reqData.Token))
		w.WriteHeader(500)
		return
	}
	json.NewEncoder(w).Encode(respData)
}

func (s *Server) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Token  string `json:"token"`
		CharID int    `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		w.Write([]byte("Invalid data received"))
		return
	}
	userID, err := s.userIDFromToken(ctx, reqData.Token)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	if err := s.deleteCharacter(ctx, userID, reqData.CharID); err != nil {
		s.logger.Error("Failed to delete character", zap.Error(err), zap.String("token", reqData.Token), zap.Int("charID", reqData.CharID))
		w.WriteHeader(500)
		return
	}
	json.NewEncoder(w).Encode(struct{}{})
}
