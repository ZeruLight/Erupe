package channelserver

import (
	"encoding/hex"
	"encoding/json"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
)

func handleMsgMhfGetUdTacticsPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsPoint)
	personalPoints := make(map[uint16]int32)
	var totalPoints int32
	var temp []byte
	s.server.db.QueryRow(`SELECT interception_points FROM guild_characters WHERE id=$1`, s.charID).Scan(&temp)
	json.Unmarshal(temp, &personalPoints)
	for _, i := range personalPoints {
		totalPoints += i
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0x00) // Unk, some kind of error code?
	bf.WriteInt32(totalPoints)
	bf.WriteUint8(uint8(len(personalPoints)))
	for i := range personalPoints {
		bf.WriteUint16(i)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAddUdTacticsPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddUdTacticsPoint)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil || guild == nil {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	isApplicant, _ := guild.HasApplicationForCharID(s, s.charID)
	if err != nil || isApplicant {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	var personalPoints map[uint16]int32
	var temp []byte
	s.server.db.QueryRow(`SELECT interception_points FROM guild_characters WHERE id=$1`, s.charID).Scan(&temp)
	json.Unmarshal(temp, &personalPoints)
	if personalPoints == nil {
		personalPoints = make(map[uint16]int32)
		personalPoints[pkt.QuestFileID] = pkt.Points
	} else {
		personalPoints[pkt.QuestFileID] += pkt.Points
	}
	val, _ := json.Marshal(personalPoints)
	s.server.db.Exec(`UPDATE guild_characters SET interception_points=$1 WHERE id=$2`, val, s.charID)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0) // Unk
	bf.WriteUint8(uint8(len(personalPoints)))
	for i := range personalPoints {
		bf.WriteUint16(i)
	}

	if pkt.QuestFileID < 58079 || pkt.QuestFileID > 58083 {
		pkt.QuestFileID = 0
	}
	var mapData *InterceptionMaps
	s.server.db.QueryRow(`SELECT interception_maps FROM guilds WHERE id = $1`, guild.ID).Scan(&mapData)
	currID, _ := mapData.CurrPrevID()
	for i := range mapData.Maps {
		if mapData.Maps[i].ID == currID {
			mapData.Maps[i].Points[pkt.QuestFileID] += pkt.Points
		}
	}
	s.server.db.Exec(`UPDATE guilds SET interception_maps = $1 WHERE id = $2`, mapData, guild.ID)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

type DivaReward struct {
	Points     uint32 `db:"points_req"`
	ItemType   uint8  `db:"item_type"`
	ItemID     uint16 `db:"item_id"`
	Quantity   uint16 `db:"quantity"`
	GR         bool   `db:"gr"`
	Repeatable bool   `db:"repeatable"`
}

func handleMsgMhfGetUdTacticsRewardList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsRewardList)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	var personalRewards, guildRewards []DivaReward
	var tempReward DivaReward
	rows, err := s.server.db.Queryx(`SELECT points_req, item_type, item_id, quantity, gr, repeatable FROM diva_prizes WHERE type='personal'`)
	if err == nil {
		for rows.Next() {
			rows.StructScan(&tempReward)
			personalRewards = append(personalRewards, tempReward)
		}
	}
	rows, err = s.server.db.Queryx(`SELECT points_req, item_type, item_id, quantity, gr, repeatable FROM diva_prizes WHERE type='guild'`)
	if err == nil {
		for rows.Next() {
			rows.StructScan(&tempReward)
			guildRewards = append(guildRewards, tempReward)
		}
	}
	bf.WriteUint16(uint16(len(personalRewards)))
	for _, reward := range personalRewards {
		bf.WriteUint32(reward.Points)
		bf.WriteUint8(reward.ItemType)
		bf.WriteUint16(reward.ItemID)
		bf.WriteUint16(reward.Quantity)
		bf.WriteBool(reward.GR)
		bf.WriteBool(reward.Repeatable)
	}
	bf.WriteUint16(uint16(len(guildRewards)))
	for _, reward := range guildRewards {
		bf.WriteUint32(reward.Points)
		bf.WriteUint8(reward.ItemType)
		bf.WriteUint16(reward.ItemID)
		bf.WriteUint16(reward.Quantity)
		bf.WriteBool(reward.GR)
		bf.WriteBool(reward.Repeatable)
	}
	data, _ := hex.DecodeString("002607000E00C8000000010000000307000F0032000000010000000307001000320000000100000003070011003200000001000000030700120032000000010000000307000E0096000000040000000A07000F0028000000040000000A0700100028000000040000000A0700110028000000040000000A0700120028000000040000000A07000E00640000000B0000001907000F001E0000000B00000019070010001E0000000B00000019070011001E0000000B00000019070012001E0000000B0000001907000E00320000001A0000002807000F00140000001A0000002807001000140000001A0000002807001100140000001A0000002807001200140000001A0000002807000E001E000000290000004607000F000A0000002900000046070010000A000000290000004607001100010000002900000046070012000A000000290000004607000E0019000000470000006407000F0008000000470000006407001000080000004700000064070011000100000047000000640700120008000000470000006407000E000F000000650000009607000F0006000000650000009607001000010000006500000096070011000600000065000000960700120006000000650000009607000E000500000097000001F407000F000500000097000001F4070010000500000097000001F4")
	bf.WriteBytes(data)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdTacticsFollower(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsFollower)
	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetUdTacticsBonusQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsBonusQuest)
	// Temporary canned response
	data, _ := hex.DecodeString("14E2F55DCBFE505DCC1A7003E8E2C55DCC6ED05DCC8AF00258E2CE5DCCDF505DCCFB700279E3075DCD4FD05DCD6BF0041AE2F15DCDC0505DCDDC700258E2C45DCE30D05DCE4CF00258E2F55DCEA1505DCEBD7003E8E2C25DCF11D05DCF2DF00258E2CE5DCF82505DCF9E700279E3075DCFF2D05DD00EF0041AE2CE5DD063505DD07F700279E2F35DD0D3D05DD0EFF0028AE2C35DD144505DD160700258E2F05DD1B4D05DD1D0F00258E2CE5DD225505DD241700279E2F55DD295D05DD2B1F003E8E2F25DD306505DD3227002EEE2CA5DD376D05DD392F00258E3075DD3E7505DD40370041AE2F55DD457D05DD473F003E82027313220686F757273273A3A696E74657276616C29202B2027313220686F757273273A3A696E74657276616C2047524F5550204259206D6170204F52444552204259206D61703B2000C7312B000032")
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfGetUdTacticsFirstQuestBonus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsFirstQuestBonus)
	bonus := []struct {
		ID     uint8
		Points uint32
	}{
		{0, 1500},
		{1, 2000},
		{2, 2500},
		{3, 3000},
		{4, 4500},
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(bonus)))
	for i := range bonus {
		bf.WriteUint8(bonus[i].ID)
		bf.WriteUint32(bonus[i].Points)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdTacticsRemainingPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsRemainingPoint)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // Points until Special Guild Hall earned
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdTacticsRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsRanking)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // ranking
	bf.WriteUint32(0) // rankingDupe?
	bf.WriteUint32(0) // guildPoints
	bf.WriteUint32(0) // unk
	bf.WriteUint32(0) // unkDupe?
	bf.WriteUint32(0) // guildPointsDupe?
	bf.WriteBytes(stringsupport.PaddedString("", 25, true))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfSetUdTacticsFollower(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetUdTacticsLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsLog)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0) // Logs
	// Log format:
	// uint8 LogType, 0=addPoints, 1=tileClaimed, 5=newDate, 6=branchFinished
	// uint8 Unk
	// uint32 CharID
	// []byte CharName[32]
	// uint32 Value
	// uint32 Timestamp
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
