package channelserver

import (
	"crypto/rand"
	"encoding/hex"
	"erupe-ce/config"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/db"
	"erupe-ce/utils/logger"
	"erupe-ce/utils/mhfcid"
	"erupe-ce/utils/mhfcourse"
	"erupe-ce/utils/token"
	"sync"

	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"go.uber.org/zap"
)

// MSG_SYS_CAST[ED]_BINARY types enum
const (
	BinaryMessageTypeState      = 0
	BinaryMessageTypeChat       = 1
	BinaryMessageTypeQuest      = 2
	BinaryMessageTypeData       = 3
	BinaryMessageTypeMailNotify = 4
	BinaryMessageTypeEmote      = 6
)

// MSG_SYS_CAST[ED]_BINARY broadcast types enum
const (
	BroadcastTypeTargeted = 0x01
	BroadcastTypeStage    = 0x03
	BroadcastTypeServer   = 0x06
	BroadcastTypeWorld    = 0x0a
)

var (
	commands   map[string]config.Command
	once       sync.Once  // Ensures that initialization happens only once
	commandsMu sync.Mutex // Mutex to ensure thread safety for commands map
)

func (server *ChannelServer) initCommands() {
	once.Do(func() {
		commands = make(map[string]config.Command)

		commandLogger := logger.Get().Named("command")
		cmds := config.GetConfig().Commands
		for _, cmd := range cmds {
			commands[cmd.Name] = cmd
			if cmd.Enabled {
				commandLogger.Info(fmt.Sprintf("%s: Enabled, prefix: %s", cmd.Name, cmd.Prefix))
			} else {
				commandLogger.Info(fmt.Sprintf("%s: Disabled", cmd.Name))
			}
		}
	})
}

func sendDisabledCommandMessage(s *Session, cmd config.Command) {
	sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.disabled, cmd.Name))
}

func sendServerChatMessage(s *Session, message string) {
	// Make the inside of the casted binary
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	msgBinChat := &binpacket.MsgBinChat{
		Unk0:       0,
		Type:       5,
		Flags:      0x80,
		Message:    message,
		SenderName: "Erupe",
	}
	msgBinChat.Build(bf)

	castedBin := &mhfpacket.MsgSysCastedBinary{
		CharID:         0,
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}

	s.QueueSendMHF(castedBin)
}

func parseChatCommand(s *Session, command string) {
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	args := strings.Split(command[len(config.GetConfig().CommandPrefix):], " ")
	switch args[0] {
	case commands["Ban"].Prefix:
		if s.isOp() {
			if len(args) > 1 {
				var expiry time.Time
				if len(args) > 2 {
					var length int
					var unit string
					n, err := fmt.Sscanf(args[2], `%d%s`, &length, &unit)
					if err == nil && n == 2 {
						switch unit {
						case "s", "second", "seconds":
							expiry = time.Now().Add(time.Duration(length) * time.Second)
						case "m", "mi", "minute", "minutes":
							expiry = time.Now().Add(time.Duration(length) * time.Minute)
						case "h", "hour", "hours":
							expiry = time.Now().Add(time.Duration(length) * time.Hour)
						case "d", "day", "days":
							expiry = time.Now().Add(time.Duration(length) * time.Hour * 24)
						case "mo", "month", "months":
							expiry = time.Now().Add(time.Duration(length) * time.Hour * 24 * 30)
						case "y", "year", "years":
							expiry = time.Now().Add(time.Duration(length) * time.Hour * 24 * 365)
						}
					} else {
						sendServerChatMessage(s, s.Server.i18n.commands.ban.error)
						return
					}
				}
				cid := mhfcid.ConvertCID(args[1])
				if cid > 0 {
					var uid uint32
					var uname string
					err := database.QueryRow(`SELECT id, username FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, cid).Scan(&uid, &uname)
					if err == nil {
						if expiry.IsZero() {
							database.Exec(`INSERT INTO bans VALUES ($1)
                 				ON CONFLICT (user_id) DO UPDATE SET expires=NULL`, uid)
							sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.ban.success, uname))
						} else {
							database.Exec(`INSERT INTO bans VALUES ($1, $2)
                 				ON CONFLICT (user_id) DO UPDATE SET expires=$2`, uid, expiry)
							sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.ban.success, uname)+fmt.Sprintf(s.Server.i18n.commands.ban.length, expiry.Format(time.DateTime)))
						}
						s.Server.DisconnectUser(uid)
					} else {
						sendServerChatMessage(s, s.Server.i18n.commands.ban.noUser)
					}
				} else {
					sendServerChatMessage(s, s.Server.i18n.commands.ban.invalid)
				}
			} else {
				sendServerChatMessage(s, s.Server.i18n.commands.ban.error)
			}
		} else {
			sendServerChatMessage(s, s.Server.i18n.commands.noOp)
		}
	case commands["Timer"].Prefix:
		if commands["Timer"].Enabled || s.isOp() {
			var state bool
			database.QueryRow(`SELECT COALESCE(timer, false) FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.CharID).Scan(&state)
			database.Exec(`UPDATE users u SET timer=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, !state, s.CharID)
			if state {
				sendServerChatMessage(s, s.Server.i18n.commands.timer.disabled)
			} else {
				sendServerChatMessage(s, s.Server.i18n.commands.timer.enabled)
			}
		} else {
			sendDisabledCommandMessage(s, commands["Timer"])
		}
	case commands["PSN"].Prefix:
		if commands["PSN"].Enabled || s.isOp() {
			if len(args) > 1 {
				var exists int
				database.QueryRow(`SELECT count(*) FROM users WHERE psn_id = $1`, args[1]).Scan(&exists)
				if exists == 0 {
					_, err := database.Exec(`UPDATE users u SET psn_id=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, args[1], s.CharID)
					if err == nil {
						sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.psn.success, args[1]))
					}
				} else {
					sendServerChatMessage(s, s.Server.i18n.commands.psn.exists)
				}
			} else {
				sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.psn.error, commands["PSN"].Prefix))
			}
		} else {
			sendDisabledCommandMessage(s, commands["PSN"])
		}
	case commands["Reload"].Prefix:
		if commands["Reload"].Enabled || s.isOp() {
			sendServerChatMessage(s, s.Server.i18n.commands.reload)
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
		} else {
			sendDisabledCommandMessage(s, commands["Reload"])
		}
	case commands["KeyQuest"].Prefix:
		if commands["KeyQuest"].Enabled || s.isOp() {
			if config.GetConfig().ClientID < config.G10 {
				sendServerChatMessage(s, s.Server.i18n.commands.kqf.version)
			} else {
				if len(args) > 1 {
					if args[1] == "get" {
						sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.kqf.get, s.kqf))
					} else if args[1] == "set" {
						if len(args) > 2 && len(args[2]) == 16 {
							hexd, _ := hex.DecodeString(args[2])
							s.kqf = hexd
							s.kqfOverride = true
							sendServerChatMessage(s, s.Server.i18n.commands.kqf.set.success)
						} else {
							sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.kqf.set.error, commands["KeyQuest"].Prefix))
						}
					}
				}
			}
		} else {
			sendDisabledCommandMessage(s, commands["KeyQuest"])
		}
	case commands["Rights"].Prefix:
		if commands["Rights"].Enabled || s.isOp() {
			if len(args) > 1 {
				v, _ := strconv.Atoi(args[1])
				_, err := database.Exec("UPDATE users u SET rights=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", v, s.CharID)
				if err == nil {
					sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.rights.success, v))
				} else {
					sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.rights.error, commands["Rights"].Prefix))
				}
			} else {
				sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.rights.error, commands["Rights"].Prefix))
			}
		} else {
			sendDisabledCommandMessage(s, commands["Rights"])
		}
	case commands["Course"].Prefix:
		if commands["Course"].Enabled || s.isOp() {
			if len(args) > 1 {
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
										sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.course.disabled, course.Aliases()[0]))
									}
								} else {
									delta = uint32(math.Pow(2, float64(course.ID)))
									sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.course.enabled, course.Aliases()[0]))
								}
								err := database.QueryRow("SELECT rights FROM users u INNER JOIN characters c ON u.id = c.user_id WHERE c.id = $1", s.CharID).Scan(&rightsInt)
								if err == nil {
									database.Exec("UPDATE users u SET rights=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", rightsInt+delta, s.CharID)
								}
								updateRights(s)
							} else {
								sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.course.locked, course.Aliases()[0]))
							}
							return
						}
					}
				}
			} else {
				sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.course.error, commands["Course"].Prefix))
			}
		} else {
			sendDisabledCommandMessage(s, commands["Course"])
		}
	case commands["Raviente"].Prefix:
		if commands["Raviente"].Enabled || s.isOp() {
			if len(args) > 1 {
				if s.Server.getRaviSemaphore() != nil {
					switch args[1] {
					case "start":
						if s.Server.raviente.register[1] == 0 {
							s.Server.raviente.register[1] = s.Server.raviente.register[3]
							sendServerChatMessage(s, s.Server.i18n.commands.ravi.start.success)
							s.notifyRavi()
						} else {
							sendServerChatMessage(s, s.Server.i18n.commands.ravi.start.error)
						}
					case "cm", "check", "checkmultiplier", "multiplier":
						sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.ravi.multiplier, s.Server.GetRaviMultiplier()))
					case "sr", "sendres", "resurrection", "ss", "sendsed", "rs", "reqsed":
						if config.GetConfig().ClientID == config.ZZ {
							switch args[1] {
							case "sr", "sendres", "resurrection":
								if s.Server.raviente.state[28] > 0 {
									sendServerChatMessage(s, s.Server.i18n.commands.ravi.res.success)
									s.Server.raviente.state[28] = 0
								} else {
									sendServerChatMessage(s, s.Server.i18n.commands.ravi.res.error)
								}
							case "ss", "sendsed":
								sendServerChatMessage(s, s.Server.i18n.commands.ravi.sed.success)
								// Total BerRavi HP
								HP := s.Server.raviente.state[0] + s.Server.raviente.state[1] + s.Server.raviente.state[2] + s.Server.raviente.state[3] + s.Server.raviente.state[4]
								s.Server.raviente.support[1] = HP
							case "rs", "reqsed":
								sendServerChatMessage(s, s.Server.i18n.commands.ravi.request)
								// Total BerRavi HP
								HP := s.Server.raviente.state[0] + s.Server.raviente.state[1] + s.Server.raviente.state[2] + s.Server.raviente.state[3] + s.Server.raviente.state[4]
								s.Server.raviente.support[1] = HP + 1
							}
						} else {
							sendServerChatMessage(s, s.Server.i18n.commands.ravi.version)
						}
					default:
						sendServerChatMessage(s, s.Server.i18n.commands.ravi.error)
					}
				} else {
					sendServerChatMessage(s, s.Server.i18n.commands.ravi.noPlayers)
				}
			} else {
				sendServerChatMessage(s, s.Server.i18n.commands.ravi.error)
			}
		} else {
			sendDisabledCommandMessage(s, commands["Raviente"])
		}
	case commands["Teleport"].Prefix:
		if commands["Teleport"].Enabled || s.isOp() {
			if len(args) > 2 {
				x, _ := strconv.ParseInt(args[1], 10, 16)
				y, _ := strconv.ParseInt(args[2], 10, 16)
				payload := byteframe.NewByteFrame()
				payload.SetLE()
				payload.WriteUint8(2)        // SetState type(position == 2)
				payload.WriteInt16(int16(x)) // X
				payload.WriteInt16(int16(y)) // Y
				payloadBytes := payload.Data()
				s.QueueSendMHF(&mhfpacket.MsgSysCastedBinary{
					CharID:         s.CharID,
					MessageType:    BinaryMessageTypeState,
					RawDataPayload: payloadBytes,
				})
				sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.teleport.success, x, y))
			} else {
				sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.teleport.error, commands["Teleport"].Prefix))
			}
		} else {
			sendDisabledCommandMessage(s, commands["Teleport"])
		}
	case commands["Discord"].Prefix:
		if commands["Discord"].Enabled || s.isOp() {
			var _token string
			err := database.QueryRow(`SELECT discord_token FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.CharID).Scan(&_token)
			if err != nil {
				randToken := make([]byte, 4)
				rand.Read(randToken)
				_token = fmt.Sprintf("%x-%x", randToken[:2], randToken[2:])
				database.Exec(`UPDATE users u SET discord_token = $1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)`, _token, s.CharID)
			}
			sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.commands.discord.success, _token))
		} else {
			sendDisabledCommandMessage(s, commands["Discord"])
		}
	case commands["Help"].Prefix:
		if commands["Help"].Enabled || s.isOp() {
			for _, command := range commands {
				if command.Enabled || s.isOp() {
					sendServerChatMessage(s, fmt.Sprintf("%s%s: %s", config.GetConfig().CommandPrefix, command.Prefix, command.Description))
				}
			}
		} else {
			sendDisabledCommandMessage(s, commands["Help"])
		}
	}
}

func handleMsgSysCastBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCastBinary)
	tmp := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	if pkt.BroadcastType == 0x03 && pkt.MessageType == 0x03 && len(pkt.RawDataPayload) == 0x10 {
		if tmp.ReadUint16() == 0x0002 && tmp.ReadUint8() == 0x18 {
			var timer bool
			database.QueryRow(`SELECT COALESCE(timer, false) FROM users u WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)`, s.CharID).Scan(&timer)
			if timer {
				_ = tmp.ReadBytes(9)
				tmp.SetLE()
				frame := tmp.ReadUint32()
				sendServerChatMessage(s, fmt.Sprintf(s.Server.i18n.timer, frame/30/60/60, frame/30/60, frame/30%60, int(math.Round(float64(frame%30*100)/3)), frame))
			}
		}
	}

	if config.GetConfig().DebugOptions.QuestTools {
		if pkt.BroadcastType == 0x03 && pkt.MessageType == 0x02 && len(pkt.RawDataPayload) > 32 {
			// This is only correct most of the time
			tmp.ReadBytes(20)
			tmp.SetLE()
			x := tmp.ReadFloat32()
			y := tmp.ReadFloat32()
			z := tmp.ReadFloat32()
			s.Logger.Debug("Coord", zap.Float32s("XYZ", []float32{x, y, z}))
		}
	}

	// Parse out the real casted binary payload
	var msgBinTargeted *binpacket.MsgBinTargeted
	var message, author string
	var returnToSender bool
	if pkt.MessageType == BinaryMessageTypeChat {
		tmp.SetLE()
		tmp.Seek(8, 0)
		message = string(tmp.ReadNullTerminatedBytes())
		author = string(tmp.ReadNullTerminatedBytes())
	}

	// Customise payload
	realPayload := pkt.RawDataPayload
	if pkt.BroadcastType == BroadcastTypeTargeted {
		tmp.SetBE()
		tmp.Seek(0, 0)
		msgBinTargeted = &binpacket.MsgBinTargeted{}
		err := msgBinTargeted.Parse(tmp)
		if err != nil {
			s.Logger.Warn("Failed to parse targeted cast binary")
			return
		}
		realPayload = msgBinTargeted.RawDataPayload
	} else if pkt.MessageType == BinaryMessageTypeChat {
		if message == "@dice" {
			returnToSender = true
			m := binpacket.MsgBinChat{
				Type:       BinaryMessageTypeChat,
				Flags:      4,
				Message:    fmt.Sprintf(`%d`, token.RNG.Intn(100)+1),
				SenderName: author,
			}
			bf := byteframe.NewByteFrame()
			bf.SetLE()
			m.Build(bf)
			realPayload = bf.Data()
		} else {
			bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
			bf.SetLE()
			chatMessage := &binpacket.MsgBinChat{}
			chatMessage.Parse(bf)
			if strings.HasPrefix(chatMessage.Message, config.GetConfig().CommandPrefix) {
				parseChatCommand(s, chatMessage.Message)
				return
			}
			if (pkt.BroadcastType == BroadcastTypeStage && s.stage.id == "sl1Ns200p0a0u0") || pkt.BroadcastType == BroadcastTypeWorld {
				s.Server.DiscordChannelSend(chatMessage.SenderName, chatMessage.Message)
			}
		}
	}

	// Make the response to forward to the other client(s).
	resp := &mhfpacket.MsgSysCastedBinary{
		CharID:         s.CharID,
		BroadcastType:  pkt.BroadcastType, // (The client never uses Type0 upon receiving)
		MessageType:    pkt.MessageType,
		RawDataPayload: realPayload,
	}

	// Send to the proper recipients.
	switch pkt.BroadcastType {
	case BroadcastTypeWorld:
		s.Server.WorldcastMHF(resp, s, nil)
	case BroadcastTypeStage:
		if returnToSender {
			s.stage.BroadcastMHF(resp, nil)
		} else {
			s.stage.BroadcastMHF(resp, s)
		}
	case BroadcastTypeServer:
		if pkt.MessageType == 1 {
			raviSema := s.Server.getRaviSemaphore()
			if raviSema != nil {
				raviSema.BroadcastMHF(resp, s)
			}
		} else {
			s.Server.BroadcastMHF(resp, s)
		}
	case BroadcastTypeTargeted:
		for _, targetID := range (*msgBinTargeted).TargetCharIDs {
			char := s.Server.FindSessionByCharID(targetID)

			if char != nil {
				char.QueueSendMHF(resp)
			}
		}
	default:
		s.Lock()
		haveStage := s.stage != nil
		if haveStage {
			s.stage.BroadcastMHF(resp, s)
		}
		s.Unlock()
	}
}

func handleMsgSysCastedBinary(s *Session, p mhfpacket.MHFPacket) {}
