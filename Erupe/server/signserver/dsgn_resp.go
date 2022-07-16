package signserver

import (
	"fmt"
	"math/rand"
	"time"
	"erupe-ce/server/channelserver"
	"erupe-ce/common/stringsupport"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/byteframe"

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
	bf.WriteUint32(uint32(time.Now().Unix())) // unk timestamp
	ps.Uint8(bf, fmt.Sprintf("%s:%d", s.server.erupeConfig.HostIP, s.server.erupeConfig.Entrance.Port), false)

	for _, char := range chars {
		bf.WriteUint32(char.ID)

		// Exp, HR[x] is split by 0, 1, 30, 50, 99, 299, 998, 999
		if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.MaxLauncherHR {
			bf.WriteUint16(999)
		} else {
			bf.WriteUint16(char.HRP)
		}

		bf.WriteUint16(char.WeaponType)                     // Weapon, 0-13.
		bf.WriteUint32(char.LastLogin)                      // Last login date, unix timestamp in seconds.
		bf.WriteBool(char.IsFemale)                         // Sex, 0=male, 1=female.
		bf.WriteBool(char.IsNewCharacter)                   // Is new character, 1 replaces character name with ?????.
		bf.WriteUint8(0)                                    // Old GR
		bf.WriteBool(true)                                  // Use uint16 GR, no reason not to
		bf.WriteBytes(stringsupport.PaddedString(char.Name, 16, true))          // Character name
		bf.WriteBytes(stringsupport.PaddedString(char.UnkDescString, 32, false)) // unk str
		bf.WriteUint16(char.GR)
		bf.WriteUint16(0) // Unk
	}

	bf.WriteUint8(0)           // friends_list_count
	bf.WriteUint8(0)           // guild_members_count
	bf.WriteUint8(0)           // notice_count

	// noticeText := "<BODY><CENTER><SIZE_3><C_4>Welcome to Erupe SU9!<BR><BODY><LEFT><SIZE_2><C_5>Erupe is experimental software<C_7>, we are not liable for any<BR><BODY>issues caused by installing the software!<BR><BODY><BR><BODY><C_4>■Report bugs on Discord!<C_7><BR><BODY><BR><BODY><C_4>■Test everything!<C_7><BR><BODY><BR><BODY><C_4>■Don't talk to softlocking NPCs!<C_7><BR><BODY><BR><BODY><C_4>■Fork the code on GitHub!<C_7><BR><BODY><BR><BODY>Thank you to all of the contributors,<BR><BODY><BR><BODY>this wouldn't exist without you."
	// ps.Uint32(bf, noticeText, true)

	bf.WriteUint32(0)          // some_last_played_character_id
	bf.WriteUint32(14)         // unk_flags
	ps.Uint16(bf, "", false) // filters
	bf.WriteUint32(0xCA104E20)
	ps.Uint16(bf, "", false) // encryption
	bf.WriteUint8(0x00)
	bf.WriteUint32(0xCA110001)
	bf.WriteUint32(0x4E200000)

	returning := false
	// return course end time
	if returning {
		bf.WriteUint32(uint32(channelserver.Time_Current_Adjusted().Add(30 * 24 * time.Hour).Unix()))
	} else {
		bf.WriteUint32(0)
	}

	bf.WriteUint32(0x00000000)
	bf.WriteUint32(0x0A5197DF)

	mezfes := true
	alt := false
	if mezfes {
		bf.WriteUint32(uint32(channelserver.Time_Current_Adjusted().Add(-5 * time.Minute).Unix())) // Start time
		bf.WriteUint32(uint32(channelserver.Time_Current_Adjusted().Add(24 * time.Hour * 7).Unix())) // End time
		bf.WriteUint8(2) // Unk
		bf.WriteUint32(0) // Single tickets
		bf.WriteUint32(0) // Group tickets
		bf.WriteUint8(8) // Stalls open
		bf.WriteUint8(0xA) // Unk
		bf.WriteUint8(0x3) // Pachinko
		bf.WriteUint8(0x6) // Nyanrendo
		bf.WriteUint8(0x9) // Point stall
		if alt {
			bf.WriteUint8(0x2) // Tokotoko
		} else {
			bf.WriteUint8(0x4) // Volpkun
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
