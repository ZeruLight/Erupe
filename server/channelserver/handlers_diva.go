package channelserver

import (
	"encoding/hex"
	"erupe-ce/common/stringsupport"
	"golang.org/x/exp/slices"
	"math/rand"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func cleanupDiva(s *Session) {
	s.server.db.Exec("DELETE FROM events WHERE event_type='diva'")
}

func generateDivaTimestamps(s *Session, start uint32, debug bool) []uint32 {
	timestamps := make([]uint32, 6)
	midnight := Time_Current_Midnight()
	if debug && start <= 3 {
		midnight := uint32(midnight.Unix())
		switch start {
		case 1:
			timestamps[0] = midnight
			timestamps[1] = timestamps[0] + 601200
			timestamps[2] = timestamps[1] + 3900
			timestamps[3] = timestamps[1] + 604800
			timestamps[4] = timestamps[3] + 3900
			timestamps[5] = timestamps[3] + 604800
		case 2:
			timestamps[0] = midnight - 605100
			timestamps[1] = midnight - 3900
			timestamps[2] = midnight
			timestamps[3] = timestamps[1] + 604800
			timestamps[4] = timestamps[3] + 3900
			timestamps[5] = timestamps[3] + 604800
		case 3:
			timestamps[0] = midnight - 1213800
			timestamps[1] = midnight - 608700
			timestamps[2] = midnight - 604800
			timestamps[3] = midnight - 3900
			timestamps[4] = midnight
			timestamps[5] = timestamps[3] + 604800
		}
		return timestamps
	}
	if start == 0 || Time_Current_Adjusted().Unix() > int64(start)+2977200 {
		cleanupDiva(s)
		// Generate a new diva defense, starting midnight tomorrow
		start = uint32(midnight.Add(24 * time.Hour).Unix())
		s.server.db.Exec("INSERT INTO events (event_type, start_time) VALUES ('diva', to_timestamp($1)::timestamp without time zone)", start)
	}
	timestamps[0] = start
	timestamps[1] = timestamps[0] + 601200
	timestamps[2] = timestamps[1] + 3900
	timestamps[3] = timestamps[1] + 604800
	timestamps[4] = timestamps[3] + 3900
	timestamps[5] = timestamps[3] + 604800
	return timestamps
}

func handleMsgMhfGetUdSchedule(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdSchedule)
	bf := byteframe.NewByteFrame()

	id, start := uint32(0xCAFEBEEF), uint32(0)
	rows, _ := s.server.db.Queryx("SELECT id, (EXTRACT(epoch FROM start_time)::int) as start_time FROM events WHERE event_type='diva'")
	for rows.Next() {
		rows.Scan(&id, &start)
	}

	var timestamps []uint32
	if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.DivaEvent >= 0 {
		if s.server.erupeConfig.DevModeOptions.DivaEvent == 0 {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 36))
			return
		}
		timestamps = generateDivaTimestamps(s, uint32(s.server.erupeConfig.DevModeOptions.DivaEvent), true)
	} else {
		timestamps = generateDivaTimestamps(s, start, false)
	}

	bf.WriteUint32(id)
	for _, timestamp := range timestamps {
		bf.WriteUint32(timestamp)
	}

	bf.WriteUint16(0x19) // Unk 00011001
	bf.WriteUint16(0x2D) // Unk 00101101
	bf.WriteUint16(0x02) // Unk 00000010
	bf.WriteUint16(0x02) // Unk 00000010

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdInfo)
	// Message that appears on the Diva Defense NPC and triggers the green exclamation mark
	udInfos := []struct {
		Text      string
		StartTime time.Time
		EndTime   time.Time
	}{}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(udInfos)))
	for _, udInfo := range udInfos {
		resp.WriteBytes(stringsupport.PaddedString(udInfo.Text, 1024, true))
		resp.WriteUint32(uint32(udInfo.StartTime.Unix()))
		resp.WriteUint32(uint32(udInfo.EndTime.Unix()))
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func getKijuStrings(effectID uint8) (string, string) {
	switch effectID {
	case 1:
		return "暴風の祈珠", "ーあらしまかぜのきじゅー\n暴風とは猛る思い。\n聞く者に勇気を与える。"
	case 3:
		return "断力の祈珠", "ーだんりきのきじゅー\n断力とは断ち切る思い。\n聴く者に新たな利からを授ける。"
	case 4:
		return "風韻の祈珠", "ーふういんのきじゅー\n風韻とは歌姫の艶。\n時々で異なる趣を醸し出す。"
	case 8:
		return "斬刃の祈珠", "ーざんばのきじゅー\n斬刃とはすべてを切り裂く力。\n集めるほどに声の透明感は増す。"
	case 9:
		return "打明の祈珠", "ーうちあかりのきじゅー\n打明とは熱い力。\n聴く者に活力を与える。"
	case 10:
		return "弾起の祈珠", "ーたまおこしのきじゅー\n弾起とは悠遠の記憶。\n聴く者に更なる力を授ける。"
	case 11:
		return "変続の祈珠", "ーへんぞくのきじゅー\n変続とは永久の言葉。\n聴く者に継続力を授ける。"
	case 14:
		return "万雷の祈珠", "ーばんらいのきじゅー\n万雷とは歌姫に集う民の意識。\n歌姫の声を伝播させる。"
	case 15:
		return "不動の祈珠", "ーうごかずのきじゅー\n不動とは圧力。聞く者に圧倒する力を与える。"
	case 16:
		return "鏗鏗の祈珠", "ーこうこうのきじゅー\n鏗鏗とは歌姫の声。\n集めるほどに歌姫の声量は増す。"
	case 17:
		return "結集の祈珠", "ーけっしゅうのきじゅー\n結集とは確固たる信頼。\n集めるほどに狩人たちの精神力となる。"
	case 18:
		return "歌護の祈珠", "ーうたまもりのきじゅー\n歌護とは歌姫の護り。\n集めるほどに狩人たちの支えとなる。"
	case 19:
		return "強撃の祈珠", "ーきょうげきのきじゅー\n強撃とは強い声色。\n聞く者の力を研ぎ澄ます。"
	case 20:
		return "封火の祈珠", "ーふうかのきじゅー"
	case 21:
		return "封水の祈珠", "ーふうすいのきじゅー"
	case 22:
		return "封氷の祈珠", "ーふうひょうのきじゅー"
	case 23:
		return "封龍の祈珠", "ーふうりゅうのきじゅー"
	case 24:
		return "封雷の祈珠", "ーふうらいのきじゅー"
	case 25:
		return "封属の祈珠", "ーふうぞくのきじゅー"
	}
	return "Unknown", ""
}

func handleMsgMhfGetKijuInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKijuInfo)
	kijuInfo := []struct {
		Color  uint8
		Effect uint8
	}{
		{1, 1},
		{2, 3},
		{3, 4},
		{4, 8},
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(kijuInfo)))
	for _, kiju := range kijuInfo {
		name, description := getKijuStrings(kiju.Effect)
		bf.WriteBytes(stringsupport.PaddedString(name, 32, true))
		bf.WriteBytes(stringsupport.PaddedString(description, 512, true))
		bf.WriteUint8(kiju.Color)
		bf.WriteUint8(kiju.Effect)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfSetKiju(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetKiju)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfAddUdPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddUdPoint)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetUdMyPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdMyPoint)
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 145))
}

func handleMsgMhfGetUdTotalPointInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdTotalPointInfo)
	// Temporary canned response
	data, _ := hex.DecodeString("00000000000007A12000000000000F424000000000001E848000000000002DC6C000000000003D090000000000004C4B4000000000005B8D8000000000006ACFC000000000007A1200000000000089544000000000009896800000000000E4E1C00000000001312D0000000000017D78400000000001C9C3800000000002160EC00000000002625A000000000002AEA5400000000002FAF0800000000003473BC0000000000393870000000000042C1D800000000004C4B40000000000055D4A800000000005F5E10000000000008954400000000001C9C3800000000003473BC00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001020300000000000000000000000000000000000000000000000000000000000000000000000000000000101F1420")
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfGetUdSelectedColorInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdSelectedColorInfo)

	// Unk
	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x01, 0x01, 0x01, 0x02, 0x03, 0x02, 0x00, 0x00})
}

func handleMsgMhfGetUdMonsterPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdMonsterPoint)

	monsterPoints := []struct {
		MID    uint8
		Points uint16
	}{
		{MID: 0x01, Points: 0x3C}, // em1 Rathian
		{MID: 0x02, Points: 0x5A}, // em2 Fatalis
		{MID: 0x06, Points: 0x14}, // em6 Yian Kut-Ku
		{MID: 0x07, Points: 0x50}, // em7 Lao-Shan Lung
		{MID: 0x08, Points: 0x28}, // em8 Cephadrome
		{MID: 0x0B, Points: 0x3C}, // em11 Rathalos
		{MID: 0x0E, Points: 0x3C}, // em14 Diablos
		{MID: 0x0F, Points: 0x46}, // em15 Khezu
		{MID: 0x11, Points: 0x46}, // em17 Gravios
		{MID: 0x14, Points: 0x28}, // em20 Gypceros
		{MID: 0x15, Points: 0x3C}, // em21 Plesioth
		{MID: 0x16, Points: 0x32}, // em22 Basarios
		{MID: 0x1A, Points: 0x32}, // em26 Monoblos
		{MID: 0x1B, Points: 0x0A}, // em27 Velocidrome
		{MID: 0x1C, Points: 0x0A}, // em28 Gendrome
		{MID: 0x1F, Points: 0x0A}, // em31 Iodrome
		{MID: 0x21, Points: 0x50}, // em33 Kirin
		{MID: 0x24, Points: 0x64}, // em36 Crimson Fatalis
		{MID: 0x25, Points: 0x3C}, // em37 Pink Rathian
		{MID: 0x26, Points: 0x1E}, // em38 Blue Yian Kut-Ku
		{MID: 0x27, Points: 0x28}, // em39 Purple Gypceros
		{MID: 0x28, Points: 0x50}, // em40 Yian Garuga
		{MID: 0x29, Points: 0x5A}, // em41 Silver Rathalos
		{MID: 0x2A, Points: 0x50}, // em42 Gold Rathian
		{MID: 0x2B, Points: 0x3C}, // em43 Black Diablos
		{MID: 0x2C, Points: 0x3C}, // em44 White Monoblos
		{MID: 0x2D, Points: 0x46}, // em45 Red Khezu
		{MID: 0x2E, Points: 0x3C}, // em46 Green Plesioth
		{MID: 0x2F, Points: 0x50}, // em47 Black Gravios
		{MID: 0x30, Points: 0x1E}, // em48 Daimyo Hermitaur
		{MID: 0x31, Points: 0x3C}, // em49 Azure Rathalos
		{MID: 0x32, Points: 0x50}, // em50 Ashen Lao-Shan Lung
		{MID: 0x33, Points: 0x3C}, // em51 Blangonga
		{MID: 0x34, Points: 0x28}, // em52 Congalala
		{MID: 0x35, Points: 0x50}, // em53 Rajang
		{MID: 0x36, Points: 0x6E}, // em54 Kushala Daora
		{MID: 0x37, Points: 0x50}, // em55 Shen Gaoren
		{MID: 0x3A, Points: 0x50}, // em58 Yama Tsukami
		{MID: 0x3B, Points: 0x6E}, // em59 Chameleos
		{MID: 0x40, Points: 0x64}, // em64 Lunastra
		{MID: 0x41, Points: 0x6E}, // em65 Teostra
		{MID: 0x43, Points: 0x28}, // em67 Shogun Ceanataur
		{MID: 0x44, Points: 0x0A}, // em68 Bulldrome
		{MID: 0x47, Points: 0x6E}, // em71 White Fatalis
		{MID: 0x4A, Points: 0xFA}, // em74 Hypnocatrice
		{MID: 0x4B, Points: 0xFA}, // em75 Lavasioth
		{MID: 0x4C, Points: 0x46}, // em76 Tigrex
		{MID: 0x4D, Points: 0x64}, // em77 Akantor
		{MID: 0x4E, Points: 0xFA}, // em78 Bright Hypnoc
		{MID: 0x4F, Points: 0xFA}, // em79 Lavasioth Subspecies
		{MID: 0x50, Points: 0xFA}, // em80 Espinas
		{MID: 0x51, Points: 0xFA}, // em81 Orange Espinas
		{MID: 0x52, Points: 0xFA}, // em82 White Hypnoc
		{MID: 0x53, Points: 0xFA}, // em83 Akura Vashimu
		{MID: 0x54, Points: 0xFA}, // em84 Akura Jebia
		{MID: 0x55, Points: 0xFA}, // em85 Berukyurosu
		{MID: 0x59, Points: 0xFA}, // em89 Pariapuria
		{MID: 0x5A, Points: 0xFA}, // em90 White Espinas
		{MID: 0x5B, Points: 0xFA}, // em91 Kamu Orugaron
		{MID: 0x5C, Points: 0xFA}, // em92 Nono Orugaron
		{MID: 0x5E, Points: 0xFA}, // em94 Dyuragaua
		{MID: 0x5F, Points: 0xFA}, // em95 Doragyurosu
		{MID: 0x60, Points: 0xFA}, // em96 Gurenzeburu
		{MID: 0x63, Points: 0xFA}, // em99 Rukodiora
		{MID: 0x65, Points: 0xFA}, // em101 Gogomoa
		{MID: 0x67, Points: 0xFA}, // em103 Taikun Zamuza
		{MID: 0x68, Points: 0xFA}, // em104 Abiorugu
		{MID: 0x69, Points: 0xFA}, // em105 Kuarusepusu
		{MID: 0x6A, Points: 0xFA}, // em106 Odibatorasu
		{MID: 0x6B, Points: 0xFA}, // em107 Disufiroa
		{MID: 0x6C, Points: 0xFA}, // em108 Rebidiora
		{MID: 0x6D, Points: 0xFA}, // em109 Anorupatisu
		{MID: 0x6E, Points: 0xFA}, // em110 Hyujikiki
		{MID: 0x6F, Points: 0xFA}, // em111 Midogaron
		{MID: 0x70, Points: 0xFA}, // em112 Giaorugu
		{MID: 0x72, Points: 0xFA}, // em114 Farunokku
		{MID: 0x73, Points: 0xFA}, // em115 Pokaradon
		{MID: 0x74, Points: 0xFA}, // em116 Shantien
		{MID: 0x77, Points: 0xFA}, // em119 Goruganosu
		{MID: 0x78, Points: 0xFA}, // em120 Aruganosu
		{MID: 0x79, Points: 0xFA}, // em121 Baruragaru
		{MID: 0x7A, Points: 0xFA}, // em122 Zerureusu
		{MID: 0x7B, Points: 0xFA}, // em123 Gougarf
		{MID: 0x7D, Points: 0xFA}, // em125 Forokururu
		{MID: 0x7E, Points: 0xFA}, // em126 Meraginasu
		{MID: 0x7F, Points: 0xFA}, // em127 Diorekkusu
		{MID: 0x80, Points: 0xFA}, // em128 Garuba Daora
		{MID: 0x81, Points: 0xFA}, // em129 Inagami
		{MID: 0x82, Points: 0xFA}, // em130 Varusaburosu
		{MID: 0x83, Points: 0xFA}, // em131 Poborubarumu
		{MID: 0x8B, Points: 0xFA}, // em139 Gureadomosu
		{MID: 0x8C, Points: 0xFA}, // em140 Harudomerugu
		{MID: 0x8D, Points: 0xFA}, // em141 Toridcless
		{MID: 0x8E, Points: 0xFA}, // em142 Gasurabazura
		{MID: 0x90, Points: 0xFA}, // em144 Yama Kurai
		{MID: 0x92, Points: 0x78}, // em146 Zinogre
		{MID: 0x93, Points: 0x78}, // em147 Deviljho
		{MID: 0x94, Points: 0x78}, // em148 Brachydios
		{MID: 0x96, Points: 0xFA}, // em150 Toa Tesukatora
		{MID: 0x97, Points: 0x78}, // em151 Barioth
		{MID: 0x98, Points: 0x78}, // em152 Uragaan
		{MID: 0x99, Points: 0x78}, // em153 Stygian Zinogre
		{MID: 0x9A, Points: 0xFA}, // em154 Guanzorumu
		{MID: 0x9E, Points: 0xFA}, // em158 Voljang
		{MID: 0x9F, Points: 0x78}, // em159 Nargacuga
		{MID: 0xA0, Points: 0xFA}, // em160 Keoaruboru
		{MID: 0xA1, Points: 0xFA}, // em161 Zenaserisu
		{MID: 0xA2, Points: 0x78}, // em162 Gore Magala
		{MID: 0xA4, Points: 0x78}, // em164 Shagaru Magala
		{MID: 0xA5, Points: 0x78}, // em165 Amatsu
		{MID: 0xA6, Points: 0xFA}, // em166 Elzelion
		{MID: 0xA9, Points: 0x78}, // em169 Seregios
		{MID: 0xAA, Points: 0xFA}, // em170 Bogabadorumu
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint8(uint8(len(monsterPoints)))
	for _, mp := range monsterPoints {
		resp.WriteUint8(mp.MID)
		resp.WriteUint16(mp.Points)
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetUdDailyPresentList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdDailyPresentList)
	// Temporary canned response
	data, _ := hex.DecodeString("0100001600000A5397DF00000000000000000000000000000000")
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfGetUdNormaPresentList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdNormaPresentList)
	// Temporary canned response
	data, _ := hex.DecodeString("0100001600000A5397DF00000000000000000000000000000000")
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfAcquireUdItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireUdItem)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGetUdRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdRanking)
	bf := byteframe.NewByteFrame()
	// Temporary
	for i := 0; i < 100; i++ {
		bf.WriteUint16(uint16(i + 1))
		stringsupport.PaddedString("", 25, false)
		bf.WriteUint32(0)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdMyRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdMyRanking)
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

type Tile struct {
	ID          uint16
	NextID      uint16
	BranchID    uint16
	QuestFile   uint16
	Unk0        uint32
	BranchIndex uint8
	Type        uint8
	PointsReq   int32
	Claimed     bool
	Unk1        uint8
	Unk2        uint32
}

type MapData struct {
	ID     uint32
	NextID uint32
	Tiles  []Tile
}

type MapProg struct {
	ID    uint32
	Unk   uint16
	Tiles []Tile
	Bytes *byteframe.ByteFrame
}

func handleMsgMhfGetUdGuildMapInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdGuildMapInfo)

	// rudimentary example
	interceptionPoints := map[uint16]int32{0: 200000, 58079: 50}
	var guildProg []MapProg
	mapData := GenerateUdGuildMaps()

	unkData := []struct {
		Unk0 uint32
		Unk1 uint8
		Unk2 uint8
		Unk3 uint8
		Unk4 uint16
		Unk5 uint16
		Unk6 uint16
		Unk7 uint8
	}{}

	bf := byteframe.NewByteFrame()

	bf.WriteUint16(uint16(len(mapData)))
	for _, _map := range mapData {
		guildProg = append(guildProg, MapProg{ID: _map.ID, Unk: 1, Tiles: _map.Tiles})
		bf.WriteUint32(_map.ID)
		bf.WriteUint32(_map.NextID)
		for _, tile := range _map.Tiles {
			bf.WriteUint16(tile.ID)
			bf.WriteUint16(tile.NextID)
			bf.WriteUint16(tile.BranchID)
			bf.WriteUint16(tile.QuestFile)
			bf.WriteUint32(tile.Unk0)
			bf.WriteUint8(tile.BranchIndex)
			bf.WriteUint8(tile.Type)
			bf.WriteInt32(tile.PointsReq)

			bf.WriteUint8(tile.Unk1)
			bf.WriteUint32(tile.Unk2)
		}
		bf.WriteBytes(make([]byte, 23*(64-len(_map.Tiles)))) // Fill out 64 tiles
	}

	bf.WriteUint16(uint16(len(unkData)))
	for _, unk := range unkData {
		bf.WriteUint32(unk.Unk0)
		bf.WriteUint8(unk.Unk1)
		bf.WriteUint8(unk.Unk2)
		bf.WriteUint8(unk.Unk3)
		bf.WriteUint16(unk.Unk4)
		bf.WriteUint16(unk.Unk5)
		bf.WriteUint16(unk.Unk6)
		bf.WriteUint8(unk.Unk7)
	}

	var tilesClaimed uint32
	var currentMapID uint32
	var prevMapID uint32

	for i, prog := range guildProg {
		guildProg[i].Bytes = byteframe.NewByteFrame()
		guildProg[i].Bytes.WriteUint32(prog.ID)
		guildProg[i].Bytes.WriteUint16(prog.Unk)
		guildProg[i].Bytes.WriteUint8(uint8(len(prog.Tiles)))
		for _, tile := range prog.Tiles {
			if tile.Type != 1 && interceptionPoints[tile.QuestFile] > 0 {
				if tile.PointsReq-interceptionPoints[tile.QuestFile] < 0 {
					interceptionPoints[tile.QuestFile] -= tile.PointsReq
					guildProg[i].Bytes.WriteInt32(tile.PointsReq)
					tilesClaimed++
					tile.Claimed = true
				} else {
					if tile.QuestFile == 0 {
						currentMapID = prog.ID
						if i > 0 {
							prevMapID = guildProg[i-1].ID
						}
					}
					guildProg[i].Bytes.WriteInt32(interceptionPoints[tile.QuestFile])
					interceptionPoints[tile.QuestFile] = 0
				}
			} else {
				guildProg[i].Bytes.WriteUint32(0)
			}
			guildProg[i].Bytes.WriteInt32(tile.PointsReq)
			guildProg[i].Bytes.WriteUint16(tile.ID)
			guildProg[i].Bytes.WriteUint16(tile.NextID)
			guildProg[i].Bytes.WriteUint16(tile.BranchID)
			guildProg[i].Bytes.WriteUint16(tile.QuestFile)
			guildProg[i].Bytes.WriteUint32(tile.Unk0)
			guildProg[i].Bytes.WriteUint8(tile.BranchIndex)
			guildProg[i].Bytes.WriteUint8(tile.Type)
			guildProg[i].Bytes.WriteBool(tile.Claimed)
		}
	}

	if prevMapID != 0 {
		bf.WriteUint8(2)
		for _, prog := range guildProg {
			if prog.ID == currentMapID {
				bf.WriteBytes(prog.Bytes.Data())
			}
		}
		for _, prog := range guildProg {
			if prog.ID == prevMapID {
				bf.WriteBytes(prog.Bytes.Data())
			}
		}
	} else {
		bf.WriteUint8(1)
		for _, prog := range guildProg {
			if prog.ID == currentMapID {
				bf.WriteBytes(prog.Bytes.Data())
			}
		}
	}

	bf.WriteUint32(tilesClaimed)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func getNeighbourTiles(tiles [][]uint16, tile uint16) []uint16 {
	var vals []uint16
	var temp []uint16
	if tile%2 == 0 {
		temp = []uint16{tile - 100, tile - 1, tile + 1, tile + 99, tile + 100, tile + 101}
	} else {
		temp = []uint16{tile - 101, tile - 100, tile - 99, tile - 1, tile + 1, tile + 100}
	}

	for _, val := range temp {
		for x := range tiles {
			for y := range tiles[x] {
				if tiles[x][y] == val {
					vals = append(vals, val)
				}
			}
		}
	}
	return vals
}

func GenerateUdGuildMaps() []MapData {
	tiles := make([][]uint16, 5)
	for i := range tiles {
		tiles[i] = make([]uint16, 12)
		for j := range tiles[i] {
			tiles[i][j] = uint16(((i + 1) * 100) + j + 1)
		}
	}

	var guildMaps []MapData

	for i := 0; i < 5; i++ {
		var startTile, endTile uint16
		var randTemp []uint16
		rand.Seed(time.Now().UnixNano())
		randTemp = tiles[rand.Intn(len(tiles))]
		startTile = randTemp[rand.Intn(len(randTemp))]
		for {
			rand.Seed(time.Now().UnixNano())
			randTemp = tiles[rand.Intn(len(tiles))]
			endTile = randTemp[rand.Intn(len(randTemp))]
			invalidTiles := append(getNeighbourTiles(tiles, startTile), startTile)
			if !slices.Contains(invalidTiles, endTile) {
				break
			}
		}

		var tilePath []uint16
		var iterations int
		var tooDifficult bool
		for {
			var pathFailed bool
			var evictedTiles []uint16
			tilePath = []uint16{startTile}
			for {
				var possibleTiles []uint16
				tempTiles := getNeighbourTiles(tiles, tilePath[len(tilePath)-1])
				for _, tile := range tempTiles {
					if !slices.Contains(evictedTiles, tile) {
						possibleTiles = append(possibleTiles, tile)
					}
				}
				if len(possibleTiles) == 0 {
					pathFailed = true
					break
				}
				for _, tile := range possibleTiles {
					evictedTiles = append(evictedTiles, tile)
				}
				newTile := possibleTiles[rand.Intn(len(possibleTiles))]
				tilePath = append(tilePath, newTile)
				if tilePath[len(tilePath)-1] == endTile {
					if len(tilePath) < 20 {
						pathFailed = true
					}
					break
				}
			}
			if !pathFailed {
				break
			}
			if pathFailed {
				iterations = iterations + 1
			}
			if iterations > 1000 {
				tooDifficult = true
				break
			}
		}

		if tooDifficult {
			i--
			continue
		}

		var mapTiles []Tile
		for j, tile := range tilePath {
			mapTile := Tile{}
			mapTile.ID = tile
			mapTile.BranchIndex = uint8(j + 1)
			switch j {
			case 0:
				mapTile.Type = 1
				mapTile.NextID = tilePath[j+1]
			case len(tilePath) - 1:
				mapTile.Type = 2
			default:
				mapTile.NextID = tilePath[j+1]
			}
			switch i {
			case 0:
				mapTile.PointsReq = int32(2500 + 150*(j-1))
			case 1:
				mapTile.PointsReq = int32(5500 + 600*(j-1))
			case 2:
				mapTile.PointsReq = int32(6500 + 800*(j-1))
			case 3:
				mapTile.PointsReq = int32(7500 + 1000*(j-1))
			case 4:
				mapTile.PointsReq = int32(8500 + 1000*(j-1))
			}
			mapTiles = append(mapTiles, mapTile)
		}

		if i >= 4 {
			guildMaps = append(guildMaps, MapData{uint32(i + 1), 3, mapTiles})
		} else {
			guildMaps = append(guildMaps, MapData{uint32(i + 1), uint32(i + 2), mapTiles})
		}
	}
	return guildMaps
}

func handleMsgMhfGenerateUdGuildMap(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGenerateUdGuildMap)

	// GenerateUdGuildMaps()

	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}
