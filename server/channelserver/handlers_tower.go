package channelserver

import (
	"fmt"
	"go.uber.org/zap"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
)

type TowerInfoTRP struct {
	TR  int32
	TRP int32
}

type TowerInfoSkill struct {
	TSP  int32
	Unk1 []int16 // 40
}

type TowerInfoHistory struct {
	Unk0 []int16 // 5
	Unk1 []int16 // 5
}

type TowerInfoLevel struct {
	Floors int32
	Unk1   int32
	Unk2   int32
	Unk3   int32
}

func handleMsgMhfGetTowerInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTowerInfo)
	var data []*byteframe.ByteFrame
	type TowerInfo struct {
		TRP     []TowerInfoTRP
		Skill   []TowerInfoSkill
		History []TowerInfoHistory
		Level   []TowerInfoLevel
	}

	towerInfo := TowerInfo{
		TRP:     []TowerInfoTRP{{0, 0}},
		Skill:   []TowerInfoSkill{{0, make([]int16, 40)}},
		History: []TowerInfoHistory{{make([]int16, 5), make([]int16, 5)}},
		Level:   []TowerInfoLevel{{0, 0, 0, 0}, {0, 0, 0, 0}},
	}

	tempSkills := "0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0"

	err := s.server.db.QueryRow(`SELECT COALESCE(tr, 0),  COALESCE(trp, 0),  COALESCE(tsp, 0), COALESCE(block1, 0), COALESCE(block2, 0), skills FROM tower WHERE char_id=$1
		`, s.charID).Scan(&towerInfo.TRP[0].TR, &towerInfo.TRP[0].TRP, &towerInfo.Skill[0].TSP, &towerInfo.Level[0].Floors, &towerInfo.Level[1].Floors, &tempSkills)
	if err != nil {
		s.server.db.Exec(`INSERT INTO tower (char_id) VALUES ($1)`, s.charID)
	}

	for i, skill := range stringsupport.CSVElems(tempSkills) {
		towerInfo.Skill[0].Unk1[i] = int16(skill)
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
			for i := range skills.Unk1 {
				bf.WriteInt16(skills.Unk1[i])
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
	case 5:
		for _, level := range towerInfo.Level {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(level.Floors)
			bf.WriteInt32(level.Unk1)
			bf.WriteInt32(level.Unk2)
			bf.WriteInt32(level.Unk3)
			data = append(data, bf)
		}
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostTowerInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTowerInfo)

	if s.server.erupeConfig.DevModeOptions.QuestDebugTools {
		s.logger.Debug(
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
		skills := "0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0"
		s.server.db.QueryRow(`SELECT skills FROM tower WHERE char_id=$1`, s.charID).Scan(&skills)
		s.server.db.Exec(`UPDATE tower SET skills=$1, tsp=tsp-$2 WHERE char_id=$3`, stringsupport.CSVSetIndex(skills, int(pkt.Skill), stringsupport.CSVGetIndex(skills, int(pkt.Skill))+1), pkt.Cost, s.charID)
	case 7:
		s.server.db.Exec(`UPDATE tower SET tr=$1, trp=trp+$2, block1=block1+$3 WHERE char_id=$4`, pkt.TR, pkt.TRP, pkt.Block1, s.charID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

// Default missions
var tenrouiraiData = []TenrouiraiData{
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

type TenrouiraiProgress struct {
	Page     uint8
	Mission1 uint16
	Mission2 uint16
	Mission3 uint16
}

type TenrouiraiReward struct {
	Index    uint8
	Item     []uint16 // 5
	Quantity []uint8  // 5
}

type TenrouiraiKeyScore struct {
	Unk0 uint8
	Unk1 int32
}

type TenrouiraiData struct {
	Block   uint8
	Mission uint8
	// 1 = Floors climbed
	// 2 = Collect antiques
	// 3 = Open chests
	// 4 = Cats saved
	// 5 = TRP acquisition
	// 6 = Monster slays
	Goal   uint16
	Cost   uint16
	Skill1 uint8 // 80
	Skill2 uint8 // 40
	Skill3 uint8 // 40
	Skill4 uint8 // 20
	Skill5 uint8 // 40
	Skill6 uint8 // 50
}

type TenrouiraiCharScore struct {
	Score int32
	Name  string
}

type TenrouiraiTicket struct {
	Unk0 uint8
	RP   uint32
	Unk2 uint32
}

type Tenrouirai struct {
	Progress  []TenrouiraiProgress
	Reward    []TenrouiraiReward
	KeyScore  []TenrouiraiKeyScore
	Data      []TenrouiraiData
	CharScore []TenrouiraiCharScore
	Ticket    []TenrouiraiTicket
}

func handleMsgMhfGetTenrouirai(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTenrouirai)
	var data []*byteframe.ByteFrame

	tenrouirai := Tenrouirai{
		Progress: []TenrouiraiProgress{{1, 0, 0, 0}},
		Data:     tenrouiraiData,
		Ticket:   []TenrouiraiTicket{{0, 0, 0}},
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
		s.server.db.QueryRow(`SELECT tower_mission_page FROM guilds WHERE id=$1`, pkt.GuildID).Scan(&tenrouirai.Progress[0].Page)
		s.server.db.QueryRow(`SELECT SUM(tower_mission_1) AS _, SUM(tower_mission_2) AS _, SUM(tower_mission_3) AS _ FROM guild_characters WHERE guild_id=$1
				`, pkt.GuildID).Scan(&tenrouirai.Progress[0].Mission1, &tenrouirai.Progress[0].Mission2, &tenrouirai.Progress[0].Mission3)

		if tenrouirai.Progress[0].Mission1 > tenrouiraiData[(tenrouirai.Progress[0].Page*3)-3].Goal {
			tenrouirai.Progress[0].Mission1 = tenrouiraiData[(tenrouirai.Progress[0].Page*3)-3].Goal
		}
		if tenrouirai.Progress[0].Mission2 > tenrouiraiData[(tenrouirai.Progress[0].Page*3)-2].Goal {
			tenrouirai.Progress[0].Mission2 = tenrouiraiData[(tenrouirai.Progress[0].Page*3)-2].Goal
		}
		if tenrouirai.Progress[0].Mission1 > tenrouiraiData[(tenrouirai.Progress[0].Page*3)-1].Goal {
			tenrouirai.Progress[0].Mission1 = tenrouiraiData[(tenrouirai.Progress[0].Page*3)-1].Goal
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
		rows, _ := s.server.db.Query(fmt.Sprintf(`SELECT name, tower_mission_%d FROM guild_characters gc INNER JOIN characters c ON gc.character_id = c.id WHERE guild_id=$1 AND tower_mission_%d IS NOT NULL ORDER BY tower_mission_%d DESC`, pkt.Unk3, pkt.Unk3, pkt.Unk3), pkt.GuildID)
		for rows.Next() {
			temp := TenrouiraiCharScore{}
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
		for _, ticket := range tenrouirai.Ticket {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(ticket.Unk0)
			bf.WriteUint32(ticket.RP)
			bf.WriteUint32(ticket.Unk2)
			data = append(data, bf)
		}
	}

	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostTenrouirai(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTenrouirai)

	if s.server.erupeConfig.DevModeOptions.QuestDebugTools {
		s.logger.Debug(
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

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfPresentBox(s *Session, p mhfpacket.MHFPacket) {
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
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

type GemInfo struct {
	Gem      uint16
	Quantity uint16
}

type GemHistory struct {
	Unk0 uint16
	Unk1 uint16
	Unk2 uint32
	Unk3 string
}

func handleMsgMhfGetGemInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGemInfo)
	var data []*byteframe.ByteFrame
	gemInfo := []GemInfo{}
	gemHistory := []GemHistory{}

	tempGems := "0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0"
	s.server.db.QueryRow(`SELECT gems FROM tower WHERE char_id=$1`, s.charID).Scan(&tempGems)
	for i, v := range stringsupport.CSVElems(tempGems) {
		gemInfo = append(gemInfo, GemInfo{uint16(((i / 5) * 256) + ((i % 5) + 1)), uint16(v)})
	}

	switch pkt.Unk0 {
	case 1:
		for _, info := range gemInfo {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(info.Gem)
			bf.WriteUint16(info.Quantity)
			data = append(data, bf)
		}
	default:
		for _, history := range gemHistory {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(history.Unk0)
			bf.WriteUint16(history.Unk1)
			bf.WriteUint32(history.Unk2)
			bf.WriteBytes(stringsupport.PaddedString(history.Unk3, 14, true))
			data = append(data, bf)
		}
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostGemInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostGemInfo)

	if s.server.erupeConfig.DevModeOptions.QuestDebugTools {
		s.logger.Debug(
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

	gems := "0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0"
	s.server.db.QueryRow(`SELECT gems FROM tower WHERE char_id=$1`, s.charID).Scan(&gems)
	switch pkt.Op {
	case 1: // Add gem
		i := int(((pkt.Gem / 256) * 5) + (((pkt.Gem - ((pkt.Gem / 256) * 256)) - 1) % 5))
		s.server.db.Exec(`UPDATE tower SET gems=$1 WHERE char_id=$2`, stringsupport.CSVSetIndex(gems, i, stringsupport.CSVGetIndex(gems, i)+int(pkt.Quantity)), s.charID)
	case 2: // Transfer gem
		// no way im doing this for now
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
