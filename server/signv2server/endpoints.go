package signv2server

import (
	"database/sql"
	"encoding/json"
	"errors"
	_config "erupe-ce/config"
	"erupe-ce/server/channelserver"
	"net/http"
	"strings"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	NotificationDefault = iota
	NotificationNew
)

type LauncherResponse struct {
	Banners  []_config.SignV2Banner  `json:"banners"`
	Messages []_config.SignV2Message `json:"messages"`
	Links    []_config.SignV2Link    `json:"links"`
}

type User struct {
	TokenID uint32 `json:"tokenId"`
	Token   string `json:"token"`
	Rights  uint32 `json:"rights"`
}

type Character struct {
	ID        uint32 `json:"id"`
	Name      string `json:"name"`
	IsFemale  bool   `json:"isFemale" db:"is_female"`
	Weapon    uint32 `json:"weapon" db:"weapon_type"`
	HR        uint32 `json:"hr" db:"hrp"`
	GR        uint32 `json:"gr"`
	LastLogin int32  `json:"lastLogin" db:"last_login"`
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
	Notices       []string    `json:"notices"`
	User          User        `json:"user"`
	Characters    []Character `json:"characters"`
	MezFes        *MezFes     `json:"mezFes"`
	PatchServer   string      `json:"patchServer"`
}

type ExportData struct {
	Character map[string]interface{} `json:"character"`
}

func (s *Server) newAuthData(userID uint32, userRights uint32, userTokenID uint32, userToken string, characters []Character) AuthData {
	resp := AuthData{
		CurrentTS:     uint32(channelserver.TimeAdjusted().Unix()),
		ExpiryTS:      uint32(s.getReturnExpiry(userID).Unix()),
		EntranceCount: 1,
		User: User{
			Rights:  userRights,
			TokenID: userTokenID,
			Token:   userToken,
		},
		Characters:  characters,
		PatchServer: s.erupeConfig.SignV2.PatchServer,
		Notices:     []string{},
	}
	if s.erupeConfig.DevModeOptions.MaxLauncherHR {
		for i := range resp.Characters {
			resp.Characters[i].HR = 7
		}
	}
	stalls := []uint32{10, 3, 6, 9, 4, 8, 5, 7}
	if s.erupeConfig.GameplayOptions.MezFesSwitchMinigame {
		stalls[4] = 2
	}
	resp.MezFes = &MezFes{
		ID:           uint32(channelserver.TimeWeekStart().Unix()),
		Start:        uint32(channelserver.TimeWeekStart().Add(-time.Duration(s.erupeConfig.GameplayOptions.MezFesDuration) * time.Second).Unix()),
		End:          uint32(channelserver.TimeWeekNext().Unix()),
		SoloTickets:  s.erupeConfig.GameplayOptions.MezfesSoloTickets,
		GroupTickets: s.erupeConfig.GameplayOptions.MezfesGroupTickets,
		Stalls:       stalls,
	}
	if !s.erupeConfig.HideLoginNotice {
		resp.Notices = append(resp.Notices, strings.Join(s.erupeConfig.LoginNotices[:], "<PAGE>"))
	}
	return resp
}

func (s *Server) Launcher(w http.ResponseWriter, r *http.Request) {
	var respData LauncherResponse
	respData.Banners = s.erupeConfig.SignV2.Banners
	respData.Messages = s.erupeConfig.SignV2.Messages
	respData.Links = s.erupeConfig.SignV2.Links
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
		w.Write([]byte("username-error"))
		return
	} else if err != nil {
		s.logger.Warn("SQL query error", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(password), []byte(reqData.Password)) != nil {
		w.WriteHeader(400)
		w.Write([]byte("password-error"))
		return
	}

	userTokenID, userToken, err := s.createLoginToken(ctx, userID)
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
	if characters == nil {
		characters = []Character{}
	}
	respData := s.newAuthData(userID, userRights, userTokenID, userToken, characters)
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
		return
	}
	if reqData.Username == "" || reqData.Password == "" {
		w.WriteHeader(400)
		return
	}
	s.logger.Info("Creating account", zap.String("username", reqData.Username))
	userID, userRights, err := s.createNewUser(ctx, reqData.Username, reqData.Password)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Constraint == "users_username_key" {
			w.WriteHeader(400)
			w.Write([]byte("username-exists-error"))
			return
		}
		s.logger.Error("Error checking user", zap.Error(err), zap.String("username", reqData.Username))
		w.WriteHeader(500)
		return
	}

	userTokenID, userToken, err := s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Error("Error registering login token", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	respData := s.newAuthData(userID, userRights, userTokenID, userToken, []Character{})
	w.Header().Add("Content-Type", "application/json")
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
	if s.erupeConfig.DevModeOptions.MaxLauncherHR {
		character.HR = 7
	}
	w.Header().Add("Content-Type", "application/json")
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
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct{}{})
}

func (s *Server) ExportSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var reqData struct {
		Token  string `json:"token"`
		CharID uint32 `json:"charId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		s.logger.Error("JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		return
	}
	userID, err := s.userIDFromToken(ctx, reqData.Token)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	character, err := s.exportSave(ctx, userID, reqData.CharID)
	if err != nil {
		s.logger.Error("Failed to export save", zap.Error(err), zap.String("token", reqData.Token), zap.Uint32("charID", reqData.CharID))
		w.WriteHeader(500)
		return
	}
	save := ExportData{
		Character: character,
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(save)
}
