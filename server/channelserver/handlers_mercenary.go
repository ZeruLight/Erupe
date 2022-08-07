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
	"math/rand"
	"os"
	"path/filepath"
)

// THERE ARE [PARTENER] [MERCENARY] [OTOMO AIRU]

///////////////////////////////////////////
///				 PARTENER				 //
///////////////////////////////////////////

func handleMsgMhfLoadPartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPartner)
	// load partner from database
	var data []byte
	err := s.server.db.QueryRow("SELECT partner FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get partner savedata from db", zap.Error(err))
	}
	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}
	// TODO(Andoryuuta): Figure out unusual double ack. One sized, one not.
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfSavePartner(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePartner)

	dumpSaveData(s, pkt.RawDataPayload, "_partner")

	_, err := s.server.db.Exec("UPDATE characters SET partner=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update partner savedata in db", zap.Error(err))
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
	if err != nil {
		s.logger.Fatal("Failed to get hunter navigation savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		// set first byte to 1 to avoid pop up every time without save
		body := make([]byte, 0x226)
		body[0] = 1
		doAckBufSucceed(s, pkt.AckHandle, body)
	}
}

func handleMsgMhfSaveHunterNavi(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveHunterNavi)

	dumpSaveData(s, pkt.RawDataPayload, "_hunternavi")

	if pkt.IsDataDiff {
		var data []byte

		// Load existing save
		err := s.server.db.QueryRow("SELECT hunternavi FROM characters WHERE id = $1", s.charID).Scan(&data)
		if err != nil {
			s.logger.Fatal("Failed to get hunternavi savedata from db", zap.Error(err))
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
			s.logger.Fatal("Failed to update hunternavi savedata in db", zap.Error(err))
		}

		s.logger.Info("Wrote recompressed hunternavi back to DB.")
	} else {
		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET hunternavi=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update hunternavi savedata in db", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

///////////////////////////////////////////

///////////////////////////////////////////
///				 MERCENARY				 //
///////////////////////////////////////////

func handleMsgMhfMercenaryHuntdata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfMercenaryHuntdata)
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 0x0A))
}

func handleMsgMhfEnumerateMercenaryLog(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCreateMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateMercenary)

	bf := byteframe.NewByteFrame()

	bf.WriteUint32(0x00)          // Unk
	bf.WriteUint32(rand.Uint32()) // Partner ID?

	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfSaveMercenary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveMercenary)
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
	GCPValue := bf.ReadUint32()
	_ = bf.ReadUint32() // unk
	MercDataSize := bf.ReadUint32()
	MercData := bf.ReadBytes(uint(MercDataSize))
	_ = bf.ReadUint32() // unk

	if MercDataSize > 0 {
		// the save packet has an extra null byte after its size
		_, err := s.server.db.Exec("UPDATE characters SET savemercenary=$1 WHERE id=$2", MercData[:MercDataSize], s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update savemercenary and gcp in db", zap.Error(err))
		}
	}
	// gcp value is always present regardless
	_, err := s.server.db.Exec("UPDATE characters SET gcp=$1 WHERE id=$2", GCPValue, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update savemercenary and gcp in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfReadMercenaryW(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMercenaryW)
	var data []byte
	var gcp uint32
	// still has issues
	err := s.server.db.QueryRow("SELECT savemercenary FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get savemercenary data from db", zap.Error(err))
	}

	err = s.server.db.QueryRow("SELECT COALESCE(gcp, 0) FROM characters WHERE id = $1", s.charID).Scan(&gcp)
	if err != nil {
		panic(err)
	}
	if len(data) == 0 {
		data = []byte{0x00}
	}

	resp := byteframe.NewByteFrame()
	resp.WriteBytes(data)
	resp.WriteUint16(0)
	resp.WriteUint32(gcp)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfReadMercenaryM(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadMercenaryM)
	// accessing actual rasta data of someone else still unsure of the formatting of this
	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfContractMercenary(s *Session, p mhfpacket.MHFPacket) {}

///////////////////////////////////////////

///////////////////////////////////////////
///				OTOMO AIRU				 //
///////////////////////////////////////////

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
	decomp, err := nullcomp.Decompress(pkt.RawDataPayload[1:])
	if err != nil {
		s.logger.Error("Failed to decompress airou", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf := byteframe.NewByteFrameFromBytes(decomp)
	cats := bf.ReadUint8()
	for i := 0; i < int(cats); i++ {
		dataLen := bf.ReadUint32()
		catID := bf.ReadUint32()
		if catID == 0 {
			var nextID uint32
			_ = s.server.db.QueryRow("SELECT nextval('airou_id_seq')").Scan(&nextID)
			bf.Seek(-4, io.SeekCurrent)
			bf.WriteUint32(nextID)
		}
		_ = bf.ReadBytes(uint(dataLen) - 4)
	}
	comp, err := nullcomp.Compress(bf.Data())
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

///////////////////////////////////////////
