package channelserver

import (
	"encoding/hex"
	"erupe-ce/common/byteframe"
	"erupe-ce/config"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"golang.org/x/exp/slices"
	"math"
	"math/rand"
	"strings"
	"time"

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

var commands map[string]config.Command

func init() {
	commands = make(map[string]config.Command)
	zapLogger, _ := zap.NewDevelopment()
	defer zapLogger.Sync()
	logger := zapLogger.Named("commands")
	cmds := config.ErupeConfig.Commands
	for _, cmd := range cmds {
		commands[cmd.Name] = cmd
		if cmd.Enabled {
			logger.Info(fmt.Sprintf("%s command is enabled, prefix: %s", cmd.Name, cmd.Prefix))
		} else {
			logger.Info(fmt.Sprintf("%s command is disabled", cmd.Name))
		}
	}
}

func sendDisabledCommandMessage(s *Session, cmd config.Command) {
	sendServerChatMessage(s, fmt.Sprintf("%s command is disabled", cmd.Name))
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
		CharID:         s.charID,
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}

	s.QueueSendMHF(castedBin)
}

func handleMsgSysCastBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCastBinary)
	tmp := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)

	if pkt.BroadcastType == 0x03 && pkt.MessageType == 0x03 && len(pkt.RawDataPayload) == 0x10 {
		if tmp.ReadUint16() == 0x0002 && tmp.ReadUint8() == 0x18 {
			_ = tmp.ReadBytes(9)
			tmp.SetLE()
			frame := tmp.ReadUint32()
			sendServerChatMessage(s, fmt.Sprintf("TIME : %d'%d.%03d (%dframe)", frame/30/60, frame/30%60, int(math.Round(float64(frame%30*100)/3)), frame))
		}
	}

	if s.server.erupeConfig.DevModeOptions.QuestDebugTools == true && s.server.erupeConfig.DevMode {
		if pkt.BroadcastType == 0x03 && pkt.MessageType == 0x02 && len(pkt.RawDataPayload) > 32 {
			// This is only correct most of the time
			tmp.ReadBytes(20)
			tmp.SetLE()
			x := tmp.ReadFloat32()
			y := tmp.ReadFloat32()
			z := tmp.ReadFloat32()
			s.logger.Debug("Coord", zap.Float32s("XYZ", []float32{x, y, z}))
		}
	}

	// Parse out the real casted binary payload
	var msgBinTargeted *binpacket.MsgBinTargeted
	var authorLen, msgLen uint16
	var msg []byte

	isDiceCommand := false
	if pkt.MessageType == BinaryMessageTypeChat {
		tmp.SetLE()
		tmp.Seek(int64(0), 0)
		_ = tmp.ReadUint32()
		authorLen = tmp.ReadUint16()
		msgLen = tmp.ReadUint16()
		msg = tmp.ReadNullTerminatedBytes()
	}

	// Customise payload
	realPayload := pkt.RawDataPayload
	if pkt.BroadcastType == BroadcastTypeTargeted {
		tmp.SetBE()
		tmp.Seek(int64(0), 0)
		msgBinTargeted = &binpacket.MsgBinTargeted{}
		err := msgBinTargeted.Parse(tmp)
		if err != nil {
			s.logger.Warn("Failed to parse targeted cast binary")
			return
		}
		realPayload = msgBinTargeted.RawDataPayload
	} else if pkt.MessageType == BinaryMessageTypeChat {
		if msgLen == 6 && string(msg) == "@dice" {
			isDiceCommand = true
			roll := byteframe.NewByteFrame()
			roll.WriteInt16(1) // Unk
			roll.SetLE()
			roll.WriteUint16(4) // Unk
			roll.WriteUint16(authorLen)
			rand.Seed(time.Now().UnixNano())
			dice := fmt.Sprintf("%d", rand.Intn(100)+1)
			roll.WriteUint16(uint16(len(dice) + 1))
			roll.WriteNullTerminatedBytes([]byte(dice))
			roll.WriteNullTerminatedBytes(tmp.ReadNullTerminatedBytes())
			realPayload = roll.Data()
		}
	}

	// Make the response to forward to the other client(s).
	resp := &mhfpacket.MsgSysCastedBinary{
		CharID:         s.charID,
		BroadcastType:  pkt.BroadcastType, // (The client never uses Type0 upon receiving)
		MessageType:    pkt.MessageType,
		RawDataPayload: realPayload,
	}

	// Send to the proper recipients.
	switch pkt.BroadcastType {
	case BroadcastTypeWorld:
		s.server.WorldcastMHF(resp, s, nil)
	case BroadcastTypeStage:
		if isDiceCommand {
			s.stage.BroadcastMHF(resp, nil) // send dice result back to caller
		} else {
			s.stage.BroadcastMHF(resp, s)
		}
	case BroadcastTypeServer:
		if pkt.MessageType == 1 {
			raviSema := getRaviSemaphore(s)
			if raviSema != "" {
				s.server.BroadcastMHF(resp, s)
			}
		} else {
			s.server.BroadcastMHF(resp, s)
		}
	case BroadcastTypeTargeted:
		for _, targetID := range (*msgBinTargeted).TargetCharIDs {
			char := s.server.FindSessionByCharID(targetID)

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

	// Handle chat
	if pkt.MessageType == BinaryMessageTypeChat {
		bf := byteframe.NewByteFrameFromBytes(realPayload)

		// IMPORTANT! Casted binary objects are sent _as they are in memory_,
		// this means little endian for LE CPUs, might be different for PS3/PS4/PSP/XBOX.
		bf.SetLE()

		chatMessage := &binpacket.MsgBinChat{}
		chatMessage.Parse(bf)

		fmt.Printf("Got chat message: %+v\n", chatMessage)

		// Discord integration
		if (pkt.BroadcastType == BroadcastTypeStage && s.stage.id == "sl1Ns200p0a0u0") || pkt.BroadcastType == BroadcastTypeWorld {
			s.server.DiscordChannelSend(chatMessage.SenderName, chatMessage.Message)
		}

		if strings.HasPrefix(chatMessage.Message, commands["Reload"].Prefix) {
			// Flush all objects and users and reload
			if commands["Reload"].Enabled {
				sendServerChatMessage(s, "Reloading players...")
				var temp mhfpacket.MHFPacket
				deleteNotif := byteframe.NewByteFrame()
				for _, object := range s.stage.objects {
					if object.ownerCharID == s.charID {
						continue
					}
					temp = &mhfpacket.MsgSysDeleteObject{ObjID: object.id}
					deleteNotif.WriteUint16(uint16(temp.Opcode()))
					temp.Build(deleteNotif, s.clientContext)
				}
				for _, session := range s.server.sessions {
					if s == session {
						continue
					}
					temp = &mhfpacket.MsgSysDeleteUser{CharID: session.charID}
					deleteNotif.WriteUint16(uint16(temp.Opcode()))
					temp.Build(deleteNotif, s.clientContext)
				}
				deleteNotif.WriteUint16(0x0010)
				s.QueueSend(deleteNotif.Data())
				time.Sleep(500 * time.Millisecond)
				reloadNotif := byteframe.NewByteFrame()
				for _, session := range s.server.sessions {
					if s == session {
						continue
					}
					temp = &mhfpacket.MsgSysInsertUser{CharID: session.charID}
					reloadNotif.WriteUint16(uint16(temp.Opcode()))
					temp.Build(reloadNotif, s.clientContext)
					for i := 0; i < 3; i++ {
						temp = &mhfpacket.MsgSysNotifyUserBinary{
							CharID:     session.charID,
							BinaryType: uint8(i + 1),
						}
						reloadNotif.WriteUint16(uint16(temp.Opcode()))
						temp.Build(reloadNotif, s.clientContext)
					}
				}
				for _, obj := range s.stage.objects {
					if obj.ownerCharID == s.charID {
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
					reloadNotif.WriteUint16(uint16(temp.Opcode()))
					temp.Build(reloadNotif, s.clientContext)
				}
				reloadNotif.WriteUint16(0x0010)
				s.QueueSend(reloadNotif.Data())
			} else {
				sendDisabledCommandMessage(s, commands["Reload"])
			}
		}

		if strings.HasPrefix(chatMessage.Message, commands["KeyQuest"].Prefix) {
			if commands["KeyQuest"].Enabled {
				if strings.HasPrefix(chatMessage.Message, "!kqf get") {
					sendServerChatMessage(s, fmt.Sprintf("KQF: %x", s.kqf))
				} else if strings.HasPrefix(chatMessage.Message, "!kqf set") {
					var hexs string
					n, numerr := fmt.Sscanf(chatMessage.Message, "!kqf set %s", &hexs)
					if numerr != nil || n != 1 || len(hexs) != 16 {
						sendServerChatMessage(s, "Error in command. Format: !kqf set xxxxxxxxxxxxxxxx")
					} else {
						hexd, _ := hex.DecodeString(hexs)
						s.kqf = hexd
						s.kqfOverride = true
						sendServerChatMessage(s, "KQF set, please switch Land/World")
					}
				}
			} else {
				sendDisabledCommandMessage(s, commands["KeyQuest"])
			}
		}

		if strings.HasPrefix(chatMessage.Message, commands["Rights"].Prefix) {
			// Set account rights
			if commands["Rights"].Enabled {
				var v uint32
				n, err := fmt.Sscanf(chatMessage.Message, "!rights %d", &v)
				if err != nil || n != 1 {
					sendServerChatMessage(s, "Error in command. Format: !rights n")
				} else {
					_, err = s.server.db.Exec("UPDATE users u SET rights=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", v, s.charID)
					if err == nil {
						sendServerChatMessage(s, fmt.Sprintf("Set rights integer: %d", v))
					}
				}
			} else {
				sendDisabledCommandMessage(s, commands["Rights"])
			}
		}

		if strings.HasPrefix(chatMessage.Message, commands["Course"].Prefix) {
			if commands["Course"].Enabled {
				var name string
				n, err := fmt.Sscanf(chatMessage.Message, "!course %s", &name)
				if err != nil || n != 1 {
					sendServerChatMessage(s, "Error in command. Format: !course <name>")
				} else {
					name = strings.ToLower(name)
					for _, course := range mhfpacket.Courses() {
						for _, alias := range course.Aliases {
							if strings.ToLower(name) == strings.ToLower(alias) {
								if slices.Contains(s.server.erupeConfig.Courses, config.Course{Name: course.Aliases[0], Enabled: true}) {
									if s.FindCourse(name).ID != 0 {
										ei := slices.IndexFunc(s.courses, func(c mhfpacket.Course) bool {
											for _, alias := range c.Aliases {
												if strings.ToLower(name) == strings.ToLower(alias) {
													return true
												}
											}
											return false
										})
										if ei != -1 {
											s.courses = append(s.courses[:ei], s.courses[ei+1:]...)
											sendServerChatMessage(s, fmt.Sprintf(`%s Course disabled`, course.Aliases[0]))
										}
									} else {
										s.courses = append(s.courses, course)
										sendServerChatMessage(s, fmt.Sprintf(`%s Course enabled`, course.Aliases[0]))
									}
									var newInt uint32
									for _, course := range s.courses {
										newInt += uint32(math.Pow(2, float64(course.ID)))
									}
									s.server.db.Exec("UPDATE users u SET rights=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", newInt, s.charID)
									updateRights(s)
								} else {
									sendServerChatMessage(s, fmt.Sprintf(`%s Course is locked`, course.Aliases[0]))
								}
							}
						}
					}
				}
			} else {
				sendDisabledCommandMessage(s, commands["Course"])
			}
		}

		if strings.HasPrefix(chatMessage.Message, commands["Raviente"].Prefix) {
			if commands["Raviente"].Enabled {
				if getRaviSemaphore(s) != "" {
					s.server.raviente.Lock()
					if !strings.HasPrefix(chatMessage.Message, "!ravi ") {
						sendServerChatMessage(s, "No Raviente command specified!")
					} else {
						if strings.HasPrefix(chatMessage.Message, "!ravi start") {
							if s.server.raviente.register.startTime == 0 {
								s.server.raviente.register.startTime = s.server.raviente.register.postTime
								sendServerChatMessage(s, "The Great Slaying will begin in a moment")
								s.notifyRavi()
							} else {
								sendServerChatMessage(s, "The Great Slaying has already begun!")
							}
						} else if strings.HasPrefix(chatMessage.Message, "!ravi sm") || strings.HasPrefix(chatMessage.Message, "!ravi setmultiplier") {
							var num uint16
							n, numerr := fmt.Sscanf(chatMessage.Message, "!ravi sm %d", &num)
							if numerr != nil || n != 1 {
								sendServerChatMessage(s, "Error in command. Format: !ravi sm n")
							} else if s.server.raviente.state.damageMultiplier == 1 {
								if num > 32 {
									sendServerChatMessage(s, "Raviente multiplier too high, defaulting to 32x")
									s.server.raviente.state.damageMultiplier = 32
								} else {
									sendServerChatMessage(s, fmt.Sprintf("Raviente multiplier set to %dx", num))
									s.server.raviente.state.damageMultiplier = uint32(num)
								}
							} else {
								sendServerChatMessage(s, fmt.Sprintf("Raviente multiplier is already set to %dx!", s.server.raviente.state.damageMultiplier))
							}
						} else if strings.HasPrefix(chatMessage.Message, "!ravi cm") || strings.HasPrefix(chatMessage.Message, "!ravi checkmultiplier") {
							sendServerChatMessage(s, fmt.Sprintf("Raviente multiplier is currently %dx", s.server.raviente.state.damageMultiplier))
						} else if strings.HasPrefix(chatMessage.Message, "!ravi sr") || strings.HasPrefix(chatMessage.Message, "!ravi sendres") {
							if s.server.raviente.state.stateData[28] > 0 {
								sendServerChatMessage(s, "Sending resurrection support!")
								s.server.raviente.state.stateData[28] = 0
							} else {
								sendServerChatMessage(s, "Resurrection support has not been requested!")
							}
						} else if strings.HasPrefix(chatMessage.Message, "!ravi ss") || strings.HasPrefix(chatMessage.Message, "!ravi sendsed") {
							sendServerChatMessage(s, "Sending sedation support if requested!")
							// Total BerRavi HP
							HP := s.server.raviente.state.stateData[0] + s.server.raviente.state.stateData[1] + s.server.raviente.state.stateData[2] + s.server.raviente.state.stateData[3] + s.server.raviente.state.stateData[4]
							s.server.raviente.support.supportData[1] = HP
						} else if strings.HasPrefix(chatMessage.Message, "!ravi rs") || strings.HasPrefix(chatMessage.Message, "!ravi reqsed") {
							sendServerChatMessage(s, "Requesting sedation support!")
							// Total BerRavi HP
							HP := s.server.raviente.state.stateData[0] + s.server.raviente.state.stateData[1] + s.server.raviente.state.stateData[2] + s.server.raviente.state.stateData[3] + s.server.raviente.state.stateData[4]
							s.server.raviente.support.supportData[1] = HP + 12
						} else {
							sendServerChatMessage(s, "Raviente command not recognised!")
						}
					}
					s.server.raviente.Unlock()
				} else {
					sendServerChatMessage(s, "No one has joined the Great Slaying!")
				}
			} else {
				sendDisabledCommandMessage(s, commands["Raviente"])
			}
		}

		if strings.HasPrefix(chatMessage.Message, commands["Teleport"].Prefix) {
			if commands["Teleport"].Enabled {
				var x, y int16
				n, err := fmt.Sscanf(chatMessage.Message, "!tele %d %d", &x, &y)
				if err != nil || n != 2 {
					sendServerChatMessage(s, "Invalid command. Usage:\"!tele 500 500\"")
				} else {
					sendServerChatMessage(s, fmt.Sprintf("Teleporting to %d %d", x, y))

					// Make the inside of the casted binary
					payload := byteframe.NewByteFrame()
					payload.SetLE()
					payload.WriteUint8(2) // SetState type(position == 2)
					payload.WriteInt16(x) // X
					payload.WriteInt16(y) // Y
					payloadBytes := payload.Data()

					s.QueueSendMHF(&mhfpacket.MsgSysCastedBinary{
						CharID:         s.charID,
						MessageType:    BinaryMessageTypeState,
						RawDataPayload: payloadBytes,
					})
				}
			} else {
				sendDisabledCommandMessage(s, commands["Teleport"])
			}
		}
	}
}

func handleMsgSysCastedBinary(s *Session, p mhfpacket.MHFPacket) {}
