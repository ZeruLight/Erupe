package channelserver

import (
	"erupe-ce/config"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/db"
	"erupe-ce/utils/logger"
	"erupe-ce/utils/token"
	"sync"

	"fmt"
	"math"
	"strings"

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
		for cmd, conf := range cmds {
			if conf.Enabled {
				commandLogger.Info(fmt.Sprintf("%s: Enabled", cmd))
			} else {
				commandLogger.Info(fmt.Sprintf("%s: Disabled", cmd))
			}
		}
	})
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
				s.sendMessage(t("timer", v{
					"hours":        fmt.Sprintf("%d", int(frame/30/60/60)),
					"minutes":      fmt.Sprintf("%d", frame/30/60),
					"seconds":      fmt.Sprintf("%d", frame/30%60),
					"milliseconds": fmt.Sprintf("%d", int(math.Round(float64(frame%30*100)/3))),
					"frames":       fmt.Sprintf("%d", frame),
				}))
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
				err = executeCommand(s, chatMessage.Message)
				if err != nil {
					s.Logger.Error(fmt.Sprintf("Failed to execute command: %s", err))
				}
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
