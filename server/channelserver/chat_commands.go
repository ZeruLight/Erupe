package channelserver

import (
	"crypto/rand"
	"encoding/hex"
	"erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/db"
	"erupe-ce/utils/mhfcid"
	"erupe-ce/utils/mhfcourse"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

type commandFunc func(s *Session, args []string) error

var commandMap = map[string]commandFunc{
	"ban":      ban,
	"timer":    timer,
	"psn":      psn,
	"reload":   reload,
	"kqf":      kqf,
	"rights":   rights,
	"course":   course,
	"ravi":     ravi,
	"teleport": teleport,
	"discord":  discord,
	"help":     help,
}

func executeCommand(s *Session, input string) error {
	args := strings.Split(input[len(config.GetConfig().CommandPrefix):], " ")
	if command, exists := commandMap[args[0]]; exists {
		if !s.isOp() {
			if commandConfig, exists := config.GetConfig().Commands[args[0]]; exists {
				if !commandConfig.Enabled {
					s.sendMessage(t("commands.disabled", v{"command": args[0]}))
					return nil
				}
			}
		}
		return command(s, args[1:])
	} else {
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func ban(s *Session, args []string) error {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}

	if !s.isOp() {
		s.sendMessage(t("commands.no_op", v{}))
		return nil
	}

	if len(args) < 1 {
		s.sendMessage(t("commands.ban.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
		return nil
	}

	cid := mhfcid.ConvertCID(args[0])
	if cid == 0 {
		s.sendMessage(t("commands.ban.error.invalid", v{}))
		return nil
	}

	var expiry time.Time
	if len(args) > 1 {
		duration, err := parseDuration(args[1])
		if err != nil {
			return err
		}
		expiry = time.Now().Add(duration)
	}

	var uid uint32
	var uname string
	err = db.QueryRow(`SELECT id, username FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, cid).Scan(&uid, &uname)
	if err == nil {
		if expiry.IsZero() {
			db.Exec(`INSERT INTO bans VALUES ($1)
                 				ON CONFLICT (user_id) DO UPDATE SET expires=NULL`, uid)
			s.sendMessage(t("commands.ban.success.permanent", v{"username": uname}))
		} else {
			db.Exec(`INSERT INTO bans VALUES ($1, $2)
                 				ON CONFLICT (user_id) DO UPDATE SET expires=$2`, uid, expiry)
			s.sendMessage(t("commands.ban.success.temporary", v{"username": uname, "expiry": expiry.Format(time.DateTime)}))
		}
		s.Server.DisconnectUser(uid)
	} else {
		s.sendMessage(t("commands.ban.error.invalid", v{}))
	}
	return nil
}

func timer(s *Session, _ []string) error {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var state bool
	db.QueryRow(`SELECT COALESCE(timer, false) FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.CharID).Scan(&state)
	db.Exec(`UPDATE users u SET timer=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, !state, s.CharID)
	if state {
		s.sendMessage(t("commands.timer.disabled", v{}))
	} else {
		s.sendMessage(t("commands.timer.enabled", v{}))
	}
	return nil
}

func psn(s *Session, args []string) error {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	if len(args) < 1 {
		s.sendMessage(t("commands.psn.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
		return nil
	}

	var exists int
	db.QueryRow(`SELECT count(*) FROM users WHERE psn_id = $1`, args[1]).Scan(&exists)
	if exists == 0 {
		_, err := db.Exec(`UPDATE users u SET psn_id=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, args[1], s.CharID)
		if err == nil {
			s.sendMessage(t("commands.psn.success", v{"psn": args[1]}))
		} else {
			return err
		}
	} else {
		s.sendMessage(t("commands.psn.error.exists", v{"psn": args[1]}))
	}
	return nil
}

func reload(s *Session, _ []string) error {
	s.sendMessage(t("commands.reload", v{}))
	var temp mhfpacket.MHFPacket
	for _, object := range s.stage.objects {
		if object.ownerCharID == s.CharID {
			continue
		}
		temp = &mhfpacket.MsgSysDeleteObject{ObjID: object.id}
		s.QueueSendMHF(temp)
	}
	for _, session := range s.Server.sessions {
		if s == session {
			continue
		}
		temp = &mhfpacket.MsgSysDeleteUser{CharID: session.CharID}
		s.QueueSendMHF(temp)
	}
	time.Sleep(500 * time.Millisecond)
	for _, session := range s.Server.sessions {
		if s == session {
			continue
		}
		temp = &mhfpacket.MsgSysInsertUser{CharID: session.CharID}
		s.QueueSendMHF(temp)
		for i := 0; i < 3; i++ {
			temp = &mhfpacket.MsgSysNotifyUserBinary{
				CharID:     session.CharID,
				BinaryType: uint8(i + 1),
			}
			s.QueueSendMHF(temp)
		}
	}
	for _, obj := range s.stage.objects {
		if obj.ownerCharID == s.CharID {
			continue
		}
		temp = &mhfpacket.MsgSysDuplicateObject{
			ObjID:       obj.id,
			X:           obj.x,
			Y:           obj.y,
			Z:           obj.z,
			Unk0:        0,
			OwnerCharID: obj.ownerCharID,
		}
		s.QueueSendMHF(temp)
	}
	return nil
}

func kqf(s *Session, args []string) error {
	if len(args) < 1 {
		s.sendMessage(t("commands.kqf.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
		return nil
	}

	if args[0] == "get" {
		s.sendMessage(t("commands.kqf.get", v{"kqf": fmt.Sprintf("%x", s.kqf)}))
		return nil
	} else if args[0] == "set" {
		if len(args) < 2 || len(args[1]) != 16 {
			s.sendMessage(t("commands.kqf.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
			return nil
		}

		hexd, _ := hex.DecodeString(args[1])
		s.kqf = hexd
		s.kqfOverride = true
		s.sendMessage(t("commands.kqf.set", v{}))
		return nil
	} else {
		s.sendMessage(t("commands.kqf.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
		return nil
	}
}

func rights(s *Session, args []string) error {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	if len(args) < 1 {
		s.sendMessage(t("commands.rights.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
		return nil
	}

	r, _ := strconv.Atoi(args[0])
	_, err = db.Exec("UPDATE users u SET rights=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", r, s.CharID)
	if err != nil {
		return err
	}

	s.sendMessage(t("commands.rights.success", v{"rights": args[0]}))
	return nil
}

func course(s *Session, args []string) error {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	if len(args) < 1 {
		s.sendMessage(t("commands.course.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
		return nil
	}

	for _, course := range mhfcourse.Courses() {
		for _, alias := range course.Aliases() {
			if strings.ToLower(args[1]) == strings.ToLower(alias) {
				if slices.Contains(config.GetConfig().Courses, config.Course{Name: course.Aliases()[0], Enabled: true}) {
					var delta, rightsInt uint32
					if mhfcourse.CourseExists(course.ID, s.courses) {
						ei := slices.IndexFunc(s.courses, func(c mhfcourse.Course) bool {
							for _, alias := range c.Aliases() {
								if strings.ToLower(args[1]) == strings.ToLower(alias) {
									return true
								}
							}
							return false
						})
						if ei != -1 {
							delta = uint32(-1 * math.Pow(2, float64(course.ID)))
							s.sendMessage(t("commands.course.disabled", v{"course": course.Aliases()[0]}))
						}
					} else {
						delta = uint32(math.Pow(2, float64(course.ID)))
						s.sendMessage(t("commands.course.enabled", v{"course": course.Aliases()[0]}))
					}
					err := db.QueryRow("SELECT rights FROM users u INNER JOIN characters c ON u.id = c.user_id WHERE c.id = $1", s.CharID).Scan(&rightsInt)
					if err != nil {
						return err
					} else {
						_, err = db.Exec("UPDATE users u SET rights=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", rightsInt+delta, s.CharID)
						if err != nil {
							return err
						}
					}
					updateRights(s)
				} else {
					s.sendMessage(t("commands.course.locked", v{"course": course.Aliases()[0]}))
				}
				return nil
			}
		}
	}
	return nil
}

func ravi(s *Session, args []string) error {
	if len(args) < 1 {
		s.sendMessage(t("commands.ravi.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
		return nil
	}

	if s.Server.getRaviSemaphore() == nil {
		s.sendMessage(t("commands.ravi.error.no_players", v{}))
	}

	switch args[0] {
	case "start":
		if s.Server.raviente.register[1] == 0 {
			s.Server.raviente.register[1] = s.Server.raviente.register[3]
			s.sendMessage(t("commands.ravi.start", v{}))
			s.notifyRavi()
		} else {
			s.sendMessage(t("commands.ravi.error.start", v{}))
		}
	case "multiplier":
		s.sendMessage(t("commands.ravi.multiplier", v{"multiplier": fmt.Sprintf("%.2f", s.Server.GetRaviMultiplier())}))
	case "sr", "sendres", "resurrection", "resurrect", "res":
		if config.GetConfig().ClientID != config.ZZ {
			s.sendMessage(t("commands.ravi.version", v{}))
			return nil
		}
		if s.Server.raviente.state[28] > 0 {
			s.sendMessage(t("commands.ravi.resurrect.send", v{}))
			s.Server.raviente.state[28] = 0
		} else {
			s.sendMessage(t("commands.ravi.resurrect.error", v{}))
		}
	case "ss", "sendsed":
		if config.GetConfig().ClientID != config.ZZ {
			s.sendMessage(t("commands.ravi.version", v{}))
			return nil
		}
		HP := s.Server.raviente.state[0] + s.Server.raviente.state[1] + s.Server.raviente.state[2] + s.Server.raviente.state[3] + s.Server.raviente.state[4]
		s.Server.raviente.support[1] = HP
	case "rs", "reqsed":
		if config.GetConfig().ClientID != config.ZZ {
			s.sendMessage(t("commands.ravi.version", v{}))
			return nil
		}
		HP := s.Server.raviente.state[0] + s.Server.raviente.state[1] + s.Server.raviente.state[2] + s.Server.raviente.state[3] + s.Server.raviente.state[4]
		s.Server.raviente.support[1] = HP + 1
	default:
		s.sendMessage(t("commands.ravi.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
	}

	return nil
}

func teleport(s *Session, args []string) error {
	if len(args) < 2 {
		s.sendMessage(t("commands.teleport.error.syntax", v{"prefix": config.GetConfig().CommandPrefix}))
		return nil
	}

	x, _ := strconv.ParseInt(args[0], 10, 16)
	y, _ := strconv.ParseInt(args[1], 10, 16)
	payload := byteframe.NewByteFrame()
	payload.SetLE()
	payload.WriteUint8(2) // SetState type(position == 2)
	payload.WriteInt16(int16(x))
	payload.WriteInt16(int16(y))
	s.QueueSendMHF(&mhfpacket.MsgSysCastedBinary{
		CharID:         s.CharID,
		MessageType:    BinaryMessageTypeState,
		RawDataPayload: payload.Data(),
	})
	s.sendMessage(t("commands.teleport.success", v{"x": fmt.Sprintf("%d", x), "y": fmt.Sprintf("%d", y)}))
	return nil
}

func discord(s *Session, _ []string) error {
	db, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var _token string
	err = db.QueryRow(`SELECT discord_token FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.CharID).Scan(&_token)
	if err != nil {
		randToken := make([]byte, 4)
		_, err = rand.Read(randToken)
		if err != nil {
			return err
		}
		_token = fmt.Sprintf("%x-%x", randToken[:2], randToken[2:])
		_, err = db.Exec(`UPDATE users u SET discord_token = $1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, _token, s.CharID)
		if err != nil {
			return err
		}
	}
	s.sendMessage(t("commands.discord.success", v{"token": _token}))
	return nil
}

func help(s *Session, _ []string) error {
	for command, _config := range config.GetConfig().Commands {
		if _config.Enabled || s.isOp() {
			s.sendMessage(fmt.Sprintf("%s%s: %s", config.GetConfig().CommandPrefix, command, _config.Description))
		}
	}
	return nil
}

func parseDuration(input string) (time.Duration, error) {
	unit := input[len(input)-1:]
	value := input[:len(input)-1]

	num, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %s", value)
	}

	switch unit {
	case "s":
		return time.Duration(num) * time.Second, nil
	case "m":
		return time.Duration(num) * time.Minute, nil
	case "h":
		return time.Duration(num) * time.Hour, nil
	case "d":
		return time.Duration(num) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("invalid duration unit: %s", unit)
	}
}
