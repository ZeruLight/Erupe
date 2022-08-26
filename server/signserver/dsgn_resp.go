package signserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	"erupe-ce/server/channelserver"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

func makeSignInFailureResp(respID RespID) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(respID))
	return bf.Data()
}

func randSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (s *Session) makeSignInResp(uid int) []byte {
	returnExpiry := s.server.getReturnExpiry(uid)

	// Get the characters from the DB.
	chars, err := s.server.getCharactersForUser(uid)
	if err != nil {
		s.logger.Warn("Error getting characters from DB", zap.Error(err))
	}

	rand.Seed(time.Now().UnixNano())
	token := randSeq(16)
	s.server.registerToken(uid, token)

	bf := byteframe.NewByteFrame()

	bf.WriteUint8(1)                          // resp_code
	bf.WriteUint8(0)                          // file/patch server count
	bf.WriteUint8(1)                          // entrance server count
	bf.WriteUint8(uint8(len(chars)))          // character count
	bf.WriteUint32(0xFFFFFFFF)                // login_token_number
	bf.WriteBytes([]byte(token))              // login_token
	bf.WriteUint32(uint32(time.Now().Unix())) // current time
	ps.Uint8(bf, fmt.Sprintf("%s:%d", s.server.erupeConfig.Host, s.server.erupeConfig.Entrance.Port), false)

	lastPlayed := uint32(0)
	for _, char := range chars {
		if lastPlayed == 0 {
			lastPlayed = char.ID
		}
		bf.WriteUint32(char.ID)

		// Exp, HR[x] is split by 0, 1, 30, 50, 99, 299, 998, 999
		if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.MaxLauncherHR {
			bf.WriteUint16(999)
		} else {
			bf.WriteUint16(char.HRP)
		}

		bf.WriteUint16(char.WeaponType)                                          // Weapon, 0-13.
		bf.WriteUint32(char.LastLogin)                                           // Last login date, unix timestamp in seconds.
		bf.WriteBool(char.IsFemale)                                              // Sex, 0=male, 1=female.
		bf.WriteBool(char.IsNewCharacter)                                        // Is new character, 1 replaces character name with ?????.
		bf.WriteUint8(0)                                                         // Old GR
		bf.WriteBool(true)                                                       // Use uint16 GR, no reason not to
		bf.WriteBytes(stringsupport.PaddedString(char.Name, 16, true))           // Character name
		bf.WriteBytes(stringsupport.PaddedString(char.UnkDescString, 32, false)) // unk str
		bf.WriteUint16(char.GR)
		bf.WriteUint16(0) // Unk
	}

	friends := s.server.getFriendsForCharacters(chars)
	if len(friends) == 0 {
		bf.WriteUint8(0)
	} else {
		bf.WriteUint8(uint8(len(friends)))
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
		bf.WriteUint8(uint8(len(guildmates)))
		for _, guildmate := range guildmates {
			bf.WriteUint32(guildmate.CID)
			bf.WriteUint32(guildmate.ID)
			ps.Uint8(bf, guildmate.Name, true)
		}
	}

	if s.server.erupeConfig.DevModeOptions.HideLoginNotice {
		bf.WriteUint8(0)
	} else {
		bf.WriteUint8(1) // Notice count
		noticeText := s.server.erupeConfig.DevModeOptions.LoginNotice
		ps.Uint32(bf, noticeText, true)
	}

	bf.WriteUint32(s.server.getLastCID(uid))
	bf.WriteUint32(s.server.getUserRights(uid))
	ps.Uint16(bf, "", false) // filters
	bf.WriteUint32(0xCA104E20)
	ps.Uint16(bf, "", false) // encryption
	bf.WriteUint8(0x00)
	bf.WriteUint32(0xCA110001)
	bf.WriteUint32(0x4E200000)
	bf.WriteUint32(uint32(returnExpiry.Unix()))
	bf.WriteUint32(0x00000000)
	bf.WriteUint32(0x0A5197DF)

	mezfes := s.server.erupeConfig.DevModeOptions.MezFesEvent
	alt := s.server.erupeConfig.DevModeOptions.MezFesAlt
	if mezfes {
		// Start time
		bf.WriteUint32(uint32(channelserver.Time_Current_Adjusted().Add(-5 * time.Minute).Unix()))
		// End time
		bf.WriteUint32(uint32(channelserver.Time_Current_Adjusted().Add(24 * time.Hour * 7).Unix()))
		bf.WriteUint8(2)   // Unk
		bf.WriteUint32(20) // Single tickets
		bf.WriteUint32(10) // Group tickets
		bf.WriteUint8(8)   // Stalls open
		bf.WriteUint8(0xA) // Unk
		bf.WriteUint8(0x3) // Pachinko
		bf.WriteUint8(0x6) // Nyanrendo
		bf.WriteUint8(0x9) // Point stall
		if alt {
			bf.WriteUint8(0x2) // Tokotoko
		} else {
			bf.WriteUint8(0x4) // Volpakkun
		}
		bf.WriteUint8(0x8) // Battle cats
		bf.WriteUint8(0x5) // Gook
		bf.WriteUint8(0x7) // Honey
	} else {
		bf.WriteUint32(0)
		bf.WriteUint32(0)
	}
	return bf.Data()
}
