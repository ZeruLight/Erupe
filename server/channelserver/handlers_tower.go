package channelserver

import (
	"erupe-ce/config"
	"erupe-ce/internal/model"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/stringsupport"
)

func EmptyTowerCSV(len int) string {
	temp := make([]string, len)
	for i := range temp {
		temp[i] = "0"
	}
	return strings.Join(temp, ",")
}

func handleMsgMhfGetTowerInfo(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTowerInfo)
	var data []*byteframe.ByteFrame

	towerInfo := model.TowerInfo{
		TRP:     []model.TowerInfoTRP{{0, 0}},
		Skill:   []model.TowerInfoSkill{{0, make([]int16, 64)}},
		History: []model.TowerInfoHistory{{make([]int16, 5), make([]int16, 5)}},
		Level:   []model.TowerInfoLevel{{0, 0, 0, 0}, {0, 0, 0, 0}},
	}

	var tempSkills string
	err := db.QueryRow(`SELECT COALESCE(tr, 0),  COALESCE(trp, 0),  COALESCE(tsp, 0), COALESCE(block1, 0), COALESCE(block2, 0), COALESCE(skills, $1) FROM tower WHERE char_id=$2
		`, EmptyTowerCSV(64), s.CharID).Scan(&towerInfo.TRP[0].TR, &towerInfo.TRP[0].TRP, &towerInfo.Skill[0].TSP, &towerInfo.Level[0].Floors, &towerInfo.Level[1].Floors, &tempSkills)
	if err != nil {
		db.Exec(`INSERT INTO tower (char_id) VALUES ($1)`, s.CharID)
	}

	if config.GetConfig().ClientID <= config.G7 {
		towerInfo.Level = towerInfo.Level[:1]
	}

	for i, skill := range stringsupport.CSVElems(tempSkills) {
		towerInfo.Skill[0].Skills[i] = int16(skill)
	}

	switch pkt.InfoType {
	case 1:
		for _, trp := range towerInfo.TRP {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(trp.TR)
			bf.WriteInt32(trp.TRP)
			data = append(data, bf)
		}
	case 2:
		for _, skills := range towerInfo.Skill {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(skills.TSP)
			for i := range skills.Skills {
				bf.WriteInt16(skills.Skills[i])
			}
			data = append(data, bf)
		}
	case 4:
		for _, history := range towerInfo.History {
			bf := byteframe.NewByteFrame()
			for i := range history.Unk0 {
				bf.WriteInt16(history.Unk0[i])
			}
			for i := range history.Unk1 {
				bf.WriteInt16(history.Unk1[i])
			}
			data = append(data, bf)
		}
	case 3, 5:
		for _, level := range towerInfo.Level {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(level.Floors)
			bf.WriteInt32(level.Unk1)
			bf.WriteInt32(level.Unk2)
			bf.WriteInt32(level.Unk3)
			data = append(data, bf)
		}
	}
	s.DoAckEarthSucceed(pkt.AckHandle, data)
}

func handleMsgMhfPostTowerInfo(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTowerInfo)

	if config.GetConfig().DebugOptions.QuestTools {
		s.Logger.Debug(
			p.Opcode().String(),
			zap.Uint32("InfoType", pkt.InfoType),
			zap.Uint32("Unk1", pkt.Unk1),
			zap.Int32("Skill", pkt.Skill),
			zap.Int32("TR", pkt.TR),
			zap.Int32("TRP", pkt.TRP),
			zap.Int32("Cost", pkt.Cost),
			zap.Int32("Unk6", pkt.Unk6),
			zap.Int32("Unk7", pkt.Unk7),
			zap.Int32("Block1", pkt.Block1),
			zap.Int64("Unk9", pkt.Unk9),
		)
	}

	switch pkt.InfoType {
	case 2:
		var skills string
		db.QueryRow(`SELECT COALESCE(skills, $1) FROM tower WHERE char_id=$2`, EmptyTowerCSV(64), s.CharID).Scan(&skills)
		db.Exec(`UPDATE tower SET skills=$1, tsp=tsp-$2 WHERE char_id=$3`, stringsupport.CSVSetIndex(skills, int(pkt.Skill), stringsupport.CSVGetIndex(skills, int(pkt.Skill))+1), pkt.Cost, s.CharID)
	case 1, 7:
		// This might give too much TSP? No idea what the rate is supposed to be
		db.Exec(`UPDATE tower SET tr=$1, trp=COALESCE(trp, 0)+$2, tsp=COALESCE(tsp, 0)+$3, block1=COALESCE(block1, 0)+$4 WHERE char_id=$5`, pkt.TR, pkt.TRP, pkt.Cost, pkt.Block1, s.CharID)
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

// Default missions
var tenrouiraiData = []model.TenrouiraiData{
	{1, 1, 80, 0, 2, 2, 1, 1, 2, 2},
	{1, 4, 16, 0, 2, 2, 1, 1, 2, 2},
	{1, 6, 50, 0, 2, 2, 1, 0, 2, 2},
	{1, 4, 12, 50, 2, 2, 1, 1, 2, 2},
	{1, 3, 50, 0, 2, 2, 1, 1, 2, 2},
	{2, 5, 40000, 0, 2, 2, 1, 0, 2, 2},
	{1, 5, 50000, 50, 2, 2, 1, 1, 2, 2},
	{2, 1, 60, 0, 2, 2, 1, 1, 2, 2},
	{2, 3, 50, 0, 2, 1, 1, 0, 1, 2},
	{2, 3, 40, 50, 2, 1, 1, 1, 1, 2},
	{2, 4, 12, 0, 2, 1, 1, 1, 1, 2},
	{2, 6, 40, 0, 2, 1, 1, 0, 1, 2},
	{1, 1, 60, 50, 2, 1, 2, 1, 1, 2},
	{1, 5, 50000, 0, 3, 1, 2, 1, 1, 2},
	{1, 6, 50, 0, 3, 1, 2, 0, 1, 2},
	{1, 4, 16, 50, 3, 1, 2, 1, 1, 2},
	{1, 5, 50000, 0, 3, 1, 2, 1, 1, 2},
	{2, 3, 40, 0, 3, 1, 2, 0, 1, 2},
	{1, 3, 50, 50, 3, 1, 2, 1, 1, 2},
	{2, 5, 40000, 0, 3, 1, 2, 1, 1, 1},
	{2, 6, 40, 0, 3, 1, 2, 0, 1, 1},
	{2, 1, 60, 50, 3, 1, 2, 1, 1, 1},
	{2, 6, 50, 0, 3, 1, 2, 1, 1, 1},
	{2, 4, 12, 0, 3, 1, 2, 0, 1, 1},
	{1, 1, 80, 50, 3, 1, 2, 1, 1, 1},
	{1, 5, 40000, 0, 3, 1, 2, 1, 1, 1},
	{1, 3, 50, 0, 3, 1, 2, 0, 1, 1},
	{1, 4, 16, 50, 3, 1, 0, 1, 1, 1},
	{1, 6, 50, 0, 3, 1, 0, 1, 1, 1},
	{2, 3, 40, 0, 3, 1, 0, 1, 1, 1},
	{1, 1, 80, 50, 3, 1, 0, 0, 1, 1},
	{2, 5, 40000, 0, 3, 1, 0, 0, 1, 1},
	{2, 6, 40, 0, 3, 1, 0, 0, 1, 1},
}

func handleMsgMhfGetTenrouirai(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTenrouirai)
	var data []*byteframe.ByteFrame

	tenrouirai := model.Tenrouirai{
		Progress: []model.TenrouiraiProgress{{1, 0, 0, 0}},
		Data:     tenrouiraiData,
		Ticket:   []model.TenrouiraiTicket{{0, 0, 0}},
	}

	switch pkt.Unk1 {
	case 1:
		for _, tdata := range tenrouirai.Data {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(tdata.Block)
			bf.WriteUint8(tdata.Mission)
			bf.WriteUint16(tdata.Goal)
			bf.WriteUint16(tdata.Cost)
			bf.WriteUint8(tdata.Skill1)
			bf.WriteUint8(tdata.Skill2)
			bf.WriteUint8(tdata.Skill3)
			bf.WriteUint8(tdata.Skill4)
			bf.WriteUint8(tdata.Skill5)
			bf.WriteUint8(tdata.Skill6)
			data = append(data, bf)
		}
	case 2:
		for _, reward := range tenrouirai.Reward {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(reward.Index)
			bf.WriteUint16(reward.Item[0])
			bf.WriteUint16(reward.Item[1])
			bf.WriteUint16(reward.Item[2])
			bf.WriteUint16(reward.Item[3])
			bf.WriteUint16(reward.Item[4])
			bf.WriteUint8(reward.Quantity[0])
			bf.WriteUint8(reward.Quantity[1])
			bf.WriteUint8(reward.Quantity[2])
			bf.WriteUint8(reward.Quantity[3])
			bf.WriteUint8(reward.Quantity[4])
			data = append(data, bf)
		}
	case 4:
		db.QueryRow(`SELECT tower_mission_page FROM guilds WHERE id=$1`, pkt.GuildID).Scan(&tenrouirai.Progress[0].Page)
		db.QueryRow(`SELECT SUM(tower_mission_1) AS _, SUM(tower_mission_2) AS _, SUM(tower_mission_3) AS _ FROM guild_characters WHERE guild_id=$1
				`, pkt.GuildID).Scan(&tenrouirai.Progress[0].Mission1, &tenrouirai.Progress[0].Mission2, &tenrouirai.Progress[0].Mission3)

		if tenrouirai.Progress[0].Mission1 > tenrouiraiData[(tenrouirai.Progress[0].Page*3)-3].Goal {
			tenrouirai.Progress[0].Mission1 = tenrouiraiData[(tenrouirai.Progress[0].Page*3)-3].Goal
		}
		if tenrouirai.Progress[0].Mission2 > tenrouiraiData[(tenrouirai.Progress[0].Page*3)-2].Goal {
			tenrouirai.Progress[0].Mission2 = tenrouiraiData[(tenrouirai.Progress[0].Page*3)-2].Goal
		}
		if tenrouirai.Progress[0].Mission3 > tenrouiraiData[(tenrouirai.Progress[0].Page*3)-1].Goal {
			tenrouirai.Progress[0].Mission3 = tenrouiraiData[(tenrouirai.Progress[0].Page*3)-1].Goal
		}

		for _, progress := range tenrouirai.Progress {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(progress.Page)
			bf.WriteUint16(progress.Mission1)
			bf.WriteUint16(progress.Mission2)
			bf.WriteUint16(progress.Mission3)
			data = append(data, bf)
		}
	case 5:
		if pkt.Unk3 > 3 {
			pkt.Unk3 %= 3
			if pkt.Unk3 == 0 {
				pkt.Unk3 = 3
			}
		}
		rows, _ := db.Query(fmt.Sprintf(`SELECT name, tower_mission_%d FROM guild_characters gc INNER JOIN characters c ON gc.character_id = c.id WHERE guild_id=$1 AND tower_mission_%d IS NOT NULL ORDER BY tower_mission_%d DESC`, pkt.Unk3, pkt.Unk3, pkt.Unk3), pkt.GuildID)
		for rows.Next() {
			temp := model.TenrouiraiCharScore{}
			rows.Scan(&temp.Name, &temp.Score)
			tenrouirai.CharScore = append(tenrouirai.CharScore, temp)
		}
		for _, charScore := range tenrouirai.CharScore {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(charScore.Score)
			bf.WriteBytes(stringsupport.PaddedString(charScore.Name, 14, true))
			data = append(data, bf)
		}
	case 6:
		db.QueryRow(`SELECT tower_rp FROM guilds WHERE id=$1`, pkt.GuildID).Scan(&tenrouirai.Ticket[0].RP)
		for _, ticket := range tenrouirai.Ticket {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(ticket.Unk0)
			bf.WriteUint32(ticket.RP)
			bf.WriteUint32(ticket.Unk2)
			data = append(data, bf)
		}
	}

	s.DoAckEarthSucceed(pkt.AckHandle, data)
}

func handleMsgMhfPostTenrouirai(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTenrouirai)

	if config.GetConfig().DebugOptions.QuestTools {
		s.Logger.Debug(
			p.Opcode().String(),
			zap.Uint8("Unk0", pkt.Unk0),
			zap.Uint8("Op", pkt.Op),
			zap.Uint32("GuildID", pkt.GuildID),
			zap.Uint8("Unk1", pkt.Unk1),
			zap.Uint16("Floors", pkt.Floors),
			zap.Uint16("Antiques", pkt.Antiques),
			zap.Uint16("Chests", pkt.Chests),
			zap.Uint16("Cats", pkt.Cats),
			zap.Uint16("TRP", pkt.TRP),
			zap.Uint16("Slays", pkt.Slays),
		)
	}

	if pkt.Op == 2 {
		var page, requirement, donated int
		db.QueryRow(`SELECT tower_mission_page, tower_rp FROM guilds WHERE id=$1`, pkt.GuildID).Scan(&page, &donated)

		for i := 0; i < (page*3)+1; i++ {
			requirement += int(tenrouiraiData[i].Cost)
		}

		bf := byteframe.NewByteFrame()

		sd, err := GetCharacterSaveData(s, s.CharID)
		if err == nil && sd != nil {
			sd.RP -= pkt.DonatedRP
			sd.Save(s)
			if donated+int(pkt.DonatedRP) >= requirement {
				db.Exec(`UPDATE guilds SET tower_mission_page=tower_mission_page+1 WHERE id=$1`, pkt.GuildID)
				db.Exec(`UPDATE guild_characters SET tower_mission_1=NULL, tower_mission_2=NULL, tower_mission_3=NULL WHERE guild_id=$1`, pkt.GuildID)
				pkt.DonatedRP = uint16(requirement - donated)
			}
			bf.WriteUint32(uint32(pkt.DonatedRP))
			db.Exec(`UPDATE guilds SET tower_rp=tower_rp+$1 WHERE id=$2`, pkt.DonatedRP, pkt.GuildID)
		} else {
			bf.WriteUint32(0)
		}

		s.DoAckSimpleSucceed(pkt.AckHandle, bf.Data())
	} else {
		s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgMhfPresentBox(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPresentBox)
	var data []*byteframe.ByteFrame
	/*
		bf.WriteUint32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
	*/
	s.DoAckEarthSucceed(pkt.AckHandle, data)
}

func handleMsgMhfGetGemInfo(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGemInfo)
	var data []*byteframe.ByteFrame
	gemInfo := []model.GemInfo{}
	gemHistory := []model.GemHistory{}

	var tempGems string
	db.QueryRow(`SELECT COALESCE(gems, $1) FROM tower WHERE char_id=$2`, EmptyTowerCSV(30), s.CharID).Scan(&tempGems)
	for i, v := range stringsupport.CSVElems(tempGems) {
		gemInfo = append(gemInfo, model.GemInfo{uint16((i / 5 << 8) + (i%5 + 1)), uint16(v)})
	}

	switch pkt.Unk0 {
	case 1:
		for _, info := range gemInfo {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(info.Gem)
			bf.WriteUint16(info.Quantity)
			data = append(data, bf)
		}
	case 2:
		for _, history := range gemHistory {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(history.Gem)
			bf.WriteUint16(history.Message)
			bf.WriteUint32(uint32(history.Timestamp.Unix()))
			bf.WriteBytes(stringsupport.PaddedString(history.Sender, 14, true))
			data = append(data, bf)
		}
	}
	s.DoAckEarthSucceed(pkt.AckHandle, data)
}

func handleMsgMhfPostGemInfo(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostGemInfo)

	if config.GetConfig().DebugOptions.QuestTools {
		s.Logger.Debug(
			p.Opcode().String(),
			zap.Uint32("Op", pkt.Op),
			zap.Uint32("Unk1", pkt.Unk1),
			zap.Int32("Gem", pkt.Gem),
			zap.Int32("Quantity", pkt.Quantity),
			zap.Int32("CID", pkt.CID),
			zap.Int32("Message", pkt.Message),
			zap.Int32("Unk6", pkt.Unk6),
		)
	}

	var gems string
	db.QueryRow(`SELECT COALESCE(gems, $1) FROM tower WHERE char_id=$2`, EmptyTowerCSV(30), s.CharID).Scan(&gems)
	switch pkt.Op {
	case 1: // Add gem
		i := int((pkt.Gem >> 8 * 5) + (pkt.Gem - pkt.Gem&0xFF00 - 1%5))
		db.Exec(`UPDATE tower SET gems=$1 WHERE char_id=$2`, stringsupport.CSVSetIndex(gems, i, stringsupport.CSVGetIndex(gems, i)+int(pkt.Quantity)), s.CharID)
	case 2: // Transfer gem
		// no way im doing this for now
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetNotice(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetNotice)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfPostNotice(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostNotice)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}
