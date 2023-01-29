package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/deltacomp"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func handleMsgMhfLoadPartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPartner)
	var data []byte
	err := s.server.db.QueryRow("SELECT partner FROM characters WHERE id = $1", s.charID).Scan(&data)
	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		s.logger.Warn("Failed to load partner data", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfSavePartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePartner)
	dumpSaveData(s, pkt.RawDataPayload, "partner")
	_, err := s.server.db.Exec("UPDATE characters SET partner=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Warn("Failed to save partner data", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadLegendDispatch(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadLegendDispatch)
	data := []byte{0x03, 0x00, 0x00, 0x00, 0x00, 0x5e, 0x01, 0x8d, 0x40, 0x00, 0x00, 0x00, 0x00, 0x5e, 0x02, 0xde, 0xc0, 0x00, 0x00, 0x00, 0x00, 0x5e, 0x04, 0x30, 0x40}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfLoadHunterNavi(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadHunterNavi)
	var data []byte
	err := s.server.db.QueryRow("SELECT hunternavi FROM characters WHERE id = $1", s.charID).Scan(&data)
	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		s.logger.Warn("Failed to load navi data", zap.Error(err))
		// set first byte to 1 to avoid pop up every time without save
		body := make([]byte, 0x226)
		body[0] = 1
		doAckBufSucceed(s, pkt.AckHandle, body)
	}
}

func handleMsgMhfSaveHunterNavi(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveHunterNavi)
	if pkt.IsDataDiff {
		var data []byte
		// Load existing save
		err := s.server.db.QueryRow("SELECT hunternavi FROM characters WHERE id = $1", s.charID).Scan(&data)
		if err != nil {
			s.logger.Warn("Failed to save navi data", zap.Error(err))
		}

		// Check if we actually had any hunternavi data, using a blank buffer if not.
		// This is requried as the client will try to send a diff after character creation without a prior MsgMhfSaveHunterNavi packet.
		if len(data) == 0 {
			data = make([]byte, 0x226)
			data[0] = 1 // set first byte to 1 to avoid pop up every time without save
		}

		// Perform diff and compress it to write back to db
		s.logger.Info("Diffing...")
		saveOutput := deltacomp.ApplyDataDiff(pkt.RawDataPayload, data)
		_, err = s.server.db.Exec("UPDATE characters SET hunternavi=$1 WHERE id=$2", saveOutput, s.charID)
		if err != nil {
			s.logger.Warn("Failed to save navi data", zap.Error(err))
		}
		s.logger.Info("Wrote recompressed hunternavi back to DB.")
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "hunternavi")
		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET hunternavi=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Warn("Failed to save navi data", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfMercenaryHuntdata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfMercenaryHuntdata)
	if pkt.Unk0 == 1 {
		// Format:
		// uint8 Hunts
		// struct Hunt
		//   uint32 HuntID
		//   uint32 MonID
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
	} else {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 0))
	}
}

func handleMsgMhfEnumerateMercenaryLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateMercenaryLog)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	// Format:
	// struct Log
	//   uint32 Timestamp
	//   []byte Name (len 18)
	//   uint8 Unk
	//   uint8 Unk
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfCreateMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateMercenary)

	bf := byteframe.NewByteFrame()

	var nextID uint32
	s.server.db.QueryRow("SELECT nextval('rasta_id_seq')").Scan(&nextID)

	bf.WriteUint32(nextID)     // New MercID
	bf.WriteUint32(0xDEADBEEF) // Unk

	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfSaveMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveMercenary)
	dumpSaveData(s, pkt.MercData, "mercenary")
	if len(pkt.MercData) > 0 {
		s.server.db.Exec("UPDATE characters SET savemercenary=$1 WHERE id=$2", pkt.MercData, s.charID)
	}
	s.server.db.Exec("UPDATE characters SET gcp=$1 WHERE id=$2", pkt.GCP, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfReadMercenaryW(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMercenaryW)
	if pkt.Unk0 {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
		return
	}
	var data []byte
	var gcp uint32
	s.server.db.QueryRow("SELECT savemercenary FROM characters WHERE id = $1", s.charID).Scan(&data)
	s.server.db.QueryRow("SELECT COALESCE(gcp, 0) FROM characters WHERE id = $1", s.charID).Scan(&gcp)

	resp := byteframe.NewByteFrame()
	if len(data) == 0 {
		resp.WriteBytes(make([]byte, 3))
	} else {
		resp.WriteBytes(data[1:])
		resp.WriteUint32(0) // Unk
	}
	resp.WriteUint32(gcp)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfReadMercenaryM(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMercenaryM)
	var data []byte
	s.server.db.QueryRow("SELECT savemercenary FROM characters WHERE id = $1", pkt.CharID).Scan(&data)
	resp := byteframe.NewByteFrame()
	if len(data) == 0 {
		resp.WriteBool(false)
	} else {
		resp.WriteBytes(data[4:])
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfContractMercenary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfLoadOtomoAirou(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadOtomoAirou)
	var data []byte
	err := s.server.db.QueryRow("SELECT otomoairou FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get partnyaa savedata from db", zap.Error(err))
	}
	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}
}

func handleMsgMhfSaveOtomoAirou(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveOtomoAirou)
	dumpSaveData(s, pkt.RawDataPayload, "otomoairou")
	decomp, err := nullcomp.Decompress(pkt.RawDataPayload[1:])
	if err != nil {
		s.logger.Error("Failed to decompress airou", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf := byteframe.NewByteFrameFromBytes(decomp)
	save := byteframe.NewByteFrame()
	var catsExist uint8
	save.WriteUint8(0)

	cats := bf.ReadUint8()
	for i := 0; i < int(cats); i++ {
		dataLen := bf.ReadUint32()
		catID := bf.ReadUint32()
		if catID == 0 {
			_ = s.server.db.QueryRow("SELECT nextval('airou_id_seq')").Scan(&catID)
		}
		exists := bf.ReadBool()
		data := bf.ReadBytes(uint(dataLen) - 5)
		if exists {
			catsExist++
			save.WriteUint32(dataLen)
			save.WriteUint32(catID)
			save.WriteBool(exists)
			save.WriteBytes(data)
		}
	}
	save.WriteBytes(bf.DataFromCurrent())
	save.Seek(0, 0)
	save.WriteUint8(catsExist)
	comp, err := nullcomp.Compress(save.Data())
	if err != nil {
		s.logger.Error("Failed to compress airou", zap.Error(err))
	} else {
		comp = append([]byte{0x01}, comp...)
		s.server.db.Exec("UPDATE characters SET otomoairou=$1 WHERE id=$2", comp, s.charID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEnumerateAiroulist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateAiroulist)
	resp := byteframe.NewByteFrame()
	if _, err := os.Stat(filepath.Join(s.server.erupeConfig.BinPath, "airoulist.bin")); err == nil {
		data, _ := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, "airoulist.bin"))
		resp.WriteBytes(data)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
		return
	}
	airouList := getGuildAirouList(s)
	resp.WriteUint16(uint16(len(airouList)))
	resp.WriteUint16(uint16(len(airouList)))
	for _, cat := range airouList {
		resp.WriteUint32(cat.CatID)
		resp.WriteBytes(cat.CatName)
		resp.WriteUint32(cat.Experience)
		resp.WriteUint8(cat.Personality)
		resp.WriteUint8(cat.Class)
		resp.WriteUint8(cat.WeaponType)
		resp.WriteUint16(cat.WeaponID)
		resp.WriteUint32(0) // 32 bit unix timestamp, either time at which the cat stops being fatigued or the time at which it started
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

// CatDefinition holds values needed to populate the guild cat list
type CatDefinition struct {
	CatID       uint32
	CatName     []byte
	CurrentTask uint8
	Personality uint8
	Class       uint8
	Experience  uint32
	WeaponType  uint8
	WeaponID    uint16
}

func getGuildAirouList(s *Session) []CatDefinition {
	var guild *Guild
	var err error
	var guildCats []CatDefinition

	// returning 0 cats on any guild issues
	// can probably optimise all of the guild queries pretty heavily
	guild, err = GetGuildInfoByCharacterId(s, s.charID)
	if err != nil {
		return guildCats
	}

	// Get cats used recently
	// Retail reset at midday, 12 hours is a midpoint
	tempBanDuration := 43200 - (1800) // Minus hunt time
	bannedCats := make(map[uint32]int)
	var csvTemp string
	rows, err := s.server.db.Query(`SELECT cats_used
	FROM guild_hunts gh
	INNER JOIN characters c
	ON gh.host_id = c.id
	WHERE c.id=$1 AND gh.return+$2>$3`, s.charID, tempBanDuration, Time_Current_Adjusted().Unix())
	if err != nil {
		s.logger.Warn("Failed to get recently used airous", zap.Error(err))
	}
	for rows.Next() {
		rows.Scan(&csvTemp)
		for i, j := range stringsupport.CSVElems(csvTemp) {
			bannedCats[uint32(j)] = i
		}
	}

	// ellie's GetGuildMembers didn't seem to pull leader?
	rows, err = s.server.db.Query(`SELECT c.otomoairou
	FROM characters c
	INNER JOIN guild_characters gc
	ON gc.character_id = c.id
	WHERE gc.guild_id = $1 AND c.otomoairou IS NOT NULL
	ORDER BY c.id ASC
	LIMIT 60;`, guild.ID)
	if err != nil {
		s.logger.Warn("Selecting otomoairou based on guild failed", zap.Error(err))
		return guildCats
	}

	for rows.Next() {
		var data []byte
		err = rows.Scan(&data)
		if err != nil {
			s.logger.Warn("select failure", zap.Error(err))
			continue
		} else if len(data) == 0 {
			// non extant cats that aren't null in DB
			continue
		}
		// first byte has cat existence in general, can skip if 0
		if data[0] == 1 {
			decomp, err := nullcomp.Decompress(data[1:])
			if err != nil {
				s.logger.Warn("decomp failure", zap.Error(err))
				continue
			}
			bf := byteframe.NewByteFrameFromBytes(decomp)
			cats := GetCatDetails(bf)
			for _, cat := range cats {
				_, exists := bannedCats[cat.CatID]
				if cat.CurrentTask == 4 && !exists {
					guildCats = append(guildCats, cat)
				}
			}
		}
	}
	return guildCats
}

func GetCatDetails(bf *byteframe.ByteFrame) []CatDefinition {
	catCount := bf.ReadUint8()
	cats := make([]CatDefinition, catCount)
	for x := 0; x < int(catCount); x++ {
		var catDef CatDefinition
		// cat sometimes has additional bytes for whatever reason, gift items? timestamp?
		// until actual variance is known we can just seek to end based on start
		catDefLen := bf.ReadUint32()
		catStart, _ := bf.Seek(0, io.SeekCurrent)

		catDef.CatID = bf.ReadUint32()
		bf.Seek(1, io.SeekCurrent)        // unknown value, probably a bool
		catDef.CatName = bf.ReadBytes(18) // always 18 len, reads first null terminated string out of section and discards rest
		catDef.CurrentTask = bf.ReadUint8()
		bf.Seek(16, io.SeekCurrent) // appearance data and what is seemingly null bytes
		catDef.Personality = bf.ReadUint8()
		catDef.Class = bf.ReadUint8()
		bf.Seek(5, io.SeekCurrent)          // affection and colour sliders
		catDef.Experience = bf.ReadUint32() // raw cat rank points, doesn't have a rank
		bf.Seek(1, io.SeekCurrent)          // bool for weapon being equipped
		catDef.WeaponType = bf.ReadUint8()  // weapon type, presumably always 6 for melee?
		catDef.WeaponID = bf.ReadUint16()   // weapon id
		bf.Seek(catStart+int64(catDefLen), io.SeekStart)
		cats[x] = catDef
	}
	return cats
}
