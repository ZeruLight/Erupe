package channelserver

import (
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
	Zone1 int32
	Unk1  int32
	Unk2  int32
	Unk3  int32
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
		Level:   []TowerInfoLevel{{0, 0, 0, 0}},
	}

	tempSkills := "0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0"

	err := s.server.db.QueryRow(`SELECT COALESCE(tr, 0),  COALESCE(trp, 0),  COALESCE(tsp, 0), COALESCE(zone1, 0), skills FROM tower WHERE char_id=$1
		`, s.charID).Scan(&towerInfo.TRP[0].TR, &towerInfo.TRP[0].TRP, &towerInfo.Skill[0].TSP, &towerInfo.Level[0].Zone1, &tempSkills)
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
			bf.WriteInt32(level.Zone1)
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
			zap.Int32("Zone1", pkt.Zone1),
			zap.Int64("Unk9", pkt.Unk9),
		)
	}

	switch pkt.InfoType {
	case 2:
		skills := "0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0"
		s.server.db.QueryRow(`SELECT skills FROM tower WHERE char_id=$1`, s.charID).Scan(&skills)
		s.server.db.Exec(`UPDATE tower SET skills=$1, tsp=tsp-$2 WHERE char_id=$3`, stringsupport.CSVSetIndex(skills, int(pkt.Skill), stringsupport.CSVGetIndex(skills, int(pkt.Skill))+1), pkt.Cost, s.charID)
	case 7:
		s.server.db.Exec(`UPDATE tower SET tr=$1, trp=trp+$2, zone1=zone1+$3 WHERE char_id=$4`, pkt.TR, pkt.TRP, pkt.Zone1, s.charID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

type TenrouiraiCharScore struct {
	Score int32
	Name  string
}

type TenrouiraiProgress struct {
	Unk0 uint8
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
}

type TenrouiraiTicket struct {
	Unk0 uint8
	Unk1 uint32
	Unk2 uint32
}

type TenrouiraiData struct {
	Unk0 uint8
	Unk1 uint8
	Unk2 uint16
	Unk3 uint16
	Unk4 uint8
	Unk5 uint8
	Unk6 uint8
	Unk7 uint8
	Unk8 uint8
	Unk9 uint8
}

type Tenrouirai struct {
	CharScore []TenrouiraiCharScore
	Progress  []TenrouiraiProgress
	Ticket    []TenrouiraiTicket
	Data      []TenrouiraiData
}

func handleMsgMhfGetTenrouirai(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTenrouirai)
	var data []*byteframe.ByteFrame

	tenrouirai := Tenrouirai{
		Data: []TenrouiraiData{
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
		},
	}

	switch pkt.Unk1 {
	case 4:
		for _, tdata := range tenrouirai.Data {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(tdata.Unk0)
			bf.WriteUint8(tdata.Unk1)
			bf.WriteUint16(tdata.Unk2)
			bf.WriteUint16(tdata.Unk3)
			bf.WriteUint8(tdata.Unk4)
			bf.WriteUint8(tdata.Unk5)
			bf.WriteUint8(tdata.Unk6)
			bf.WriteUint8(tdata.Unk7)
			bf.WriteUint8(tdata.Unk8)
			bf.WriteUint8(tdata.Unk9)
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
			zap.Uint8("Unk1", pkt.Unk1),
			zap.Uint32("GuildID", pkt.GuildID),
			zap.Uint8("Unk3", pkt.Unk3),
			zap.Uint16("Unk4", pkt.Unk4),
			zap.Uint16("Unk5", pkt.Unk5),
			zap.Uint16("Unk6", pkt.Unk6),
			zap.Uint16("Unk7", pkt.Unk7),
			zap.Uint16("Unk8", pkt.Unk8),
			zap.Uint16("Unk9", pkt.Unk9),
		)
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetBreakSeibatuLevelReward(s *Session, p mhfpacket.MHFPacket) {}

type WeeklySeibatuRankingReward struct {
	Unk0 int32
	Unk1 int32
	Unk2 uint32
	Unk3 int32
	Unk4 int32
	Unk5 int32
}

func handleMsgMhfGetWeeklySeibatuRankingReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySeibatuRankingReward)
	var data []*byteframe.ByteFrame
	weeklySeibatuRankingRewards := []WeeklySeibatuRankingReward{
		{0, 0, 0, 0, 0, 0},
	}
	for _, reward := range weeklySeibatuRankingRewards {
		bf := byteframe.NewByteFrame()
		bf.WriteInt32(reward.Unk0)
		bf.WriteInt32(reward.Unk1)
		bf.WriteUint32(reward.Unk2)
		bf.WriteInt32(reward.Unk3)
		bf.WriteInt32(reward.Unk4)
		bf.WriteInt32(reward.Unk5)
		data = append(data, bf)
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
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

	tempGems := "0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0"
	s.server.db.QueryRow(`SELECT gems FROM tower WHERE char_id=$1`, s.charID).Scan(&tempGems)
	for i, v := range stringsupport.CSVElems(tempGems) {
		gemInfo = append(gemInfo, GemInfo{uint16(((i / 3) * 256) + ((i % 3) + 1)), uint16(v)})
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

	gems := "0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0"
	s.server.db.QueryRow(`SELECT gems FROM tower WHERE char_id=$1`, s.charID).Scan(&gems)
	switch pkt.Op {
	case 1: // Add gem
		i := int(((pkt.Gem / 256) * 3) + (((pkt.Gem - ((pkt.Gem / 256) * 256)) - 1) % 3))
		s.server.db.Exec(`UPDATE tower SET gems=$1 WHERE char_id=$2`, stringsupport.CSVSetIndex(gems, i, stringsupport.CSVGetIndex(gems, i)+int(pkt.Quantity)), s.charID)
	case 2: // Transfer gem
		// no way im doing this for now
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
