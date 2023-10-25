package channelserver

import (
	"encoding/json"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"time"
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
	bf.WriteUint8(0) // No error
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

type DivaRankReward struct {
	Unk0    uint8
	Unk1    uint16
	Unk2    uint16
	MaxRank uint32
	MinRank uint32
}

func handleMsgMhfGetUdTacticsRewardList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsRewardList)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0) // No error
	var personalRewards, guildRewards []DivaReward
	var rankRewards []DivaRankReward
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
	bf.WriteUint16(uint16(len(rankRewards)))
	for _, reward := range rankRewards {
		bf.WriteUint8(reward.Unk0)
		bf.WriteUint16(reward.Unk1)
		bf.WriteUint16(reward.Unk2)
		bf.WriteUint32(reward.MaxRank)
		bf.WriteUint32(reward.MinRank)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdTacticsFollower(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsFollower)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	bf.WriteUint32(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdTacticsBonusQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTacticsBonusQuest)
	bonusQuests := []struct {
		QuestFileID uint16
		Start       time.Time
		End         time.Time
		Points      uint16
	}{}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(bonusQuests)))
	for _, quest := range bonusQuests {
		bf.WriteUint16(quest.QuestFileID)
		bf.WriteUint32(uint32(quest.Start.Unix()))
		bf.WriteUint32(uint32(quest.End.Unix()))
		bf.WriteUint16(quest.Points)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
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
	bf.WriteUint32(0) // Own Rank
	bf.WriteUint32(0) // Own Score (Areas)
	bf.WriteBytes(stringsupport.PaddedString("", 32, true))
	bf.WriteUint8(0) // Num other ranks
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfSetUdTacticsFollower(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetUdTacticsFollower)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

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
