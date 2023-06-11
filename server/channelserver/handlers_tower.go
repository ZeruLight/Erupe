package channelserver

import (
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
	switch pkt.InfoType {
	case 2:
		skills := "0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0"
		s.server.db.QueryRow(`SELECT skills FROM tower WHERE char_id=$1`, s.charID).Scan(&skills)
		s.server.db.Exec(`UPDATE tower SET skills=$1, tsp=tsp-$2 WHERE char_id=$3`, stringsupport.CSVSetIndex(skills, int(pkt.Unk2), stringsupport.CSVGetIndex(skills, int(pkt.Unk2))+1), pkt.Unk5, s.charID)
	case 7:
		s.server.db.Exec(`UPDATE tower SET tr=$1, trp=$2, zone1=zone1+$3 WHERE char_id=$4`, pkt.Unk3, pkt.Unk4, pkt.Unk8, s.charID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
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
	// if the game gets bad responses for this it breaks the ability to save
	pkt := p.(*mhfpacket.MsgMhfGetTenrouirai)
	var data []byte
	var err error
	if pkt.Unk0 == 1 {
		data, err = hex.DecodeString("0A218EAD000000000000000000000001010000000000060010")
	} else if pkt.Unk2 == 4 {
		data, err = hex.DecodeString("0A218EAD0000000000000000000000210101005000000202010102020104001000000202010102020106003200000202010002020104000C003202020101020201030032000002020101020202059C4000000202010002020105C35000320202010102020201003C00000202010102020203003200000201010001020203002800320201010101020204000C00000201010101020206002800000201010001020101003C00320201020101020105C35000000301020101020106003200000301020001020104001000320301020101020105C350000003010201010202030028000003010200010201030032003203010201010202059C4000000301020101010206002800000301020001010201003C00320301020101010206003200000301020101010204000C000003010200010101010050003203010201010101059C40000003010201010101030032000003010200010101040010003203010001010101060032000003010001010102030028000003010001010101010050003203010000010102059C4000000301000001010206002800000301000001010010")
	} else {
		data = []byte{0x00, 0x00, 0x00, 0x00}
		s.logger.Info("GET_TENROUIRAI request for unknown type")
	}
	if err != nil {
		panic(err)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostTenrouirai(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTenrouirai)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
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
	Unk0 uint16
	Unk1 uint16
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
	switch pkt.Unk0 {
	case 1:
		for _, history := range gemHistory {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(history.Unk0)
			bf.WriteUint16(history.Unk1)
			bf.WriteUint32(history.Unk2)
			bf.WriteBytes(stringsupport.PaddedString(history.Unk3, 14, true))
			data = append(data, bf)
		}
	default:
		for _, info := range gemInfo {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(info.Unk0)
			bf.WriteUint16(info.Unk1)
			data = append(data, bf)
		}
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostGemInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostGemInfo)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
