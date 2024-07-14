package signserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"erupe-ce/server/channelserver"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
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

	namNGWords := []string{"test", "痴女", "てすと"}
	msgNGWords := []string{"test", "痴女", "てすと"}

	filters := byteframe.NewByteFrame()
	filters.WriteNullTerminatedBytes([]byte("smc"))
	smc := byteframe.NewByteFrame()
	//smcBytes, _ := hex.DecodeString("3D000000818100000000000029000000816A0000000000002800000081690000000000002100000081490000000000002F000000815E0000000000002B000000817B00000000000026000000819500000000000082500000310000000000000082DA0000837B0000CE00DE0082D9DE00837ADE0082D9814A837A814ACE00814A0000000082D7000083780000CD00DE0082D6DE008377DE0082D6814ACD00814A8377814A0000000082C5000083660000C300DE0082C4DE008365DE0082C4814A8365814AC300814A81A7814A81A7DE0089B3DE0089B3814A0000000082D1000083720000CB00DE0082D0DE008371DE0082D0814A8371814ACB00814A0000000082C7000083680000C400DE0082C6DE008367DE0082C6814A8367814AC400814A84B0DE0084B0814A84A5DE0084A5814A0000000082CE0000836F0000CA00DE0082CDDE00836EDE0094AADE0082CD814A836E814ACA00814A94AA814A0000000082C2DE00836400008363DE0082C2814A8363814AC200DE00C200814A82C3000082C1DE008362DE00AF00DE0082C1814A8362814AAF00814A0000000082D4000083750000CC00DE0083940000B300DE0082A4814A82A4DE008345DE00A900DE0082A3DE0082D3DE008374DE00CC00814A0000000082C0000083610000C100DE0082BFDE008360DE0082BF814A8360814AC100814A90E7814A90E7DE000000000082BE0000835F0000C000DE0082BDDE00835EDE00975BDE0082BD814A835E814AC000814A975B814A0000000082BC0000835D0000BF00DE0082BBDE00835CDE0082BB814A835C814ABF00814A8393DE008393814ADD00814ADD00DE00838ADE00D800DE00D800814A838A814A0000000082BA0000BE00DE0082B9DE00835ADE0082B9814A835A814ABE00814A835B00000000000082B8000083590000BD00DE0082B7DE008358DE0082B7814A8358814ABD00814A0000000082B6000083570000BC00DE0082B5DE008356DE0082B5814A8356814ABC00814A0000000082B4000083550000BB00DE0082B3DE008354DE0082B3814A8354814ABB00814A0000000082B2000083530000BA00DE0082B1DE008352DE0082B1814A8352814ABA00814A0000000082B0000083510000B900DE0082AFDE008350DE0082AF814A8350814AB900814A8396DE008396814A0000000082AE0000834F0000B800DE0082ADDE00834EDE0082AD814A834E814AB800814A0000000082AC0000834D0000B700DE0082ABDE00834CDE0082AB814A834C814AB700814A0000000082AA0000834B0000B600DE008395DE00834ADE0082A9DE0097CDDE008395814A834A814A82A9814A97CD814AB600814A0000000082F0000083920000A60000000000000082ED0000838F0000DC000000838E00000000000082EB0000838D0000DB00000081A000008CFB00000000000082EA0000838C0000DA0000000000000082E90000838B0000D90000000000000082E80000838A0000D80000000000000082E7000083890000D70000000000000082E6000083880000D6000000AE00000082E50000838700000000000082E4000083860000D5000000AD00000082E30000838500000000000082E2000083840000D4000000AC00000082E10000838300000000000082E0000083820000D30000000000000082DF000083810000D20000004D0045000000000082DE000083800000D10000000000000082DD0000837E0000D00000000000000082DC0000837D0000CF0000000000000082D90000837A0000CE0000000000000082D6000083770000CD0000000000000082D3000083740000CC0000000000000082D0000083710000CB0000000000000082CD0000836E0000CA00000094AA00000000000082CC0000836D0000C90000000000000082CB0000836C0000C80000000000000082CA0000836B0000C70000000000000082C90000836A0000C600000093F100000000000082C8000083690000C50000000000000082C6000083670000C400000084B0000084A500000000000082C4000083650000C300000081A7000089B300000000000082C2000083630000C200000082C1000083620000AF0000000000000082BF000083600000C100000090E700000000000082BD0000835E0000C0000000975B00000000000082BB0000835C0000BF0000000000000082B90000835A0000BE0000000000000082B7000083580000BD0000000000000082B5000083560000BC0000000000000082B3000083540000BB0000000000000082B1000083520000BA0000000000000082AF000083500000B9000000839600000000000082AD0000834E0000B80000000000000082AB0000834C0000B70000000000000082A90000834A0000B60000008395000097CD00000000000082A8000083490000B5000000AB00000082A70000834800000000000082A6000083470000B4000000AA00000082A50000834600008D4800000000000082A4000083450000B3000000A900000082A30000834400000000000082A2000083430000B2000000A800000082A100008342000000000000A7000000B1000000829F000082A00000834000008341000000000000815B0000815C0000815D0000816000002D000000817C0000B000000088EA00000000000039000000825800000000000038000000825700000000000037000000825600000000000036000000825500000000000035000000825400000000000034000000825300000000000033000000825200000000000032000000825100000000000082DB0000837C0000CE00DF0082D9DF00837ADF00837A818B82D9818BCE00818B0000000082D8000083790000CD00DF0082D6DF008377DF008377818B82D6818BCD00818B0000000082D5000083760000CC00DF0082D3DF008374DF008374818B82D3818BCC00818B0000000082D2000083730000CB00DF0082D0DF008371DF008371818B82D0818BCB00818B0000000082CF000083700000CA00DF0082CDDF00836EDF00836E818B82CD818BCA00818B94AADF0094AA814B0000000082790000829A00005A0000007A00000083A40000000000008278000082990000590000007900000083B200008454000084850000000000008277000082980000580000007800000083B4000083D4000084560000817E000084870000875D0000000000008276000082970000570000007700000083D6000084590000848A0000848B0000000000008275000082960000560000007600000083CB000083D2000087580000000000008274000082950000550000007500000083CA000081BE0000000000008273000082940000540000007400000083B1000083D100008453000084840000000000002400000053000000730000008190000081E70000827200008293000000000000827100008292000052000000720000008460000084910000000000008270000082910000510000007100000000000000826F000082900000500000007000000083AF000083CF0000845100008482000000000000826E0000828F00004F0000006F000000819B000083AD000083CD0000844F00008480000081FC0000815A000030000000824F000000000000826D0000828E00004E0000006E00000083AB000083C50000DD00000082F100008393000000000000826C0000828D00004D0000006D00000083AA0000844D0000847D000000000000826B0000828C00004C0000006C000000879800007C00000000000000826A0000828B00004B0000006B00000083A8000083C80000844B0000847B00000000000082690000828A00004A0000006A000000000000008268000082890000490000006900000083A7000087540000000000008267000082880000480000006800000083A50000844E0000847E000000000000826500008286000046000000660000000000000082660000828700004700000067000000000000008264000082850000450000006500000083A3000083C300008445000084460000847500008476000081B80000000000008263000082840000440000006400000000000000430000008283000063000000845200008483000082620000818E0000000000004200000082610000828200006200000083C0000083A000008442000084720000848C0000848E000081F30000000000002700000081660000000000004100000082600000610000008281000083BF000040000000819700008470000081F0000084400000839F00000000000022000000816800000000000025000000819300000000000000000000")
	//smc.WriteBytes(smcBytes)
	filters.SetLE()
	filters.WriteUint32(uint32(len(smc.Data())))
	filters.WriteBytes(smc.Data())

	filters.WriteNullTerminatedBytes([]byte("nam"))
	nam := byteframe.NewByteFrame()
	nam.SetLE()
	for _, word := range namNGWords {
		parts := stringsupport.ToNGWord(word)
		nam.WriteUint32(uint32(len(parts)))
		for _, part := range parts {
			nam.WriteUint16(part)
			nam.WriteInt16(-1) // TODO: figure out how this value relates to corresponding SMC part
		}
		nam.WriteUint16(0)
		nam.WriteInt16(-1)
	}
	filters.WriteUint32(uint32(len(nam.Data())))
	filters.WriteBytes(nam.Data())

	filters.WriteNullTerminatedBytes([]byte("msg"))
	msg := byteframe.NewByteFrame()
	msg.SetLE()
	for _, word := range msgNGWords {
		parts := stringsupport.ToNGWord(word)
		msg.WriteUint32(uint32(len(parts)))
		for _, part := range parts {
			msg.WriteUint16(part)
			msg.WriteInt16(-1)
		}
		msg.WriteUint16(0)
		msg.WriteInt16(-1)
	}
	filters.WriteUint32(uint32(len(msg.Data())))
	filters.WriteBytes(msg.Data())

	bf.WriteUint16(uint16(len(filters.Data())))
	bf.WriteBytes(filters.Data())

	if s.client == VITA || s.client == PS3 || s.client == PS4 {
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
