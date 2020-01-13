package signserver

import "github.com/Andoryuuta/byteframe"

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

func (session *Session) makeSignInResp(username string) []byte {
	bf := byteframe.NewByteFrame()

	// delete me:
	//bf.WriteUint8(8)
	//return bf.Data()

	bf.WriteUint8(1)                                   // resp_code
	bf.WriteUint8(0)                                   // file/patch server count
	bf.WriteUint8(4)                                   // entrance server count
	bf.WriteUint8(1)                                   // character count
	bf.WriteUint32(0xFFFFFFFF)                         // login_token_number
	bf.WriteBytes(paddedString("logintokenstrng", 16)) // login_token (16 byte padded string)
	bf.WriteUint32(1576761190)

	// file patch server PascalStrings here

	// Array(this.entrance_server_count, PascalString(Byte, "utf8")),
	uint8PascalString(bf, "localhost:53310")
	uint8PascalString(bf, "")
	uint8PascalString(bf, "")
	uint8PascalString(bf, "mhf-n.capcom.com.tw")

	///////////////////////////
	// Characters:

	/*
		tab = '123456789ABCDEFGHJKLMNPQRTUVWXYZ'
		def make_uid_str(cid):
			out = ''
			for i in range(6):
				v = (cid>>5*i)
				out += tab[v&0x1f]
			return out

		def make_cid_int(uid):
			v = 0
			for c in uid[::-1]:
				idx = tab.find(c)
				if idx == -1:
					raise Exception("not in tab")
				v |= idx
				v = v<<5
			return v>>5
	*/
	bf.WriteUint32(469153291) // character ID 469153291
	bf.WriteUint16(999)       // Exp, HR[x] is split by 0, 1, 30, 50, 99, 299, 998, 999

	//44.204

	/*
		0=大劍/Big sword
		1=重弩/Heavy crossbow
		2=大錘/Sledgehammer
		3=長槍/Spear
		4=單手劍/One-handed sword
		5=輕弩/Light crossbow
		6=雙劍/Double sword
		7=太刀/Tadao
		8=狩獵笛/Hunting flute
		9=銃槍/Shotgun
		10=弓/bow
		11=穿龍棍/Wear a dragon stick
		12=斬擊斧F/Chopping Axe F
		13=---
		default=不明/unknown
	*/
	bf.WriteUint16(7) // Weapon, 0-13.

	bf.WriteUint32(1576761172) // Last login date, unix timestamp in seconds.
	bf.WriteUint8(1)           // Sex, 0=male, 1=female.
	bf.WriteUint8(0)           // Is new character, 1 replaces character name with ?????.
	grMode := uint8(0)
	bf.WriteUint8(1)                          // GR level if grMode == 0
	bf.WriteUint8(grMode)                     // GR mode.
	bf.WriteBytes(paddedString(username, 16)) // Character name
	bf.WriteBytes(paddedString("0", 32))      // unk str
	if grMode == 1 {
		bf.WriteUint16(55) // GR level override.
		bf.WriteUint8(0)   // unk
		bf.WriteUint8(0)   // unk
	}

	//////////////////////////

	bf.WriteUint8(0)           // friends_list_count
	bf.WriteUint8(0)           // guild_members_count
	bf.WriteUint8(0)           // notice_count
	bf.WriteUint32(0xDEADBEEF) // some_last_played_character_id
	bf.WriteUint32(14)         // unk_flags
	uint8PascalString(bf, "")  // unk_data_blob PascalString

	bf.WriteUint16(51728)
	bf.WriteUint16(20000)
	uint16PascalString(bf, "1000672925")

	bf.WriteUint8(0)

	bf.WriteUint16(51729)
	bf.WriteUint16(1)
	bf.WriteUint16(20000)
	uint16PascalString(bf, "203.191.249.36:8080")

	bf.WriteUint32(1578905116)
	bf.WriteUint32(0)

	return bf.Data()
}
