package channelserver

import (
	"encoding/binary"
	"encoding/hex"
	"erupe-ce/common/mhfcourse"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"crypto/rand"
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
	"math/bits"
)

// Temporary function to just return no results for a MSG_MHF_ENUMERATE* packet
func stubEnumerateNoResults(s *Session, ackHandle uint32) {
	enumBf := byteframe.NewByteFrame()
	enumBf.WriteUint32(0) // Entry count (count for quests, rankings, events, etc.)

	doAckBufSucceed(s, ackHandle, enumBf.Data())
}

func doAckEarthSucceed(s *Session, ackHandle uint32, data []*byteframe.ByteFrame) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(s.server.erupeConfig.DevModeOptions.EarthIDOverride))
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(uint32(len(data)))
	for i := range data {
		bf.WriteBytes(data[i].Data())
	}
	doAckBufSucceed(s, ackHandle, bf.Data())
}

func doAckBufSucceed(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: true,
		ErrorCode:        0,
		AckData:          data,
	})
}

func doAckBufFail(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: true,
		ErrorCode:        1,
		AckData:          data,
	})
}

func doAckSimpleSucceed(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: false,
		ErrorCode:        0,
		AckData:          data,
	})
}

func doAckSimpleFail(s *Session, ackHandle uint32, data []byte) {
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        ackHandle,
		IsBufferResponse: false,
		ErrorCode:        1,
		AckData:          data,
	})
}

func updateRights(s *Session) {
	rightsInt := uint32(2)
	s.server.db.QueryRow("SELECT rights FROM users u INNER JOIN characters c ON u.id = c.user_id WHERE c.id = $1", s.charID).Scan(&rightsInt)
	s.courses, rightsInt = mhfcourse.GetCourseStruct(rightsInt)
	update := &mhfpacket.MsgSysUpdateRight{
		ClientRespAckHandle: 0,
		Bitfield:            rightsInt,
		Rights:              s.courses,
		UnkSize:             0,
	}
	s.QueueSendMHF(update)
}

func handleMsgHead(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysExtendThreshold(s *Session, p mhfpacket.MHFPacket) {
	// No data aside from header, no resp required.
}

func handleMsgSysEnd(s *Session, p mhfpacket.MHFPacket) {
	// No data aside from header, no resp required.
}

func handleMsgSysNop(s *Session, p mhfpacket.MHFPacket) {
	// No data aside from header, no resp required.
}

func handleMsgSysAck(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysTerminalLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysTerminalLog)
	for i := range pkt.Entries {
		s.server.logger.Info("SysTerminalLog",
			zap.Uint8("Type1", pkt.Entries[i].Type1),
			zap.Uint8("Type2", pkt.Entries[i].Type2),
			zap.Int16("Unk0", pkt.Entries[i].Unk0),
			zap.Int32("Unk1", pkt.Entries[i].Unk1),
			zap.Int32("Unk2", pkt.Entries[i].Unk2),
			zap.Int32("Unk3", pkt.Entries[i].Unk3),
			zap.Int32s("Unk4", pkt.Entries[i].Unk4),
		)
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(pkt.LogID + 1) // LogID to use for requests after this.
	doAckSimpleSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgSysLogin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLogin)

	if !s.server.erupeConfig.DevModeOptions.DisableTokenCheck {
		var token string
		err := s.server.db.QueryRow("SELECT token FROM sign_sessions WHERE token=$1", pkt.LoginTokenString).Scan(&token)
		if err != nil {
			s.rawConn.Close()
			s.logger.Warn(fmt.Sprintf("Invalid login token, offending CID: (%d)", pkt.CharID0))
			return
		}
	}

	s.Lock()
	s.charID = pkt.CharID0
	s.token = pkt.LoginTokenString
	s.Unlock()

	updateRights(s)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(TimeAdjusted().Unix())) // Unix timestamp

	_, err := s.server.db.Exec("UPDATE servers SET current_players=$1 WHERE server_id=$2", len(s.server.sessions), s.server.ID)
	if err != nil {
		panic(err)
	}

	_, err = s.server.db.Exec("UPDATE sign_sessions SET server_id=$1, char_id=$2 WHERE token=$3", s.server.ID, s.charID, s.token)
	if err != nil {
		panic(err)
	}

	_, err = s.server.db.Exec("UPDATE characters SET last_login=$1 WHERE id=$2", TimeAdjusted().Unix(), s.charID)
	if err != nil {
		panic(err)
	}

	_, err = s.server.db.Exec("UPDATE users u SET last_character=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)", s.charID)
	if err != nil {
		panic(err)
	}

	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())

	updateRights(s)

	s.server.BroadcastMHF(&mhfpacket.MsgSysInsertUser{CharID: s.charID}, s)
}

func handleMsgSysLogout(s *Session, p mhfpacket.MHFPacket) {
	logoutPlayer(s)
}

func logoutPlayer(s *Session) {
	s.server.Lock()
	if _, exists := s.server.sessions[s.rawConn]; exists {
		delete(s.server.sessions, s.rawConn)
	}
	s.rawConn.Close()
	delete(s.server.objectIDs, s)
	s.server.Unlock()

	for _, stage := range s.server.stages {
		// Tell sessions registered to disconnecting players quest to unregister
		if stage.host != nil && stage.host.charID == s.charID {
			for _, sess := range s.server.sessions {
				for rSlot := range stage.reservedClientSlots {
					if sess.charID == rSlot && sess.stage != nil && sess.stage.id[3:5] != "Qs" {
						sess.QueueSendMHF(&mhfpacket.MsgSysStageDestruct{})
					}
				}
			}
		}
		for session := range stage.clients {
			if session.charID == s.charID {
				delete(stage.clients, session)
			}
		}
	}

	_, err := s.server.db.Exec("UPDATE sign_sessions SET server_id=NULL, char_id=NULL WHERE token=$1", s.token)
	if err != nil {
		panic(err)
	}

	_, err = s.server.db.Exec("UPDATE servers SET current_players=$1 WHERE server_id=$2", len(s.server.sessions), s.server.ID)
	if err != nil {
		panic(err)
	}

	var timePlayed int
	var sessionTime int
	_ = s.server.db.QueryRow("SELECT time_played FROM characters WHERE id = $1", s.charID).Scan(&timePlayed)
	sessionTime = int(TimeAdjusted().Unix()) - int(s.sessionStart)
	timePlayed += sessionTime

	var rpGained int
	if mhfcourse.CourseExists(30, s.courses) {
		rpGained = timePlayed / 900
		timePlayed = timePlayed % 900
		s.server.db.Exec("UPDATE characters SET cafe_time=cafe_time+$1 WHERE id=$2", sessionTime, s.charID)
	} else {
		rpGained = timePlayed / 1800
		timePlayed = timePlayed % 1800
	}

	s.server.db.Exec("UPDATE characters SET time_played = $1 WHERE id = $2", timePlayed, s.charID)

	treasureHuntUnregister(s)

	if s.stage == nil {
		return
	}

	s.server.BroadcastMHF(&mhfpacket.MsgSysDeleteUser{
		CharID: s.charID,
	}, s)

	s.server.Lock()
	for _, stage := range s.server.stages {
		if _, exists := stage.reservedClientSlots[s.charID]; exists {
			delete(stage.reservedClientSlots, s.charID)
		}
	}
	s.server.Unlock()

	removeSessionFromSemaphore(s)
	removeSessionFromStage(s)

	saveData, err := GetCharacterSaveData(s, s.charID)
	if err != nil || saveData == nil {
		s.logger.Error("Failed to get savedata")
		return
	}
	saveData.RP += uint16(rpGained)
	if saveData.RP >= s.server.erupeConfig.GameplayOptions.MaximumRP {
		saveData.RP = s.server.erupeConfig.GameplayOptions.MaximumRP
	}
	saveData.Save(s)
}

func handleMsgSysSetStatus(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysPing(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysPing)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysTime(s *Session, p mhfpacket.MHFPacket) {
	//pkt := p.(*mhfpacket.MsgSysTime)

	resp := &mhfpacket.MsgSysTime{
		GetRemoteTime: false,
		Timestamp:     uint32(TimeAdjusted().Unix()), // JP timezone
	}
	s.QueueSendMHF(resp)
}

func handleMsgSysIssueLogkey(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysIssueLogkey)

	// Make a random log key for this session.
	logKey := make([]byte, 16)
	_, err := rand.Read(logKey)
	if err != nil {
		panic(err)
	}

	// TODO(Andoryuuta): In the offical client, the log key index is off by one,
	// cutting off the last byte in _most uses_. Find and document these accordingly.
	s.Lock()
	s.logKey = logKey
	s.Unlock()

	// Issue it.
	resp := byteframe.NewByteFrame()
	resp.WriteBytes(logKey)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgSysRecordLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysRecordLog)
	// remove a client returning to town from reserved slots to make sure the stage is hidden from board
	delete(s.stage.reservedClientSlots, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysEcho(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysLockGlobalSema(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLockGlobalSema)
	var sgid string
	for _, channel := range s.server.Channels {
		for id := range channel.stages {
			if strings.HasSuffix(id, pkt.UserIDString) {
				sgid = channel.GlobalID
			}
		}
	}
	bf := byteframe.NewByteFrame()
	if len(sgid) > 0 && sgid != s.server.GlobalID {
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		ps.Uint16(bf, sgid, false)
	} else {
		bf.WriteUint8(2)
		bf.WriteUint8(0)
		ps.Uint16(bf, pkt.ServerChannelIDString, false)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgSysUnlockGlobalSema(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysUnlockGlobalSema)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgSysUpdateRight(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAuthQuery(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAuthTerminal(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysRightsReload(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysRightsReload)
	updateRights(s)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfTransitMessage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfTransitMessage)
	resp := byteframe.NewByteFrame()
	resp.WriteUint16(0)
	var count uint16
	switch pkt.SearchType {
	case 1: // CID
		bf := byteframe.NewByteFrameFromBytes(pkt.MessageData)
		CharID := bf.ReadUint32()
		for _, c := range s.server.Channels {
			for _, session := range c.sessions {
				if session.charID == CharID {
					count++
					sessionName := stringsupport.UTF8ToSJIS(session.Name)
					sessionStage := stringsupport.UTF8ToSJIS(session.stageID)
					resp.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(c.IP).To4()))
					resp.WriteUint16(c.Port)
					resp.WriteUint32(session.charID)
					resp.WriteBool(true)
					resp.WriteUint8(uint8(len(sessionName) + 1))
					resp.WriteUint16(uint16(len(c.userBinaryParts[userBinaryPartID{charID: session.charID, index: 3}])))
					resp.WriteBytes(make([]byte, 40))
					resp.WriteUint8(uint8(len(sessionStage) + 1))
					resp.WriteBytes(make([]byte, 8))
					resp.WriteNullTerminatedBytes(sessionName)
					resp.WriteBytes(c.userBinaryParts[userBinaryPartID{session.charID, 3}])
					resp.WriteNullTerminatedBytes(sessionStage)
				}
			}
		}
	case 2: // Name
		bf := byteframe.NewByteFrameFromBytes(pkt.MessageData)
		bf.ReadUint16() // lenSearchTerm
		bf.ReadUint16() // maxResults
		bf.ReadUint8()  // Unk
		searchTerm := stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
		for _, c := range s.server.Channels {
			for _, session := range c.sessions {
				if count == 100 {
					break
				}
				if strings.Contains(session.Name, searchTerm) {
					count++
					sessionName := stringsupport.UTF8ToSJIS(session.Name)
					sessionStage := stringsupport.UTF8ToSJIS(session.stageID)
					resp.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(c.IP).To4()))
					resp.WriteUint16(c.Port)
					resp.WriteUint32(session.charID)
					resp.WriteBool(true)
					resp.WriteUint8(uint8(len(sessionName) + 1))
					resp.WriteUint16(uint16(len(c.userBinaryParts[userBinaryPartID{session.charID, 3}])))
					resp.WriteBytes(make([]byte, 40))
					resp.WriteUint8(uint8(len(sessionStage) + 1))
					resp.WriteBytes(make([]byte, 8))
					resp.WriteNullTerminatedBytes(sessionName)
					resp.WriteBytes(c.userBinaryParts[userBinaryPartID{charID: session.charID, index: 3}])
					resp.WriteNullTerminatedBytes(sessionStage)
				}
			}
		}
	case 3: // Enumerate Party
		bf := byteframe.NewByteFrameFromBytes(pkt.MessageData)
		ip := bf.ReadBytes(4)
		ipString := fmt.Sprintf("%d.%d.%d.%d", ip[3], ip[2], ip[1], ip[0])
		port := bf.ReadUint16()
		bf.ReadUint16() // lenStage
		maxResults := bf.ReadUint16()
		bf.ReadBytes(1)
		stageID := stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
		for _, c := range s.server.Channels {
			if c.IP == ipString && c.Port == port {
				for _, stage := range c.stages {
					if stage.id == stageID {
						if count == maxResults {
							break
						}
						for session := range stage.clients {
							count++
							hrp := uint16(1)
							gr := uint16(0)
							s.server.db.QueryRow("SELECT hrp, gr FROM characters WHERE id=$1", session.charID).Scan(&hrp, &gr)
							sessionStage := stringsupport.UTF8ToSJIS(session.stageID)
							sessionName := stringsupport.UTF8ToSJIS(session.Name)
							resp.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(c.IP).To4()))
							resp.WriteUint16(c.Port)
							resp.WriteUint32(session.charID)
							resp.WriteUint8(uint8(len(sessionStage) + 1))
							resp.WriteUint8(uint8(len(sessionName) + 1))
							resp.WriteUint8(0)
							resp.WriteUint8(7) // lenBinary
							resp.WriteBytes(make([]byte, 48))
							resp.WriteNullTerminatedBytes(sessionStage)
							resp.WriteNullTerminatedBytes(sessionName)
							resp.WriteUint16(hrp)
							resp.WriteUint16(gr)
							resp.WriteBytes([]byte{0x06, 0x10, 0x00}) // Unk
						}
					}
				}
			}
		}
	case 4: // Find Party
		type FindPartyParams struct {
			StagePrefix     string
			RankRestriction uint16
			Targets         []uint16
			Unk0            []uint16
			Unk1            []uint16
			QuestID         []uint16
		}
		findPartyParams := FindPartyParams{
			StagePrefix: "sl2Ls210",
		}
		bf := byteframe.NewByteFrameFromBytes(pkt.MessageData)
		numParams := int(bf.ReadUint8())
		maxResults := bf.ReadUint16()
		for i := 0; i < numParams; i++ {
			switch bf.ReadUint8() {
			case 0:
				values := int(bf.ReadUint8())
				for i := 0; i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.RankRestriction = bf.ReadUint16()
					} else {
						findPartyParams.RankRestriction = uint16(bf.ReadInt8())
					}
				}
			case 1:
				values := int(bf.ReadUint8())
				for i := 0; i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.Targets = append(findPartyParams.Targets, bf.ReadUint16())
					} else {
						findPartyParams.Targets = append(findPartyParams.Targets, uint16(bf.ReadInt8()))
					}
				}
			case 2:
				values := int(bf.ReadUint8())
				for i := 0; i < values; i++ {
					var value int16
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						value = bf.ReadInt16()
					} else {
						value = int16(bf.ReadInt8())
					}
					switch value {
					case 0: // Public Bar
						findPartyParams.StagePrefix = "sl2Ls210"
					case 1: // Tokotoko Partnya
						findPartyParams.StagePrefix = "sl2Ls463"
					case 2: // Hunting Prowess Match
						findPartyParams.StagePrefix = "sl2Ls286"
					case 3: // Volpakkun Together
						findPartyParams.StagePrefix = "sl2Ls465"
					case 5: // Quick Party
						// Unk
					}
				}
			case 3: // Unknown
				values := int(bf.ReadUint8())
				for i := 0; i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.Unk0 = append(findPartyParams.Unk0, bf.ReadUint16())
					} else {
						findPartyParams.Unk0 = append(findPartyParams.Unk0, uint16(bf.ReadInt8()))
					}
				}
			case 4: // Looking for n or already have n
				values := int(bf.ReadUint8())
				for i := 0; i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.Unk1 = append(findPartyParams.Unk1, bf.ReadUint16())
					} else {
						findPartyParams.Unk1 = append(findPartyParams.Unk1, uint16(bf.ReadInt8()))
					}
				}
			case 5:
				values := int(bf.ReadUint8())
				for i := 0; i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.QuestID = append(findPartyParams.QuestID, bf.ReadUint16())
					} else {
						findPartyParams.QuestID = append(findPartyParams.QuestID, uint16(bf.ReadInt8()))
					}
				}
			}
		}
		for _, c := range s.server.Channels {
			for _, stage := range c.stages {
				if count == maxResults {
					break
				}
				if strings.HasPrefix(stage.id, findPartyParams.StagePrefix) {
					sb3 := byteframe.NewByteFrameFromBytes(stage.rawBinaryData[stageBinaryKey{1, 3}])
					sb3.Seek(4, 0)
					stageRankRestriction := sb3.ReadUint16()
					stageTarget := sb3.ReadUint16()
					if stageRankRestriction > findPartyParams.RankRestriction {
						continue
					}
					if len(findPartyParams.Targets) > 0 {
						for _, target := range findPartyParams.Targets {
							if target == stageTarget {
								break
							}
						}
						continue
					}
					count++
					sessionStage := stringsupport.UTF8ToSJIS(stage.id)
					resp.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(c.IP).To4()))
					resp.WriteUint16(c.Port)
					resp.WriteUint16(0) // Static?
					resp.WriteUint16(0) // Unk
					resp.WriteUint16(uint16(len(stage.clients)))
					resp.WriteUint16(stage.maxPlayers)
					resp.WriteUint16(0) // Num clients entered from stage
					resp.WriteUint16(stage.maxPlayers)
					resp.WriteUint8(1) // Static?
					resp.WriteUint8(uint8(len(sessionStage) + 1))
					resp.WriteUint8(uint8(len(stage.rawBinaryData[stageBinaryKey{1, 0}])))
					resp.WriteUint8(uint8(len(stage.rawBinaryData[stageBinaryKey{1, 1}])))
					resp.WriteUint16(stageRankRestriction)
					resp.WriteUint16(stageTarget)
					resp.WriteBytes(make([]byte, 12))
					resp.WriteNullTerminatedBytes(sessionStage)
					resp.WriteBytes(stage.rawBinaryData[stageBinaryKey{1, 0}])
					resp.WriteBytes(stage.rawBinaryData[stageBinaryKey{1, 1}])
				}
			}
		}
	}
	if (pkt.SearchType == 1 || pkt.SearchType == 3) && count == 0 {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	resp.Seek(0, io.SeekStart)
	resp.WriteUint16(count)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgCaExchangeItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfServerCommand(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAnnounce(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAnnounce)
	s.server.BroadcastRaviente(pkt.IPAddress, pkt.Port, pkt.StageID, pkt.Type)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfSetLoginwindow(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysTransBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCollectBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysGetState(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSerialize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysEnumlobby(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysEnumuser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysInfokyserver(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetCaUniqueID(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfTransferItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfTransferItem)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfEnumeratePrice(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumeratePrice)
	//resp := byteframe.NewByteFrame()
	//resp.WriteUint16(0) // Entry type 1 count
	//resp.WriteUint16(0) // Entry type 2 count
	// directly lifted for now because lacking it crashes the counter on having actual events present
	data, _ := hex.DecodeString("0000000066000003E800000000007300640100000320000000000006006401000003200000000000300064010000044C00000000007200640100000384000000000034006401000003840000000000140064010000051400000000006E006401000003E8000000000016006401000003E8000000000001006401000003200000000000430064010000057800000000006F006401000003840000000000330064010000044C00000000000B006401000003E800000000000F006401000006400000000000700064010000044C0000000000110064010000057800000000004C006401000003E8000000000059006401000006A400000000006D006401000005DC00000000004B006401000005DC000000000050006401000006400000000000350064010000070800000000006C0064010000044C000000000028006401000005DC00000000005300640100000640000000000060006401000005DC00000000005E0064010000051400000000007B006401000003E80000000000740064010000070800000000006B0064010000025800000000001B0064010000025800000000001C006401000002BC00000000001F006401000006A400000000007900640100000320000000000008006401000003E80000000000150064010000070800000000007A0064010000044C00000000000E00640100000640000000000055006401000007D0000000000002006401000005DC00000000002F0064010000064000000000002A0064010000076C00000000007E006401000002BC0000000000440064010000038400000000005C0064010000064000000000005B006401000006A400000000007D0064010000076C00000000007F006401000005DC0000000000540064010000064000000000002900640100000960000000000024006401000007D0000000000081006401000008340000000000800064010000038400000000001A006401000003E800000000002D0064010000038400000000004A006401000006A400000000005A00640100000384000000000027006401000007080000000000830064010000076C000000000040006401000006400000000000690064010000044C000000000025006401000004B000000000003100640100000708000000000082006401000003E800000000006500640100000640000000000051006401000007D000000000008C0064010000070800000000004D0064010000038400000000004E0064010000089800000000008B006401000004B000000000002E006401000009600000000000920064010000076C00000000008E00640100000514000000000068006401000004B000000000002B006401000003E800000000002C00640100000BB8000000000093006401000008FC00000000009000640100000AF0000000000094006401000006A400000000008D0064010000044C000000000052006401000005DC00000000004F006401000008980000000000970064010000070800000000006A0064010000064000000000005F00640100000384000000000026006401000008FC000000000096006401000007D00000000000980064010000076C000000000041006401000006A400000000003B006401000007080000000000360064010000083400000000009F00640100000A2800000000009A0064010000076C000000000021006401000007D000000000006300640100000A8C0000000000990064010000089800000000009E006401000007080000000000A100640100000C1C0000000000A200640100000C800000000000A400640100000DAC0000000000A600640100000C800000000000A50064010010")
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfEnumerateOrder(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateOrder)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func handleMsgMhfGetExtraInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfEnumerateUnionItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateUnionItem)
	var boxContents []byte
	bf := byteframe.NewByteFrame()
	err := s.server.db.QueryRow("SELECT item_box FROM users, characters WHERE characters.id = $1 AND users.id = characters.user_id", int(s.charID)).Scan(&boxContents)
	if err != nil {
		s.logger.Error("Failed to get shared item box contents from db", zap.Error(err))
		bf.WriteBytes(make([]byte, 4))
	} else {
		if len(boxContents) == 0 {
			bf.WriteBytes(make([]byte, 4))
		} else {
			amount := len(boxContents) / 4
			bf.WriteUint16(uint16(amount))
			bf.WriteUint32(0x00)
			bf.WriteUint16(0x00)
			for i := 0; i < amount; i++ {
				bf.WriteUint32(binary.BigEndian.Uint32(boxContents[i*4 : i*4+4]))
				if i+1 != amount {
					bf.WriteUint64(0x00)
				}
			}
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateUnionItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateUnionItem)
	// Get item cache from DB
	var boxContents []byte
	var oldItems []Item

	err := s.server.db.QueryRow("SELECT item_box FROM users, characters WHERE characters.id = $1 AND users.id = characters.user_id", int(s.charID)).Scan(&boxContents)
	if err != nil {
		s.logger.Error("Failed to get shared item box contents from db", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	} else {
		amount := len(boxContents) / 4
		oldItems = make([]Item, amount)
		for i := 0; i < amount; i++ {
			oldItems[i].ItemId = binary.BigEndian.Uint16(boxContents[i*4 : i*4+2])
			oldItems[i].Amount = binary.BigEndian.Uint16(boxContents[i*4+2 : i*4+4])
		}
	}

	// Update item stacks
	newItems := make([]Item, len(oldItems))
	copy(newItems, oldItems)
	for i := 0; i < int(pkt.Amount); i++ {
		for j := 0; j <= len(oldItems); j++ {
			if j == len(oldItems) {
				var newItem Item
				newItem.ItemId = pkt.Items[i].ItemId
				newItem.Amount = pkt.Items[i].Amount
				newItems = append(newItems, newItem)
				break
			}
			if pkt.Items[i].ItemId == oldItems[j].ItemId {
				newItems[j].Amount = pkt.Items[i].Amount
				break
			}
		}
	}

	// Delete empty item stacks
	for i := len(newItems) - 1; i >= 0; i-- {
		if int(newItems[i].Amount) == 0 {
			copy(newItems[i:], newItems[i+1:])
			newItems[len(newItems)-1] = make([]Item, 1)[0]
			newItems = newItems[:len(newItems)-1]
		}
	}

	// Create new item cache
	bf := byteframe.NewByteFrame()
	for i := 0; i < len(newItems); i++ {
		bf.WriteUint16(newItems[i].ItemId)
		bf.WriteUint16(newItems[i].Amount)
	}

	// Upload new item cache
	_, err = s.server.db.Exec("UPDATE users SET item_box = $1 FROM characters WHERE  users.id = characters.user_id AND characters.id = $2", bf.Data(), int(s.charID))
	if err != nil {
		s.logger.Error("Failed to update shared item box contents in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetCogInfo(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCheckWeeklyStamp(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckWeeklyStamp)
	weekCurrentStart := TimeWeekStart()
	weekNextStart := TimeWeekNext()
	var total, redeemed, updated uint16
	var nextClaim time.Time
	err := s.server.db.QueryRow(fmt.Sprintf("SELECT %s_next FROM stamps WHERE character_id=$1", pkt.StampType), s.charID).Scan(&nextClaim)
	if err != nil {
		s.server.db.Exec("INSERT INTO stamps (character_id, hl_next, ex_next) VALUES ($1, $2, $2)", s.charID, weekNextStart)
		nextClaim = weekNextStart
	}
	if nextClaim.Before(weekCurrentStart) {
		s.server.db.Exec(fmt.Sprintf("UPDATE stamps SET %s_total=%s_total+1, %s_next=$1 WHERE character_id=$2", pkt.StampType, pkt.StampType, pkt.StampType), weekNextStart, s.charID)
		updated = 1
	}
	s.server.db.QueryRow(fmt.Sprintf("SELECT %s_total, %s_redeemed FROM stamps WHERE character_id=$1", pkt.StampType, pkt.StampType), s.charID).Scan(&total, &redeemed)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(total)
	bf.WriteUint16(redeemed)
	bf.WriteUint16(updated)
	bf.WriteUint32(0) // Unk
	bf.WriteUint32(uint32(weekCurrentStart.Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfExchangeWeeklyStamp(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfExchangeWeeklyStamp)
	var total, redeemed uint16
	var tktStack mhfpacket.WarehouseStack
	if pkt.Unk1 == 0xA { // Yearly Sub Ex
		s.server.db.QueryRow("UPDATE stamps SET hl_total=hl_total-48, hl_redeemed=hl_redeemed-48 WHERE character_id=$1 RETURNING hl_total, hl_redeemed", s.charID).Scan(&total, &redeemed)
		tktStack = mhfpacket.WarehouseStack{ItemID: 0x08A2, Quantity: 1}
	} else {
		s.server.db.QueryRow(fmt.Sprintf("UPDATE stamps SET %s_redeemed=%s_redeemed+8 WHERE character_id=$1 RETURNING %s_total, %s_redeemed", pkt.StampType, pkt.StampType, pkt.StampType, pkt.StampType), s.charID).Scan(&total, &redeemed)
		if pkt.StampType == "hl" {
			tktStack = mhfpacket.WarehouseStack{ItemID: 0x065E, Quantity: 5}
		} else {
			tktStack = mhfpacket.WarehouseStack{ItemID: 0x065F, Quantity: 5}
		}
	}
	addWarehouseGift(s, "item", tktStack)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(total)
	bf.WriteUint16(redeemed)
	bf.WriteUint16(0)
	bf.WriteUint32(0) // Unk, but has possible values
	bf.WriteUint32(uint32(TimeWeekStart().Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func getGookData(s *Session, cid uint32) (uint16, []byte) {
	var data []byte
	var count uint16
	bf := byteframe.NewByteFrame()
	for i := 0; i < 5; i++ {
		err := s.server.db.QueryRow(fmt.Sprintf("SELECT gook%d FROM gook WHERE id=$1", i), cid).Scan(&data)
		if err != nil {
			s.server.db.Exec("INSERT INTO gook (id) VALUES ($1)", s.charID)
			return 0, bf.Data()
		}
		if err == nil && data != nil {
			count++
			if s.charID == cid && count == 1 {
				gook := byteframe.NewByteFrameFromBytes(data)
				bf.WriteBytes(gook.ReadBytes(4))
				d := gook.ReadBytes(2)
				bf.WriteBytes(d)
				bf.WriteBytes(d)
				bf.WriteBytes(gook.DataFromCurrent())
			} else {
				bf.WriteBytes(data)
			}
		}
	}
	return count, bf.Data()
}

func handleMsgMhfEnumerateGuacot(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuacot)
	bf := byteframe.NewByteFrame()
	count, data := getGookData(s, s.charID)
	bf.WriteUint16(count)
	bf.WriteBytes(data)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateGuacot(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuacot)
	for _, gook := range pkt.Gooks {
		if !gook.Exists {
			s.server.db.Exec(fmt.Sprintf("UPDATE gook SET gook%d=NULL WHERE id=$1", gook.Index), s.charID)
		} else {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(gook.Index)
			bf.WriteUint16(gook.Type)
			bf.WriteBytes(gook.Data)
			bf.WriteUint8(gook.NameLen)
			bf.WriteBytes(gook.Name)
			s.server.db.Exec(fmt.Sprintf("UPDATE gook SET gook%d=$1 WHERE id=$2", gook.Index), bf.Data(), s.charID)
			dumpSaveData(s, bf.Data(), fmt.Sprintf("goocoo-%d", gook.Index))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfInfoScenarioCounter(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoScenarioCounter)
	scenarioCounter := []struct {
		MainID uint32
		Unk1   uint8 // Bool item exchange?
		// 0 = basic, 1 = veteran, 3 = other, 6 = pallone, 7 = diva
		CategoryID uint8
	}{
		//000000110000
		{
			MainID: 0x00000011, Unk1: 0, CategoryID: 0,
		},
		// 0000005D0001
		{
			MainID: 0x0000005D, Unk1: 0, CategoryID: 1,
		},
		// 0000005C0001
		{
			MainID: 0x0000005C, Unk1: 0, CategoryID: 1,
		},
		// 000000510001
		{
			MainID: 0x00000051, Unk1: 0, CategoryID: 1,
		},
		// 0000005B0001
		{
			MainID: 0x0000005B, Unk1: 0, CategoryID: 1,
		},
		// 0000005A0001
		{
			MainID: 0x0000005A, Unk1: 0, CategoryID: 1,
		},
		// 000000590001
		{
			MainID: 0x00000059, Unk1: 0, CategoryID: 1,
		},
		// 000000580001
		{
			MainID: 0x00000058, Unk1: 0, CategoryID: 1,
		},
		// 000000570001
		{
			MainID: 0x00000057, Unk1: 0, CategoryID: 1,
		},
		// 000000560001
		{
			MainID: 0x00000056, Unk1: 0, CategoryID: 1,
		},
		// 000000550001
		{
			MainID: 0x00000055, Unk1: 0, CategoryID: 1,
		},
		// 000000540001
		{
			MainID: 0x00000054, Unk1: 0, CategoryID: 1,
		},
		// 000000530001
		{
			MainID: 0x00000053, Unk1: 0, CategoryID: 1,
		},
		// 000000520001
		{
			MainID: 0x00000052, Unk1: 0, CategoryID: 1,
		},
		// 000000570103
		{
			MainID: 0x00000057, Unk1: 1, CategoryID: 3,
		},
		// 000000580103
		{
			MainID: 0x00000058, Unk1: 1, CategoryID: 3,
		},
		// 000000590103
		{
			MainID: 0x00000059, Unk1: 1, CategoryID: 3,
		},
		// 0000005A0103
		{
			MainID: 0x0000005A, Unk1: 1, CategoryID: 3,
		},
		// 0000005B0103
		{
			MainID: 0x0000005B, Unk1: 1, CategoryID: 3,
		},
		// 0000005C0103
		{
			MainID: 0x0000005C, Unk1: 1, CategoryID: 3,
		},
		// 000000530103
		{
			MainID: 0x00000053, Unk1: 1, CategoryID: 3,
		},
		// 000000560103
		{
			MainID: 0x00000056, Unk1: 1, CategoryID: 3,
		},
		// 0000003C0103
		{
			MainID: 0x0000003C, Unk1: 1, CategoryID: 3,
		},
		// 0000003A0103
		{
			MainID: 0x0000003A, Unk1: 1, CategoryID: 3,
		},
		// 0000003B0103
		{
			MainID: 0x0000003B, Unk1: 1, CategoryID: 3,
		},
		// 0000001B0103
		{
			MainID: 0x0000001B, Unk1: 1, CategoryID: 3,
		},
		// 000000190103
		{
			MainID: 0x00000019, Unk1: 1, CategoryID: 3,
		},
		// 0000001A0103
		{
			MainID: 0x0000001A, Unk1: 1, CategoryID: 3,
		},
		// 000000170103
		{
			MainID: 0x00000017, Unk1: 1, CategoryID: 3,
		},
		// 000000020103
		{
			MainID: 0x00000002, Unk1: 1, CategoryID: 3,
		},
		// 000000030103
		{
			MainID: 0x00000003, Unk1: 1, CategoryID: 3,
		},
		// 000000040103
		{
			MainID: 0x00000004, Unk1: 1, CategoryID: 3,
		},
		// 0000001F0103
		{
			MainID: 0x0000001F, Unk1: 1, CategoryID: 3,
		},
		// 000000200103
		{
			MainID: 0x00000020, Unk1: 1, CategoryID: 3,
		},
		// 000000210103
		{
			MainID: 0x00000021, Unk1: 1, CategoryID: 3,
		},
		// 000000220103
		{
			MainID: 0x00000022, Unk1: 1, CategoryID: 3,
		},
		// 000000230103
		{
			MainID: 0x00000023, Unk1: 1, CategoryID: 3,
		},
		// 000000240103
		{
			MainID: 0x00000024, Unk1: 1, CategoryID: 3,
		},
		// 000000250103
		{
			MainID: 0x00000025, Unk1: 1, CategoryID: 3,
		},
		// 000000280103
		{
			MainID: 0x00000028, Unk1: 1, CategoryID: 3,
		},
		// 000000260103
		{
			MainID: 0x00000026, Unk1: 1, CategoryID: 3,
		},
		// 000000270103
		{
			MainID: 0x00000027, Unk1: 1, CategoryID: 3,
		},
		// 000000300103
		{
			MainID: 0x00000030, Unk1: 1, CategoryID: 3,
		},
		// 0000000C0103
		{
			MainID: 0x0000000C, Unk1: 1, CategoryID: 3,
		},
		// 0000000D0103
		{
			MainID: 0x0000000D, Unk1: 1, CategoryID: 3,
		},
		// 0000001E0103
		{
			MainID: 0x0000001E, Unk1: 1, CategoryID: 3,
		},
		// 0000001D0103
		{
			MainID: 0x0000001D, Unk1: 1, CategoryID: 3,
		},
		// 0000002E0003
		{
			MainID: 0x0000002E, Unk1: 0, CategoryID: 3,
		},
		// 000000000004
		{
			MainID: 0x00000000, Unk1: 0, CategoryID: 4,
		},
		// 000000010004
		{
			MainID: 0x00000001, Unk1: 0, CategoryID: 4,
		},
		// 000000020004
		{
			MainID: 0x00000002, Unk1: 0, CategoryID: 4,
		},
		// 000000030004
		{
			MainID: 0x00000003, Unk1: 0, CategoryID: 4,
		},
		// 000000040004
		{
			MainID: 0x00000004, Unk1: 0, CategoryID: 4,
		},
		// 000000050004
		{
			MainID: 0x00000005, Unk1: 0, CategoryID: 4,
		},
		// 000000060004
		{
			MainID: 0x00000006, Unk1: 0, CategoryID: 4,
		},
		// 000000070004
		{
			MainID: 0x00000007, Unk1: 0, CategoryID: 4,
		},
		// 000000080004
		{
			MainID: 0x00000008, Unk1: 0, CategoryID: 4,
		},
		// 000000090004
		{
			MainID: 0x00000009, Unk1: 0, CategoryID: 4,
		},
		// 0000000A0004
		{
			MainID: 0x0000000A, Unk1: 0, CategoryID: 4,
		},
		// 0000000B0004
		{
			MainID: 0x0000000B, Unk1: 0, CategoryID: 4,
		},
		// 0000000C0004
		{
			MainID: 0x0000000C, Unk1: 0, CategoryID: 4,
		},
		// 0000000D0004
		{
			MainID: 0x0000000D, Unk1: 0, CategoryID: 4,
		},
		// 0000000E0004
		{
			MainID: 0x0000000E, Unk1: 0, CategoryID: 4,
		},
		// 000000320005
		{
			MainID: 0x00000032, Unk1: 0, CategoryID: 5,
		},
		// 000000330005
		{
			MainID: 0x00000033, Unk1: 0, CategoryID: 5,
		},
		// 000000340005
		{
			MainID: 0x00000034, Unk1: 0, CategoryID: 5,
		},
		// 000000350005
		{
			MainID: 0x00000035, Unk1: 0, CategoryID: 5,
		},
		// 000000360005
		{
			MainID: 0x00000036, Unk1: 0, CategoryID: 5,
		},
		// 000000370005
		{
			MainID: 0x00000037, Unk1: 0, CategoryID: 5,
		},
		// 000000380005
		{
			MainID: 0x00000038, Unk1: 0, CategoryID: 5,
		},
		// 0000003A0005
		{
			MainID: 0x0000003A, Unk1: 0, CategoryID: 5,
		},
		// 0000003F0005
		{
			MainID: 0x0000003F, Unk1: 0, CategoryID: 5,
		},
		// 000000400005
		{
			MainID: 0x00000040, Unk1: 0, CategoryID: 5,
		},
		// 000000410005
		{
			MainID: 0x00000041, Unk1: 0, CategoryID: 5,
		},
		// 000000430005
		{
			MainID: 0x00000043, Unk1: 0, CategoryID: 5,
		},
		// 000000470005
		{
			MainID: 0x00000047, Unk1: 0, CategoryID: 5,
		},
		// 0000004B0005
		{
			MainID: 0x0000004B, Unk1: 0, CategoryID: 5,
		},
		// 0000003D0005
		{
			MainID: 0x0000003D, Unk1: 0, CategoryID: 5,
		},
		// 000000440005
		{
			MainID: 0x00000044, Unk1: 0, CategoryID: 5,
		},
		// 000000420005
		{
			MainID: 0x00000042, Unk1: 0, CategoryID: 5,
		},
		// 0000004C0005
		{
			MainID: 0x0000004C, Unk1: 0, CategoryID: 5,
		},
		// 000000460005
		{
			MainID: 0x00000046, Unk1: 0, CategoryID: 5,
		},
		// 0000004D0005
		{
			MainID: 0x0000004D, Unk1: 0, CategoryID: 5,
		},
		// 000000480005
		{
			MainID: 0x00000048, Unk1: 0, CategoryID: 5,
		},
		// 0000004A0005
		{
			MainID: 0x0000004A, Unk1: 0, CategoryID: 5,
		},
		// 000000490005
		{
			MainID: 0x00000049, Unk1: 0, CategoryID: 5,
		},
		// 0000004E0005
		{
			MainID: 0x0000004E, Unk1: 0, CategoryID: 5,
		},
		// 000000450005
		{
			MainID: 0x00000045, Unk1: 0, CategoryID: 5,
		},
		// 0000003E0005
		{
			MainID: 0x0000003E, Unk1: 0, CategoryID: 5,
		},
		// 0000004F0005
		{
			MainID: 0x0000004F, Unk1: 0, CategoryID: 5,
		},
		// 000000000106
		{
			MainID: 0x00000000, Unk1: 1, CategoryID: 6,
		},
		// 000000010106
		{
			MainID: 0x00000001, Unk1: 1, CategoryID: 6,
		},
		// 000000020106
		{
			MainID: 0x00000002, Unk1: 1, CategoryID: 6,
		},
		// 000000030106
		{
			MainID: 0x00000003, Unk1: 1, CategoryID: 6,
		},
		// 000000040106
		{
			MainID: 0x00000004, Unk1: 1, CategoryID: 6,
		},
		// 000000050106
		{
			MainID: 0x00000005, Unk1: 1, CategoryID: 6,
		},
		// 000000060106
		{
			MainID: 0x00000006, Unk1: 1, CategoryID: 6,
		},
		// 000000070106
		{
			MainID: 0x00000007, Unk1: 1, CategoryID: 6,
		},
		// 000000080106
		{
			MainID: 0x00000008, Unk1: 1, CategoryID: 6,
		},
		// 000000090106
		{
			MainID: 0x00000009, Unk1: 1, CategoryID: 6,
		},
		// 000000110106
		{
			MainID: 0x00000011, Unk1: 1, CategoryID: 6,
		},
		// 0000000A0106
		{
			MainID: 0x0000000A, Unk1: 1, CategoryID: 6,
		},
		// 0000000B0106
		{
			MainID: 0x0000000B, Unk1: 1, CategoryID: 6,
		},
		// 0000000C0106
		{
			MainID: 0x0000000C, Unk1: 1, CategoryID: 6,
		},
		// 0000000D0106
		{
			MainID: 0x0000000D, Unk1: 1, CategoryID: 6,
		},
		// 0000000E0106
		{
			MainID: 0x0000000E, Unk1: 1, CategoryID: 6,
		},
		// 0000000F0106
		{
			MainID: 0x0000000F, Unk1: 1, CategoryID: 6,
		},
		// 000000100106
		{
			MainID: 0x00000010, Unk1: 1, CategoryID: 6,
		},
		// 000000320107
		{
			MainID: 0x00000032, Unk1: 1, CategoryID: 7,
		},
		// 000000350107
		{
			MainID: 0x00000035, Unk1: 1, CategoryID: 7,
		},
		// 0000003E0107
		{
			MainID: 0x0000003E, Unk1: 1, CategoryID: 7,
		},
		// 000000340107
		{
			MainID: 0x00000034, Unk1: 1, CategoryID: 7,
		},
		// 000000380107
		{
			MainID: 0x00000038, Unk1: 1, CategoryID: 7,
		},
		// 000000330107
		{
			MainID: 0x00000033, Unk1: 1, CategoryID: 7,
		},
		// 000000310107
		{
			MainID: 0x00000031, Unk1: 1, CategoryID: 7,
		},
		// 000000360107
		{
			MainID: 0x00000036, Unk1: 1, CategoryID: 7,
		},
		// 000000390107
		{
			MainID: 0x00000039, Unk1: 1, CategoryID: 7,
		},
		// 000000370107
		{
			MainID: 0x00000037, Unk1: 1, CategoryID: 7,
		},
		// 0000003D0107
		{
			MainID: 0x0000003D, Unk1: 1, CategoryID: 7,
		},
		// 0000003A0107
		{
			MainID: 0x0000003A, Unk1: 1, CategoryID: 7,
		},
		// 0000003C0107
		{
			MainID: 0x0000003C, Unk1: 1, CategoryID: 7,
		},
		// 0000003B0107
		{
			MainID: 0x0000003B, Unk1: 1, CategoryID: 7,
		},
		// 0000002A0107
		{
			MainID: 0x0000002A, Unk1: 1, CategoryID: 7,
		},
		// 000000300107
		{
			MainID: 0x00000030, Unk1: 1, CategoryID: 7,
		},
		// 000000280107
		{
			MainID: 0x00000028, Unk1: 1, CategoryID: 7,
		},
		// 000000270107
		{
			MainID: 0x00000027, Unk1: 1, CategoryID: 7,
		},
		// 0000002B0107
		{
			MainID: 0x0000002B, Unk1: 1, CategoryID: 7,
		},
		// 0000002E0107
		{
			MainID: 0x0000002E, Unk1: 1, CategoryID: 7,
		},
		// 000000290107
		{
			MainID: 0x00000029, Unk1: 1, CategoryID: 7,
		},
		// 0000002C0107
		{
			MainID: 0x0000002C, Unk1: 1, CategoryID: 7,
		},
		// 0000002D0107
		{
			MainID: 0x0000002D, Unk1: 1, CategoryID: 7,
		},
		// 0000002F0107
		{
			MainID: 0x0000002F, Unk1: 1, CategoryID: 7,
		},
		// 000000250107
		{
			MainID: 0x00000025, Unk1: 1, CategoryID: 7,
		},
		// 000000220107
		{
			MainID: 0x00000022, Unk1: 1, CategoryID: 7,
		},
		// 000000210107
		{
			MainID: 0x00000021, Unk1: 1, CategoryID: 7,
		},
		// 000000200107
		{
			MainID: 0x00000020, Unk1: 1, CategoryID: 7,
		},
		// 0000001C0107
		{
			MainID: 0x0000001C, Unk1: 1, CategoryID: 7,
		},
		// 0000001A0107
		{
			MainID: 0x0000001A, Unk1: 1, CategoryID: 7,
		},
		// 000000240107
		{
			MainID: 0x00000024, Unk1: 1, CategoryID: 7,
		},
		// 000000260107
		{
			MainID: 0x00000026, Unk1: 1, CategoryID: 7,
		},
		// 000000230107
		{
			MainID: 0x00000023, Unk1: 1, CategoryID: 7,
		},
		// 0000001B0107
		{
			MainID: 0x0000001B, Unk1: 1, CategoryID: 7,
		},
		// 0000001E0107
		{
			MainID: 0x0000001E, Unk1: 1, CategoryID: 7,
		},
		// 0000001F0107
		{
			MainID: 0x0000001F, Unk1: 1, CategoryID: 7,
		},
		// 0000001D0107
		{
			MainID: 0x0000001D, Unk1: 1, CategoryID: 7,
		},
		// 000000180107
		{
			MainID: 0x00000018, Unk1: 1, CategoryID: 7,
		},
		// 000000170107
		{
			MainID: 0x00000017, Unk1: 1, CategoryID: 7,
		},
		// 000000160107
		{
			MainID: 0x00000016, Unk1: 1, CategoryID: 7,
		},
		// 000000150107
		// Missing file
		// {
		// 	MainID: 0x00000015, Unk1: 1, CategoryID: 7,
		// },
		// 000000190107
		{
			MainID: 0x00000019, Unk1: 1, CategoryID: 7,
		},
		// 000000140107
		// Missing file
		// {
		// 	MainID: 0x00000014, Unk1: 1, CategoryID: 7,
		// },
		// 000000070107
		// Missing file
		// {
		//	MainID: 0x00000007, Unk1: 1, CategoryID: 7,
		// },
		// 000000090107
		// Missing file
		// {
		//	MainID: 0x00000009, Unk1: 1, CategoryID: 7,
		// },
		// 0000000D0107
		// Missing file
		// {
		//	MainID: 0x0000000D, Unk1: 1, CategoryID: 7,
		// },
		// 000000100107
		// Missing file
		// {
		//	MainID: 0x00000010, Unk1: 1, CategoryID: 7,
		// },
		// 0000000C0107
		// Missing file
		// {
		//	MainID: 0x0000000C, Unk1: 1, CategoryID: 7,
		// },
		// 0000000E0107
		// Missing file
		// {
		//	MainID: 0x0000000E, Unk1: 1, CategoryID: 7,
		// },
		// 0000000F0107
		// Missing file
		// {
		//	MainID: 0x0000000F, Unk1: 1, CategoryID: 7,
		// },
		// 000000130107
		// Missing file
		// {
		//	MainID: 0x00000013, Unk1: 1, CategoryID: 7,
		// },
		// 0000000A0107
		// Missing file
		// {
		//	MainID: 0x0000000A, Unk1: 1, CategoryID: 7,
		// },
		// 000000080107
		// Missing file
		// {
		//	MainID: 0x00000008, Unk1: 1, CategoryID: 7,
		// },
		// 0000000B0107
		// Missing file
		// {
		//	MainID: 0x0000000B, Unk1: 1, CategoryID: 7,
		// },
		// 000000120107
		// Missing file
		// {
		//	MainID: 0x00000012, Unk1: 1, CategoryID: 7,
		// },
		// 000000110107
		// Missing file
		// {
		// 	MainID: 0x00000011, Unk1: 1, CategoryID: 7,
		// },
		// 000000060107
		// Missing file
		// {
		// 	MainID: 0x00000006, Unk1: 1, CategoryID: 7,
		// },
		// 000000050107
		// Missing file
		// {
		// 	MainID: 0x00000005, Unk1: 1, CategoryID: 7,
		// },
		// 000000040107
		// Missing file
		// {
		//	MainID: 0x00000004, Unk1: 1, CategoryID: 7,
		// },
		// 000000030107
		{
			MainID: 0x00000003, Unk1: 1, CategoryID: 7,
		},
		// 000000020107
		{
			MainID: 0x00000002, Unk1: 1, CategoryID: 7,
		},
		// 000000010107
		{
			MainID: 0x00000001, Unk1: 1, CategoryID: 7,
		},
		// 000000000107
		{
			MainID: 0x00000000, Unk1: 1, CategoryID: 7,
		},
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(scenarioCounter))) // Entry count
	for _, entry := range scenarioCounter {
		resp.WriteUint32(entry.MainID)
		resp.WriteUint8(entry.Unk1)
		resp.WriteUint8(entry.CategoryID)
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetEtcPoints(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEtcPoints)

	var dailyTime time.Time
	_ = s.server.db.QueryRow("SELECT COALESCE(daily_time, $2) FROM characters WHERE id = $1", s.charID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).Scan(&dailyTime)
	if TimeAdjusted().After(dailyTime) {
		s.server.db.Exec("UPDATE characters SET bonus_quests = 0, daily_quests = 0 WHERE id=$1", s.charID)
	}

	var bonusQuests, dailyQuests, promoPoints uint32
	_ = s.server.db.QueryRow(`SELECT bonus_quests, daily_quests, promo_points FROM characters WHERE id = $1`, s.charID).Scan(&bonusQuests, &dailyQuests, &promoPoints)
	resp := byteframe.NewByteFrame()
	resp.WriteUint8(3) // Maybe a count of uint32(s)?
	resp.WriteUint32(bonusQuests)
	resp.WriteUint32(dailyQuests)
	resp.WriteUint32(promoPoints)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUpdateEtcPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateEtcPoint)

	var column string
	switch pkt.PointType {
	case 0:
		column = "bonus_quests"
	case 1:
		column = "daily_quests"
	case 2:
		column = "promo_points"
	}

	var value int
	err := s.server.db.QueryRow(fmt.Sprintf(`SELECT %s FROM characters WHERE id = $1`, column), s.charID).Scan(&value)
	if err == nil {
		if value-int(pkt.Delta) < 0 {
			s.server.db.Exec(fmt.Sprintf(`UPDATE characters SET %s = 0 WHERE id = $1`, column), s.charID)
		} else {
			s.server.db.Exec(fmt.Sprintf(`UPDATE characters SET %s = %s + $1 WHERE id = $2`, column, column), pkt.Delta, s.charID)
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfStampcardStamp(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStampcardStamp)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(pkt.HR)
	bf.WriteUint16(pkt.GR)
	var stamps uint16
	_ = s.server.db.QueryRow(`SELECT stampcard FROM characters WHERE id = $1`, s.charID).Scan(&stamps)
	bf.WriteUint16(stamps)
	stamps += pkt.Stamps
	bf.WriteUint16(stamps)
	s.server.db.Exec(`UPDATE characters SET stampcard = $1 WHERE id = $2`, stamps, s.charID)
	if stamps%30 == 0 {
		bf.WriteUint16(2)
		bf.WriteUint16(pkt.Reward2)
		bf.WriteUint16(pkt.Item2)
		bf.WriteUint16(pkt.Quantity2)
		addWarehouseGift(s, "item", mhfpacket.WarehouseStack{ItemID: pkt.Item2, Quantity: pkt.Quantity2})
	} else if stamps%15 == 0 {
		bf.WriteUint16(1)
		bf.WriteUint16(pkt.Reward1)
		bf.WriteUint16(pkt.Item1)
		bf.WriteUint16(pkt.Quantity1)
		addWarehouseGift(s, "item", mhfpacket.WarehouseStack{ItemID: pkt.Item1, Quantity: pkt.Quantity1})
	} else {
		bf.WriteBytes(make([]byte, 8))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfStampcardPrize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfUnreserveSrg(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfKickExportForce(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEarthStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEarthStatus)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(TimeWeekStart().Add(time.Hour * -24).Unix())) // Start
	bf.WriteUint32(uint32(TimeWeekNext().Add(time.Hour * 24).Unix()))   // End
	bf.WriteInt32(s.server.erupeConfig.DevModeOptions.EarthStatusOverride)
	bf.WriteInt32(s.server.erupeConfig.DevModeOptions.EarthIDOverride)
	bf.WriteInt32(s.server.erupeConfig.DevModeOptions.EarthMonsterOverride)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfRegistSpabiTime(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEarthValue(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEarthValue)
	type EarthValues struct {
		Value []uint32
	}

	var earthValues []EarthValues
	switch pkt.ReqType {
	case 1:
		earthValues = []EarthValues{
			{[]uint32{1, 312, 0, 0, 0, 0}},
			{[]uint32{2, 99, 0, 0, 0, 0}},
		}
	case 2:
		earthValues = []EarthValues{
			{[]uint32{1, 5771, 0, 0, 0, 0}},
			{[]uint32{2, 1847, 0, 0, 0, 0}},
		}
	case 3:
		earthValues = []EarthValues{
			{[]uint32{1001, 36, 0, 0, 0, 0}},
			{[]uint32{9001, 3, 0, 0, 0, 0}},
			{[]uint32{9002, 10, 300, 0, 0, 0}},
		}
	}

	var data []*byteframe.ByteFrame
	for _, i := range earthValues {
		bf := byteframe.NewByteFrame()
		for _, j := range i.Value {
			bf.WriteUint32(j)
		}
		data = append(data, bf)
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfDebugPostValue(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetRandFromTable(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetSenyuDailyCount(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetSenyuDailyCount)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

type SeibattleTimetable struct {
	Start time.Time
	End   time.Time
}

type SeibattleKeyScore struct {
	Unk0 uint8
	Unk1 int32
}

type SeibattleCareer struct {
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
}

type SeibattleOpponent struct {
	Unk0 int32
	Unk1 int8
}

type SeibattleConventionResult struct {
	Unk0 uint32
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
	Unk4 uint16
}

type SeibattleCharScore struct {
	Unk0 uint32
}

type SeibattleCurResult struct {
	Unk0 uint32
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
}

type Seibattle struct {
	Timetable        []SeibattleTimetable
	KeyScore         []SeibattleKeyScore
	Career           []SeibattleCareer
	Opponent         []SeibattleOpponent
	ConventionResult []SeibattleConventionResult
	CharScore        []SeibattleCharScore
	CurResult        []SeibattleCurResult
}

func handleMsgMhfGetSeibattle(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetSeibattle)
	var data []*byteframe.ByteFrame
	seibattle := Seibattle{
		Timetable: []SeibattleTimetable{
			{TimeMidnight(), TimeMidnight().Add(time.Hour * 8)},
			{TimeMidnight().Add(time.Hour * 8), TimeMidnight().Add(time.Hour * 16)},
			{TimeMidnight().Add(time.Hour * 16), TimeMidnight().Add(time.Hour * 24)},
		},
		KeyScore: []SeibattleKeyScore{
			{0, 0},
		},
		Career: []SeibattleCareer{
			{0, 0, 0},
		},
		Opponent: []SeibattleOpponent{
			{1, 1},
		},
		ConventionResult: []SeibattleConventionResult{
			{0, 0, 0, 0, 0},
		},
		CharScore: []SeibattleCharScore{
			{0},
		},
		CurResult: []SeibattleCurResult{
			{0, 0, 0, 0},
		},
	}

	switch pkt.Type {
	case 1:
		for _, timetable := range seibattle.Timetable {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(uint32(timetable.Start.Unix()))
			bf.WriteUint32(uint32(timetable.End.Unix()))
			data = append(data, bf)
		}
	case 3: // Key score?
		for _, keyScore := range seibattle.KeyScore {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(keyScore.Unk0)
			bf.WriteInt32(keyScore.Unk1)
			data = append(data, bf)
		}
	case 4: // Career?
		for _, career := range seibattle.Career {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(career.Unk0)
			bf.WriteUint16(career.Unk1)
			bf.WriteUint16(career.Unk2)
			data = append(data, bf)
		}
	case 5: // Opponent?
		for _, opponent := range seibattle.Opponent {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(opponent.Unk0)
			bf.WriteInt8(opponent.Unk1)
			data = append(data, bf)
		}
	case 6: // Convention result?
		for _, conventionResult := range seibattle.ConventionResult {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(conventionResult.Unk0)
			bf.WriteUint16(conventionResult.Unk1)
			bf.WriteUint16(conventionResult.Unk2)
			bf.WriteUint16(conventionResult.Unk3)
			bf.WriteUint16(conventionResult.Unk4)
			data = append(data, bf)
		}
	case 7: // Char score?
		for _, charScore := range seibattle.CharScore {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(charScore.Unk0)
			data = append(data, bf)
		}
	case 8: // Cur result?
		for _, curResult := range seibattle.CurResult {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(curResult.Unk0)
			bf.WriteUint16(curResult.Unk1)
			bf.WriteUint16(curResult.Unk2)
			bf.WriteUint16(curResult.Unk3)
			data = append(data, bf)
		}
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostSeibattle(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetDailyMissionMaster(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetDailyMissionPersonal(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetDailyMissionPersonal(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEquipSkinHist)
	size := 3200
	if _config.ErupeConfig.RealClientMode <= _config.Z2 {
		size = 2560
	}
	if _config.ErupeConfig.RealClientMode <= _config.Z1 {
		size = 1280
	}

	var data []byte
	err := s.server.db.QueryRow("SELECT COALESCE(skin_hist::bytea, $2::bytea) FROM characters WHERE id = $1", s.charID, make([]byte, size)).Scan(&data)
	if err != nil {
		s.logger.Error("Failed to load skin_hist", zap.Error(err))
		data = make([]byte, size)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfUpdateEquipSkinHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateEquipSkinHist)
	size := 3200
	if _config.ErupeConfig.RealClientMode <= _config.Z2 {
		size = 2560
	}
	if _config.ErupeConfig.RealClientMode <= _config.Z1 {
		size = 1280
	}

	var data []byte
	err := s.server.db.QueryRow("SELECT COALESCE(skin_hist, $2) FROM characters WHERE id = $1", s.charID, make([]byte, size)).Scan(&data)
	if err != nil {
		s.logger.Error("Failed to get skin_hist", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	bit := int(pkt.ArmourID) - 10000
	startByte := (size / 5) * int(pkt.MogType)
	// psql set_bit could also work but I couldn't get it working
	byteInd := bit / 8
	bitInByte := bit % 8
	data[startByte+byteInd] |= bits.Reverse8(1 << uint(bitInByte))
	dumpSaveData(s, data, "skinhist")
	s.server.db.Exec("UPDATE characters SET skin_hist=$1 WHERE id=$2", data, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetUdShopCoin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdShopCoin)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUseUdShopCoin(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetEnhancedMinidata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetEnhancedMinidata)
	// this looks to be the detailed chunk of information you can pull up on players in town
	var data []byte
	err := s.server.db.QueryRow("SELECT minidata FROM characters WHERE id = $1", pkt.CharID).Scan(&data)
	if err != nil {
		s.logger.Error("Failed to load minidata")
		data = make([]byte, 1)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSetEnhancedMinidata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetEnhancedMinidata)
	dumpSaveData(s, pkt.RawDataPayload, "minidata")
	_, err := s.server.db.Exec("UPDATE characters SET minidata=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Error("Failed to save minidata", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetLobbyCrowd(s *Session, p mhfpacket.MHFPacket) {
	// this requests a specific server's population but seems to have been
	// broken at some point on live as every example response across multiple
	// servers sends back the exact same information?
	// It can be worried about later if we ever get to the point where there are
	// full servers to actually need to migrate people from and empty ones to
	pkt := p.(*mhfpacket.MsgMhfGetLobbyCrowd)
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 0x320))
}

type TrendWeapon struct {
	WeaponType uint8
	WeaponID   uint16
}

func handleMsgMhfGetTrendWeapon(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTrendWeapon)
	trendWeapons := [14][3]TrendWeapon{}
	for i := uint8(0); i < 14; i++ {
		rows, err := s.server.db.Query(`SELECT weapon_id FROM trend_weapons WHERE weapon_type=$1 ORDER BY count DESC LIMIT 3`, i)
		if err != nil {
			continue
		}
		j := 0
		for rows.Next() {
			trendWeapons[i][j].WeaponType = i
			rows.Scan(&trendWeapons[i][j].WeaponID)
			j++
		}
	}

	x := uint8(0)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	for _, weaponType := range trendWeapons {
		for _, weapon := range weaponType {
			bf.WriteUint8(weapon.WeaponType)
			bf.WriteUint16(weapon.WeaponID)
			x++
		}
	}
	bf.Seek(0, 0)
	bf.WriteUint8(x)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateUseTrendWeaponLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateUseTrendWeaponLog)
	s.server.db.Exec(`INSERT INTO trend_weapons (weapon_id, weapon_type, count) VALUES ($1, $2, 1) ON CONFLICT (weapon_id) DO
		UPDATE SET count = trend_weapons.count+1`, pkt.WeaponID, pkt.WeaponType)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
