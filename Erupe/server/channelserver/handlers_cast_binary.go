package channelserver

import (
	"fmt"
	"strings"
	"math"
	"math/rand"
	"time"

	"github.com/Andoryuuta/byteframe"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
)

// MSG_SYS_CAST[ED]_BINARY types enum
const (
	BinaryMessageTypeState      = 0
	BinaryMessageTypeChat       = 1
	BinaryMessageTypeMailNotify = 4
	BinaryMessageTypeEmote      = 6
)

// MSG_SYS_CAST[ED]_BINARY broadcast types enum
const (
	BroadcastTypeTargeted  = 0x01
	BroadcastTypeStage     = 0x03
	BroadcastTypeSemaphore = 0x06
	BroadcastTypeWorld     = 0x0a
)

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
			roll.WriteUint16(uint16(len(dice)+1))
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
		s.server.BroadcastMHF(resp, s)
	case BroadcastTypeStage:
		if isDiceCommand {
			s.stage.BroadcastMHF(resp, nil) // send dice result back to caller
		} else {
			s.stage.BroadcastMHF(resp, s)
		}
	case BroadcastTypeSemaphore:
		if pkt.MessageType == 1 {
			var session *Semaphore
			if _, exists := s.server.semaphore["hs_l0u3B51J9k3"]; exists {
				session = s.server.semaphore["hs_l0u3B51J9k3"]
			} else if _, exists := s.server.semaphore["hs_l0u3B5129k3"]; exists {
				session = s.server.semaphore["hs_l0u3B5129k3"]
			} else if _, exists := s.server.semaphore["hs_l0u3B512Ak3"]; exists {
				session = s.server.semaphore["hs_l0u3B512Ak3"]
			}
			(*session).BroadcastMHF(resp, s)
		} else {
			s.Lock()
			if s.stage != nil {
				s.stage.BroadcastMHF(resp, s)
			}
			s.Unlock()
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
		if chatMessage.Type == binpacket.ChatTypeLocal || chatMessage.Type == binpacket.ChatTypeParty {
			s.server.DiscordChannelSend(chatMessage.SenderName, chatMessage.Message)
		}

		// RAVI COMMANDS V2
		if strings.HasPrefix(chatMessage.Message, "!ravi") {
			if checkRaviSemaphore(s) {
				s.server.raviente.Lock()
				if !strings.HasPrefix(chatMessage.Message, "!ravi ") {
					sendServerChatMessage(s, "No Raviente command specified!")
				} else {
					if strings.HasPrefix(chatMessage.Message, "!ravi start") {
						if s.server.raviente.register.startTime == 0 {
							s.server.raviente.register.startTime = s.server.raviente.register.postTime
							sendServerChatMessage(s, "The Great Slaying will begin in a moment")
							s.notifyall()
						} else {
							sendServerChatMessage(s, "The Great Slaying has already begun!")
						}
					} else if strings.HasPrefix(chatMessage.Message, "!ravi sm") || strings.HasPrefix(chatMessage.Message, "!ravi setmultiplier") {
						var num uint16
						n, numerr := fmt.Sscanf(chatMessage.Message, "!ravi sm %d", &num)
						if numerr != nil || n != 1 {
							sendServerChatMessage(s, "Error in command. Format: !ravi sm n")
						} else if s.server.raviente.state.damageMultiplier == 1 {
							if num > 65535 {
								sendServerChatMessage(s, "Raviente multiplier too high, defaulting to 20x")
								s.server.raviente.state.damageMultiplier = 65535
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
			} else {
				sendServerChatMessage(s, "No one has joined the Great Slaying!")
			}
			s.server.raviente.Unlock()
		}
		// END RAVI COMMANDS V2

		if strings.HasPrefix(chatMessage.Message, "!tele ") {
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
		}
	}
}

func handleMsgSysCastedBinary(s *Session, p mhfpacket.MHFPacket) {}
