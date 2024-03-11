package signserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"erupe-ce/server/channelserver"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"time"
)

func (s *Session) makeSignResponse(uid uint32) []byte {
	// Get the characters from the DB.
	chars, err := s.server.getCharactersForUser(uid)
	if len(chars) == 0 && uid != 0 {
		err = s.server.newUserChara(uid)
		if err == nil {
			chars, err = s.server.getCharactersForUser(uid)
		}
	}
	if err != nil {
		s.logger.Warn("Error getting characters from DB", zap.Error(err))
	}

	bf := byteframe.NewByteFrame()
	var tokenID uint32
	var sessToken string
	if uid == 0 && s.psn != "" {
		tokenID, sessToken, err = s.server.registerPsnToken(s.psn)
	} else {
		tokenID, sessToken, err = s.server.registerUidToken(uid)
	}
	if err != nil {
		bf.WriteUint8(uint8(SIGN_EABORT))
		return bf.Data()
	}

	if s.client == PS3 && (s.server.erupeConfig.PatchServerFile == "" || s.server.erupeConfig.PatchServerManifest == "") {
		bf.WriteUint8(uint8(SIGN_EABORT))
		return bf.Data()
	}

	bf.WriteUint8(uint8(SIGN_SUCCESS))
	bf.WriteUint8(2) // patch server count
	bf.WriteUint8(1) // entrance server count
	bf.WriteUint8(uint8(len(chars)))
	bf.WriteUint32(tokenID)
	bf.WriteBytes([]byte(sessToken))
	bf.WriteUint32(uint32(channelserver.TimeAdjusted().Unix()))
	if s.client == PS3 {
		ps.Uint8(bf, fmt.Sprintf("%s/ps3", s.server.erupeConfig.PatchServerManifest), false)
		ps.Uint8(bf, fmt.Sprintf("%s/ps3", s.server.erupeConfig.PatchServerFile), false)
	} else {
		ps.Uint8(bf, s.server.erupeConfig.PatchServerManifest, false)
		ps.Uint8(bf, s.server.erupeConfig.PatchServerFile, false)
	}
	if strings.Split(s.rawConn.RemoteAddr().String(), ":")[0] == "127.0.0.1" {
		ps.Uint8(bf, fmt.Sprintf("127.0.0.1:%d", s.server.erupeConfig.Entrance.Port), false)
	} else {
		ps.Uint8(bf, fmt.Sprintf("%s:%d", s.server.erupeConfig.Host, s.server.erupeConfig.Entrance.Port), false)
	}

	lastPlayed := uint32(0)
	for _, char := range chars {
		if lastPlayed == 0 {
			lastPlayed = char.ID
		}
		bf.WriteUint32(char.ID)
		if s.server.erupeConfig.DebugOptions.MaxLauncherHR {
			bf.WriteUint16(999)
		} else {
			bf.WriteUint16(char.HR)
		}
		bf.WriteUint16(char.WeaponType)                                          // Weapon, 0-13.
		bf.WriteUint32(char.LastLogin)                                           // Last login date, unix timestamp in seconds.
		bf.WriteBool(char.IsFemale)                                              // Sex, 0=male, 1=female.
		bf.WriteBool(char.IsNewCharacter)                                        // Is new character, 1 replaces character name with ?????.
		bf.WriteUint8(0)                                                         // Old GR
		bf.WriteBool(true)                                                       // Use uint16 GR, no reason not to
		bf.WriteBytes(stringsupport.PaddedString(char.Name, 16, true))           // Character name
		bf.WriteBytes(stringsupport.PaddedString(char.UnkDescString, 32, false)) // unk str
		if s.server.erupeConfig.RealClientMode >= _config.G7 {
			bf.WriteUint16(char.GR)
			bf.WriteUint8(0) // Unk
			bf.WriteUint8(0) // Unk
		}
	}

	friends := s.server.getFriendsForCharacters(chars)
	if len(friends) == 0 {
		bf.WriteUint8(0)
	} else {
		if len(friends) > 255 {
			bf.WriteUint8(255)
			bf.WriteUint16(uint16(len(friends)))
		} else {
			bf.WriteUint8(uint8(len(friends)))
		}
		for _, friend := range friends {
			bf.WriteUint32(friend.CID)
			bf.WriteUint32(friend.ID)
			ps.Uint8(bf, friend.Name, true)
		}
	}

	guildmates := s.server.getGuildmatesForCharacters(chars)
	if len(guildmates) == 0 {
		bf.WriteUint8(0)
	} else {
		if len(guildmates) > 255 {
			bf.WriteUint8(255)
			bf.WriteUint16(uint16(len(guildmates)))
		} else {
			bf.WriteUint8(uint8(len(guildmates)))
		}
		for _, guildmate := range guildmates {
			bf.WriteUint32(guildmate.CID)
			bf.WriteUint32(guildmate.ID)
			ps.Uint8(bf, guildmate.Name, true)
		}
	}

	if s.server.erupeConfig.HideLoginNotice {
		bf.WriteBool(false)
	} else {
		bf.WriteBool(true)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		ps.Uint16(bf, strings.Join(s.server.erupeConfig.LoginNotices[:], "<PAGE>"), true)
	}

	bf.WriteUint32(s.server.getLastCID(uid))
	bf.WriteUint32(s.server.getUserRights(uid))
	ps.Uint16(bf, "", false) // filters
	if s.client == VITA || s.client == PS3 {
		var psnUser string
		s.server.db.QueryRow("SELECT psn_id FROM users WHERE id = $1", uid).Scan(&psnUser)
		bf.WriteBytes(stringsupport.PaddedString(psnUser, 20, true))
	}

	bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[0])
	if s.server.erupeConfig.DebugOptions.CapLink.Values[0] == 51728 {
		bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[1])
		if s.server.erupeConfig.DebugOptions.CapLink.Values[1] == 20000 || s.server.erupeConfig.DebugOptions.CapLink.Values[1] == 20002 {
			ps.Uint16(bf, s.server.erupeConfig.DebugOptions.CapLink.Key, false)
		}
	}
	caStruct := []struct {
		Unk0 uint8
		Unk1 uint32
		Unk2 string
	}{}
	bf.WriteUint8(uint8(len(caStruct)))
	for i := range caStruct {
		bf.WriteUint8(caStruct[i].Unk0)
		bf.WriteUint32(caStruct[i].Unk1)
		ps.Uint8(bf, caStruct[i].Unk2, false)
	}
	bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[2])
	bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[3])
	bf.WriteUint16(s.server.erupeConfig.DebugOptions.CapLink.Values[4])
	if s.server.erupeConfig.DebugOptions.CapLink.Values[2] == 51729 && s.server.erupeConfig.DebugOptions.CapLink.Values[3] == 1 && s.server.erupeConfig.DebugOptions.CapLink.Values[4] == 20000 {
		ps.Uint16(bf, fmt.Sprintf(`%s:%d`, s.server.erupeConfig.DebugOptions.CapLink.Host, s.server.erupeConfig.DebugOptions.CapLink.Port), false)
	}

	bf.WriteUint32(uint32(s.server.getReturnExpiry(uid).Unix()))
	bf.WriteUint32(0)

	tickets := []uint32{
		s.server.erupeConfig.GameplayOptions.MezFesSoloTickets,
		s.server.erupeConfig.GameplayOptions.MezFesGroupTickets,
	}
	stalls := []uint8{
		10, 3, 6, 9, 4, 8, 5, 7,
	}
	if s.server.erupeConfig.GameplayOptions.MezFesSwitchMinigame {
		stalls[4] = 2
	}

	// We can just use the start timestamp as the event ID
	bf.WriteUint32(uint32(channelserver.TimeWeekStart().Unix()))
	// Start time
	bf.WriteUint32(uint32(channelserver.TimeWeekNext().Add(-time.Duration(s.server.erupeConfig.GameplayOptions.MezFesDuration) * time.Second).Unix()))
	// End time
	bf.WriteUint32(uint32(channelserver.TimeWeekNext().Unix()))
	bf.WriteUint8(uint8(len(tickets)))
	for i := range tickets {
		bf.WriteUint32(tickets[i])
	}
	bf.WriteUint8(uint8(len(stalls)))
	for i := range stalls {
		bf.WriteUint8(stalls[i])
	}
	return bf.Data()
}
