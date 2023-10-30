package signv2server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"erupe-ce/server/channelserver"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type LauncherMessage struct {
	Message string `json:"message"`
	Date    int64  `json:"date"`
	Link    string `json:"link"`
}

type LauncherResponse struct {
	Important []LauncherMessage `json:"important"`
	Normal    []LauncherMessage `json:"normal"`
}

type User struct {
	Token  string `json:"token"`
	Rights uint32 `json:"rights"`
}

type Character struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	IsFemale  bool   `json:"isFemale" db:"is_female"`
	Weapon    uint32 `json:"weapon" db:"weapon_type"`
	HR        uint32 `json:"hr" db:"hrp"`
	GR        uint32 `json:"gr"`
	LastLogin int64  `json:"lastLogin" db:"last_login"`
}

type MezFes struct {
	ID           uint32   `json:"id"`
	Start        uint32   `json:"start"`
	End          uint32   `json:"end"`
	SoloTickets  uint32   `json:"soloTickets"`
	GroupTickets uint32   `json:"groupTickets"`
	Stalls       []uint32 `json:"stalls"`
}

type AuthData struct {
	CurrentTS     uint32      `json:"currentTs"`
	ExpiryTS      uint32      `json:"expiryTs"`
	EntranceCount uint32      `json:"entranceCount"`
	Notifications []string    `json:"notifications"`
	User          User        `json:"user"`
	Characters    []Character `json:"characters"`
	MezFes        *MezFes     `json:"mezFes"`
}

func (s *Server) newAuthData(userID uint32, userRights uint32, userToken string, characters []Character) AuthData {
	resp := AuthData{
		CurrentTS:     uint32(channelserver.TimeAdjusted().Unix()),
		ExpiryTS:      uint32(s.getReturnExpiry(userID).Unix()),
		EntranceCount: 1,
		User: User{
			Rights: userRights,
			Token:  userToken,
		},
		Characters: characters,
	}
	if s.erupeConfig.DevModeOptions.MezFesEvent {
		stalls := []uint32{10, 3, 6, 9, 4, 8, 5, 7}
		if s.erupeConfig.DevModeOptions.MezFesAlt {
			stalls[4] = 2
		}
		resp.MezFes = &MezFes{
			ID:           uint32(channelserver.TimeWeekStart().Unix()),
			Start:        uint32(channelserver.TimeWeekStart().Unix()),
			End:          uint32(channelserver.TimeWeekNext().Unix()),
			SoloTickets:  s.erupeConfig.GameplayOptions.MezfesSoloTickets,
			GroupTickets: s.erupeConfig.GameplayOptions.MezfesGroupTickets,
			Stalls:       stalls,
		}
	}
	if !s.erupeConfig.HideLoginNotice {
		resp.Notifications = append(resp.Notifications, strings.Join(s.erupeConfig.LoginNotices[:], "<PAGE>"))
	}
	return resp
}

func (s *Server) Launcher(w http.ResponseWriter, r *http.Request) {
	var respData LauncherResponse
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
		userID     uint32
		userRights uint32
		password   string
	)
	err := s.db.QueryRow("SELECT id, password, rights FROM users WHERE username = $1", reqData.Username).Scan(&userID, &password, &userRights)
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

	userToken, err := s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Warn("Error registering login token", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	characters, err := s.getCharactersForUser(ctx, userID)
	if err != nil {
		s.logger.Warn("Error getting characters from DB", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	respData := s.newAuthData(userID, userRights, userToken, characters)
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
	userID, userRights, err := s.createNewUser(ctx, reqData.Username, reqData.Password)
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

	userToken, err := s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Error("Error registering login token", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	respData := s.newAuthData(userID, userRights, userToken, []Character{})
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

	userID, err := s.userIDFromToken(ctx, reqData.Token)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	character, err := s.createCharacter(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to create character", zap.Error(err), zap.String("token", reqData.Token))
		w.WriteHeader(500)
		return
	}
	json.NewEncoder(w).Encode(character)
}

func (s *Server) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Token  string `json:"token"`
		CharID uint32 `json:"charId"`
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
		s.logger.Error("Failed to delete character", zap.Error(err), zap.String("token", reqData.Token), zap.Uint32("charID", reqData.CharID))
		w.WriteHeader(500)
		return
	}
	json.NewEncoder(w).Encode(struct{}{})
}
