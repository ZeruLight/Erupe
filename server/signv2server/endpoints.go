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

const (
	NotificationDefault = iota
	NotificationNew
)

type LauncherBanner struct {
	Src  string `json:"src"`
	Link string `json:"link"`
}

type LauncherMessage struct {
	Message string `json:"message"`
	Date    int64  `json:"date"`
	Link    string `json:"link"`
	Kind    int    `json:"kind"`
}

type LauncherLink struct {
	Name string `json:"name"`
	Link string `json:"link"`
	Icon string `json:"icon"`
}

type LauncherResponse struct {
	Banners  []LauncherBanner  `json:"banners"`
	Messages []LauncherMessage `json:"messages"`
	Links    []LauncherLink    `json:"links"`
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

func (s *Server) newAuthData(userID uint32, userRights uint32, userToken string, characters []Character) AuthData {
	resp := AuthData{
		CurrentTS:     uint32(channelserver.TimeAdjusted().Unix()),
		ExpiryTS:      uint32(s.getReturnExpiry(userID).Unix()),
		EntranceCount: 1,
		User: User{
			Rights: userRights,
			Token:  userToken,
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
		resp.Notices = append(resp.Notices, strings.Join(s.erupeConfig.LoginNotices[:], "<PAGE>"))
	}
	return resp
}

func (s *Server) Launcher(w http.ResponseWriter, r *http.Request) {
	var respData LauncherResponse
	respData.Banners = []LauncherBanner{
		{
			Src:  "http://zerulight.cc/launcher/en/images/bnr/1030_0.jpg",
			Link: "http://localhost",
		},
		{
			Src:  "http://zerulight.cc/launcher/en/images/bnr/0801_3.jpg",
			Link: "http://localhost",
		},
		{
			Src:  "http://zerulight.cc/launcher/en/images/bnr/0705_3.jpg",
			Link: "http://localhost",
		},
		{
			Src:  "http://zerulight.cc/launcher/en/images/bnr/1211_11.jpg",
			Link: "http://localhost",
		},
		{
			Src:  "http://zerulight.cc/launcher/en/images/bnr/reg_mezefes.jpg",
			Link: "http://localhost",
		},
	}
	respData.Messages = []LauncherMessage{
		{
			Message: "Server Update 9.2 — Quest fixes,\nGacha support and tons of bug fixes!",
			Date:    time.Date(2023, 4, 1, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.com/channels/368424389416583169/929509970624532511/1003985850255818762",
			Kind:    NotificationNew,
		},
		{
			Message: "English Patch 4.1 — Fix \"Unknown\" weapons, NPC changes & Diva Support.",
			Date:    time.Date(2023, 2, 27, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.com/channels/368424389416583169/929509970624532511/969305400795078656",
			Kind:    NotificationNew,
		},
		{
			Message: "Server Update 9.1! Hunter Festival, Return worlds and NetCafe are back!",
			Date:    time.Date(2022, 11, 4, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.com/channels/368424389416583169/929509970624532511/969286397301248050",
			Kind:    NotificationDefault,
		},
		{
			Message: "Deerby & Supream have been updating Ferias! You can find any and all MHF info/data there!",
			Date:    time.Date(2022, 7, 7, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.gg/CFnzbhQ",
			Kind:    NotificationDefault,
		},
		{
			Message: "Server hosts, get Chakratos' Save Manager! Use it to enhance your Erupe server!",
			Date:    time.Date(2022, 7, 7, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.gg/CFnzbhQ",
			Kind:    NotificationDefault,
		},
		{
			Message: "Server Update 9.0 is out! Enjoy MezFes and all the other new content!",
			Date:    time.Date(2022, 8, 2, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.gg/CFnzbhQ",
			Kind:    NotificationDefault,
		},
		{
			Message: "English Community Translation 2 is here! Get the latest translation patch!",
			Date:    time.Date(2022, 5, 4, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.gg/CFnzbhQ",
			Kind:    NotificationDefault,
		},
		{
			Message: "Join the community Discord for future updates!",
			Date:    time.Date(2022, 5, 4, 0, 0, 0, 0, time.UTC).Unix(),
			Link:    "https://discord.gg/CFnzbhQ",
			Kind:    NotificationDefault,
		},
	}
	respData.Links = []LauncherLink{
		{
			Name: "GitHub",
			Link: "https://github.com/ZeruLight/Erupe",
			Icon: "https://cdn-icons-png.flaticon.com/512/25/25231.png",
		},
		{
			Name: "Discord",
			Link: "https://discord.gg/DnwcpXM488",
			Icon: "https://assets-global.website-files.com/6257adef93867e50d84d30e2/636e0a6a49cf127bf92de1e2_icon_clyde_blurple_RGB.png",
		},
		{
			Name: "Equal Dragon Weapon Info",
			Link: "https://discord.gg/DnwcpXM488",
			Icon: "",
		},
	}
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
	if characters == nil {
		characters = []Character{}
	}
	respData := s.newAuthData(userID, userRights, userToken, characters)
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

	userToken, err := s.createLoginToken(ctx, userID)
	if err != nil {
		s.logger.Error("Error registering login token", zap.Error(err))
		w.WriteHeader(500)
		return
	}
	respData := s.newAuthData(userID, userRights, userToken, []Character{})
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
