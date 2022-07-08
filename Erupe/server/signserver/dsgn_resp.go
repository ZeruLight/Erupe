package signserver

import (
	"fmt"
	"math/rand"
	"time"
	"erupe-ce/server/channelserver"

	"github.com/Andoryuuta/byteframe"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func paddedString(x string, size uint) []byte {
	out := make([]byte, size)
	copy(out, x)

	// Null terminate it.
	out[len(out)-1] = 0
	return out
}

func uint8PascalString(bf *byteframe.ByteFrame, x string) {
	bf.WriteUint8(uint8(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

func uint16PascalString(bf *byteframe.ByteFrame, x string) {
	bf.WriteUint16(uint16(len(x) + 1))
	bf.WriteNullTerminatedBytes([]byte(x))
}

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
	// TODO: register token to db, users table

	t := japanese.ShiftJIS.NewEncoder()
	bf := byteframe.NewByteFrame()

	bf.WriteUint8(1)                          // resp_code
	bf.WriteUint8(0)                          // file/patch server count
	bf.WriteUint8(1)                          // entrance server count
	bf.WriteUint8(uint8(len(chars)))          // character count
	bf.WriteUint32(0xFFFFFFFF)                // login_token_number
	bf.WriteBytes(paddedString(token, 16))    // login_token (16 byte padded string)
	bf.WriteUint32(uint32(time.Now().Unix())) // unk timestamp
	uint8PascalString(bf, fmt.Sprintf("%s:%d", s.server.erupeConfig.HostIP, s.server.erupeConfig.Entrance.Port))

	for _, char := range chars {
		bf.WriteUint32(char.ID)

		// Exp, HR[x] is split by 0, 1, 30, 50, 99, 299, 998, 999
		if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.MaxLauncherHR {
			bf.WriteUint16(999)
		} else {
			bf.WriteUint16(char.HRP)
		}

		str_name, _, err := transform.String(t, char.Name)
		if err != nil {
		  str_name = char.Name
		}

		bf.WriteUint16(char.WeaponType)                     // Weapon, 0-13.
		bf.WriteUint32(char.LastLogin)                      // Last login date, unix timestamp in seconds.
		bf.WriteBool(char.IsFemale)                         // Sex, 0=male, 1=female.
		bf.WriteBool(char.IsNewCharacter)                   // Is new character, 1 replaces character name with ?????.
		bf.WriteUint8(0)                                    // Old GR
		bf.WriteBool(true)                                  // Use uint16 GR, no reason not to
		bf.WriteBytes(paddedString(str_name, 16))          // Character name
		bf.WriteBytes(paddedString(char.UnkDescString, 32)) // unk str
		bf.WriteUint16(char.GR)
		bf.WriteUint16(0) // Unk
	}

	bf.WriteUint8(0)           // friends_list_count
	bf.WriteUint8(0)           // guild_members_count
	bf.WriteUint8(0)           // notice_count

	// noticeText := "<BODY><CENTER><SIZE_3><C_4>Welcome to Erupe SU9!<BR><BODY><LEFT><SIZE_2><C_5>Erupe is experimental software<C_7>, we are not liable for any<BR><BODY>issues caused by installing the software!<BR><BODY><BR><BODY><C_4>■Report bugs on Discord!<C_7><BR><BODY><BR><BODY><C_4>■Test everything!<C_7><BR><BODY><BR><BODY><C_4>■Don't talk to softlocking NPCs!<C_7><BR><BODY><BR><BODY><C_4>■Fork the code on GitHub!<C_7><BR><BODY><BR><BODY>Thank you to all of the contributors,<BR><BODY><BR><BODY>this wouldn't exist without you."
	// notice_transformed, _, err := transform.String(t, noticeText)
	// if err != nil {
	// 	panic(err)
	// }
	// bf.WriteUint32(uint32(len(notice_transformed)+1))
	// bf.WriteNullTerminatedBytes([]byte(notice_transformed))

	bf.WriteUint32(0)          // some_last_played_character_id
	bf.WriteUint32(14)         // unk_flags
	uint16PascalString(bf, "") // filters
	bf.WriteUint32(0xCA104E20)
	uint16PascalString(bf, "") // encryption
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
