package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/token"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"sort"
	"time"
)

func handleMsgMhfSaveMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveMezfesData)
	s.server.db.Exec(`UPDATE characters SET mezfes=$1 WHERE id=$2`, pkt.RawDataPayload, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadMezfesData)
	var data []byte
	s.server.db.QueryRow(`SELECT mezfes FROM characters WHERE id=$1`, s.charID).Scan(&data)
	bf := byteframe.NewByteFrame()
	if len(data) > 0 {
		bf.WriteBytes(data)
	} else {
		bf.WriteUint32(0)
		bf.WriteUint8(2)
		bf.WriteUint32(0)
		bf.WriteUint32(0)
		bf.WriteUint32(0)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func generateTournamentTimestamps(start uint32, debug bool) []uint32 {
	timestamps := make([]uint32, 4)
	midnight := TimeMidnight()
	if debug && start <= 3 {
		midnight := uint32(midnight.Unix())
		switch start {
		case 1:
			timestamps[0] = midnight
			timestamps[1] = timestamps[0] + 259200
			timestamps[2] = timestamps[1] + 766800
			timestamps[3] = timestamps[2] + 604800
		case 2:
			timestamps[0] = midnight - 259200
			timestamps[1] = midnight
			timestamps[2] = timestamps[1] + 766800
			timestamps[3] = timestamps[2] + 604800
		case 3:
			timestamps[0] = midnight - 1026000
			timestamps[1] = midnight - 766800
			timestamps[2] = midnight
			timestamps[3] = timestamps[2] + 604800
		}
		return timestamps
	}
	timestamps[0] = start
	timestamps[1] = timestamps[0] + 259200
	timestamps[2] = timestamps[1] + 766800
	timestamps[3] = timestamps[2] + 604800
	return timestamps
}

type TournamentEvent struct {
	ID        uint32
	CupGroup  uint16
	Limit     int16
	QuestFile uint32
	Name      string
}

type TournamentCup struct {
	ID          uint32
	CupGroup    uint16
	Type        uint16
	Unk2        uint16
	Name        string
	Description string
}

func handleMsgMhfEnumerateRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRanking)
	bf := byteframe.NewByteFrame()

	id, start := uint32(0xBEEFDEAD), uint32(0)
	rows, _ := s.server.db.Queryx("SELECT id, (EXTRACT(epoch FROM start_time)::int) as start_time FROM events WHERE event_type='festa'")
	for rows.Next() {
		rows.Scan(&id, &start)
	}

	var timestamps []uint32
	if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.TournamentEvent >= 0 {
		if s.server.erupeConfig.DevModeOptions.TournamentEvent == 0 {
			bf.WriteBytes(make([]byte, 16))
			bf.WriteUint32(uint32(TimeAdjusted().Unix()))
			bf.WriteUint8(0)
			ps.Uint8(bf, "", true)
			bf.WriteUint16(0)
			bf.WriteUint8(0)
			doAckBufSucceed(s, pkt.AckHandle, bf.Data())
			return
		}
		timestamps = generateTournamentTimestamps(uint32(s.server.erupeConfig.DevModeOptions.TournamentEvent), true)
	} else {
		timestamps = generateTournamentTimestamps(start, false)
	}

	if timestamps[0] > uint32(TimeAdjusted().Unix()) {
		bf.WriteBytes(make([]byte, 16))
		bf.WriteUint32(uint32(TimeAdjusted().Unix()))
		bf.WriteUint8(0)
		ps.Uint8(bf, "", true)
		bf.WriteUint16(0)
		bf.WriteUint8(0)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}

	for _, timestamp := range timestamps {
		bf.WriteUint32(timestamp)
	}
	bf.WriteUint32(uint32(TimeAdjusted().Unix()))
	bf.WriteUint8(1) // TODO: Make this dynamic depending on timestamp
	ps.Uint8(bf, "", true)

	// Temp direct port
	tournamentEvents := []TournamentEvent{
		{2644, 16, 0, 62151, ""},
		{2645, 16, 1, 62151, ""},
		{2646, 16, 2, 62151, ""},
		{2647, 16, 3, 62151, ""},
		{2648, 16, 4, 62151, ""},
		{2649, 16, 5, 62151, ""},
		{2650, 16, 6, 62151, ""},
		{2651, 16, 7, 62151, ""},
		{2652, 16, 8, 62151, ""},
		{2653, 16, 9, 62151, ""},
		{2654, 16, 10, 62151, ""},
		{2655, 16, 11, 62151, ""},
		{2656, 16, 12, 62151, ""},
		{2657, 16, 13, 62151, ""},
		{2658, 17, -1, 62150, ""},
		{2659, 6, 234, 0, ""},
		{2660, 6, 237, 0, ""},
		{2661, 6, 239, 0, ""},
	}
	tournamentCups := []TournamentCup{
		{569, 6, 6, 0, "", ""},
		{570, 17, 7, 0, "", ""},
		{571, 16, 7, 0, "", ""},
	}

	bf.WriteUint16(uint16(len(tournamentEvents)))
	for _, event := range tournamentEvents {
		bf.WriteUint32(event.ID)
		bf.WriteUint16(event.CupGroup)
		bf.WriteInt16(event.Limit)
		bf.WriteUint32(event.QuestFile)
		ps.Uint8(bf, event.Name, true)
	}
	bf.WriteUint8(uint8(len(tournamentCups)))
	for _, cup := range tournamentCups {
		bf.WriteUint32(cup.ID)
		bf.WriteUint16(cup.CupGroup)
		bf.WriteUint16(cup.Type)
		bf.WriteUint16(cup.Unk2)
		ps.Uint8(bf, cup.Name, true)
		ps.Uint16(bf, cup.Description, true)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

type TournamentRanking struct {
	Unk0 uint32
	Unk1 uint32
	Unk2 uint16
	Unk3 uint16 // Unused
	Unk4 uint16
	Unk5 uint16
	Unk6 uint16
	Unk7 uint8
	Unk8 string
	Unk9 string
}

func handleMsgMhfEnumerateOrder(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateOrder)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func cleanupFesta(s *Session) {
	s.server.db.Exec("DELETE FROM events WHERE event_type='festa'")
	s.server.db.Exec("DELETE FROM festa_registrations")
	s.server.db.Exec("DELETE FROM festa_prizes_accepted")
	s.server.db.Exec("UPDATE guild_characters SET souls=0")
}

func generateFestaTimestamps(s *Session, start uint32, debug bool) []uint32 {
	timestamps := make([]uint32, 5)
	midnight := TimeMidnight()
	if debug && start <= 3 {
		midnight := uint32(midnight.Unix())
		switch start {
		case 1:
			timestamps[0] = midnight
			timestamps[1] = timestamps[0] + 604800
			timestamps[2] = timestamps[1] + 604800
			timestamps[3] = timestamps[2] + 9000
			timestamps[4] = timestamps[3] + 1240200
		case 2:
			timestamps[0] = midnight - 604800
			timestamps[1] = midnight
			timestamps[2] = timestamps[1] + 604800
			timestamps[3] = timestamps[2] + 9000
			timestamps[4] = timestamps[3] + 1240200
		case 3:
			timestamps[0] = midnight - 1209600
			timestamps[1] = midnight - 604800
			timestamps[2] = midnight
			timestamps[3] = timestamps[2] + 9000
			timestamps[4] = timestamps[3] + 1240200
		}
		return timestamps
	}
	if start == 0 || TimeAdjusted().Unix() > int64(start)+3024000 {
		cleanupFesta(s)
		// Generate a new festa, starting 11am tomorrow
		start = uint32(midnight.Add(35 * time.Hour).Unix())
		s.server.db.Exec("INSERT INTO events (event_type, start_time) VALUES ('festa', to_timestamp($1)::timestamp without time zone)", start)
	}
	timestamps[0] = start
	timestamps[1] = timestamps[0] + 604800
	timestamps[2] = timestamps[1] + 604800
	timestamps[3] = timestamps[2] + 9000
	timestamps[4] = timestamps[3] + 1240200
	return timestamps
}

type FestaTrial struct {
	ID        uint32 `db:"id"`
	Objective uint16 `db:"objective"`
	GoalID    uint32 `db:"goal_id"`
	TimesReq  uint16 `db:"times_req"`
	Locale    uint16 `db:"locale_req"`
	Reward    uint16 `db:"reward"`
	Monopoly  uint16
	Unk       uint16
}

type FestaReward struct {
	Unk0     uint8
	Unk1     uint8
	ItemType uint16
	Quantity uint16
	ItemID   uint16
	Unk5     uint16
	Unk6     uint16
	Unk7     uint8
}

func handleMsgMhfInfoFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoFesta)
	bf := byteframe.NewByteFrame()

	id, start := uint32(0xDEADBEEF), uint32(0)
	rows, _ := s.server.db.Queryx("SELECT id, (EXTRACT(epoch FROM start_time)::int) as start_time FROM events WHERE event_type='festa'")
	for rows.Next() {
		rows.Scan(&id, &start)
	}

	var timestamps []uint32
	if s.server.erupeConfig.DevMode && s.server.erupeConfig.DevModeOptions.FestaEvent >= 0 {
		if s.server.erupeConfig.DevModeOptions.FestaEvent == 0 {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		timestamps = generateFestaTimestamps(s, uint32(s.server.erupeConfig.DevModeOptions.FestaEvent), true)
	} else {
		timestamps = generateFestaTimestamps(s, start, false)
	}

	if timestamps[0] > uint32(TimeAdjusted().Unix()) {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	var blueSouls, redSouls uint32
	s.server.db.QueryRow("SELECT SUM(gc.souls) FROM guild_characters gc INNER JOIN festa_registrations fr ON fr.guild_id = gc.guild_id WHERE fr.team = 'blue'").Scan(&blueSouls)
	s.server.db.QueryRow("SELECT SUM(gc.souls) FROM guild_characters gc INNER JOIN festa_registrations fr ON fr.guild_id = gc.guild_id WHERE fr.team = 'red'").Scan(&redSouls)

	bf.WriteUint32(id)
	for _, timestamp := range timestamps {
		bf.WriteUint32(timestamp)
	}
	bf.WriteUint32(uint32(TimeAdjusted().Unix()))
	bf.WriteUint8(4)
	ps.Uint8(bf, "", false)
	bf.WriteUint32(0)
	bf.WriteUint32(blueSouls)
	bf.WriteUint32(redSouls)

	var trials []FestaTrial
	var trial FestaTrial
	rows, _ = s.server.db.Queryx("SELECT * FROM festa_trials")
	for rows.Next() {
		err := rows.StructScan(&trial)
		if err != nil {
			continue
		}
		trials = append(trials, trial)
	}
	bf.WriteUint16(uint16(len(trials)))
	for _, trial := range trials {
		bf.WriteUint32(trial.ID)
		bf.WriteUint16(trial.Objective)
		bf.WriteUint32(trial.GoalID)
		bf.WriteUint16(trial.TimesReq)
		bf.WriteUint16(trial.Locale)
		bf.WriteUint16(trial.Reward)
		trial.Monopoly = 0xFFFF // NYI
		bf.WriteUint16(trial.Monopoly)
		bf.WriteUint16(trial.Unk)
	}

	// The Winner and Loser Armor IDs are missing
	rewards := []FestaReward{
		{1, 0, 7, 350, 1520, 0, 0, 0},
		{1, 0, 7, 1000, 7011, 0, 0, 1},
		{1, 0, 12, 1000, 0, 0, 0, 0},
		{1, 0, 13, 0, 0, 0, 0, 0},
		//{1, 0, 1, 0, 0, 0, 0, 0},
		{2, 0, 7, 350, 1520, 0, 0, 0},
		{2, 0, 7, 1000, 7011, 0, 0, 1},
		{2, 0, 12, 1000, 0, 0, 0, 0},
		{2, 0, 13, 0, 0, 0, 0, 0},
		//{2, 0, 4, 0, 0, 0, 0, 0},
		{3, 0, 7, 350, 1520, 0, 0, 0},
		{3, 0, 7, 1000, 7011, 0, 0, 1},
		{3, 0, 12, 1000, 0, 0, 0, 0},
		{3, 0, 13, 0, 0, 0, 0, 0},
		//{3, 0, 1, 0, 0, 0, 0, 0},
		{4, 0, 7, 350, 1520, 0, 0, 0},
		{4, 0, 7, 1000, 7011, 0, 0, 1},
		{4, 0, 12, 1000, 0, 0, 0, 0},
		{4, 0, 13, 0, 0, 0, 0, 0},
		//{4, 0, 4, 0, 0, 0, 0, 0},
		{5, 0, 7, 350, 1520, 0, 0, 0},
		{5, 0, 7, 1000, 7011, 0, 0, 1},
		{5, 0, 12, 1000, 0, 0, 0, 0},
		{5, 0, 13, 0, 0, 0, 0, 0},
		//{5, 0, 1, 0, 0, 0, 0, 0},
	}
	bf.WriteUint16(uint16(len(rewards)))
	for _, reward := range rewards {
		bf.WriteUint8(reward.Unk0)
		bf.WriteUint8(reward.Unk1)
		bf.WriteUint16(reward.ItemType)
		bf.WriteUint16(reward.Quantity)
		bf.WriteUint16(reward.ItemID)
		bf.WriteUint16(reward.Unk5)
		bf.WriteUint16(reward.Unk6)
		bf.WriteUint8(reward.Unk7)
	}

	if _config.ErupeConfig.RealClientMode <= _config.G61 {
		if s.server.erupeConfig.GameplayOptions.MaximumFP > 0xFFFF {
			s.server.erupeConfig.GameplayOptions.MaximumFP = 0xFFFF
		}
		bf.WriteUint16(uint16(s.server.erupeConfig.GameplayOptions.MaximumFP))
	} else {
		bf.WriteUint32(s.server.erupeConfig.GameplayOptions.MaximumFP)
	}
	bf.WriteUint16(500)

	categoryWinners := uint16(0) // NYI
	bf.WriteUint16(categoryWinners)
	for i := uint16(0); i < categoryWinners; i++ {
		bf.WriteUint32(0)      // Guild ID
		bf.WriteUint16(i + 1)  // Category ID
		bf.WriteUint16(0)      // Festa Team
		ps.Uint8(bf, "", true) // Guild Name
	}

	dailyWinners := uint16(0) // NYI
	bf.WriteUint16(dailyWinners)
	for i := uint16(0); i < dailyWinners; i++ {
		bf.WriteUint32(0)      // Guild ID
		bf.WriteUint16(i + 1)  // Category ID
		bf.WriteUint16(0)      // Festa Team
		ps.Uint8(bf, "", true) // Guild Name
	}

	// Unknown values
	bf.WriteUint32(1)
	bf.WriteUint32(5000)
	bf.WriteUint32(2000)
	bf.WriteUint32(1000)
	bf.WriteUint32(100)
	bf.WriteUint16(300)
	bf.WriteUint16(200)
	bf.WriteUint16(150)
	bf.WriteUint16(100)
	bf.WriteUint16(50)

	ps.Uint16(bf, "", false)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// state festa (U)ser
func handleMsgMhfStateFestaU(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateFestaU)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	applicant := false
	if guild != nil {
		applicant, _ = guild.HasApplicationForCharID(s, s.charID)
	}
	if err != nil || guild == nil || applicant {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	var souls, exists uint32
	s.server.db.QueryRow("SELECT souls FROM guild_characters WHERE character_id=$1", s.charID).Scan(&souls)
	err = s.server.db.QueryRow("SELECT prize_id FROM festa_prizes_accepted WHERE prize_id=0 AND character_id=$1", s.charID).Scan(&exists)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(souls)
	if err != nil {
		bf.WriteBool(true)
		bf.WriteBool(false)
	} else {
		bf.WriteBool(false)
		bf.WriteBool(true)
	}
	bf.WriteUint16(0) // Unk
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// state festa (G)uild
func handleMsgMhfStateFestaG(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateFestaG)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	applicant := false
	if guild != nil {
		applicant, _ = guild.HasApplicationForCharID(s, s.charID)
	}
	resp := byteframe.NewByteFrame()
	if err != nil || guild == nil || applicant {
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0xFFFFFFFF)
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
		return
	}
	resp.WriteUint32(guild.Souls)
	resp.WriteUint32(1) // unk
	resp.WriteUint32(1) // unk
	resp.WriteUint32(1) // unk, rank?
	resp.WriteUint32(1) // unk
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEnumerateFestaMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateFestaMember)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	members, err := GetGuildMembers(s, guild.ID, false)
	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(members)))
	bf.WriteUint16(0) // Unk
	sort.Slice(members, func(i, j int) bool {
		return members[i].Souls > members[j].Souls
	})
	for _, member := range members {
		bf.WriteUint32(member.CharID)
		bf.WriteUint32(member.Souls)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfVoteFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfVoteFesta)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEntryFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryFesta)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	team := uint32(token.RNG().Intn(2))
	switch team {
	case 0:
		s.server.db.Exec("INSERT INTO festa_registrations VALUES ($1, 'blue')", guild.ID)
	case 1:
		s.server.db.Exec("INSERT INTO festa_registrations VALUES ($1, 'red')", guild.ID)
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(team)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfChargeFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfChargeFesta)
	s.server.db.Exec("UPDATE guild_characters SET souls=souls+$1 WHERE character_id=$2", pkt.Souls, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireFesta)
	s.server.db.Exec("INSERT INTO public.festa_prizes_accepted VALUES (0, $1)", s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireFestaPersonalPrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireFestaPersonalPrize)
	s.server.db.Exec("INSERT INTO public.festa_prizes_accepted VALUES ($1, $2)", pkt.PrizeID, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireFestaIntermediatePrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireFestaIntermediatePrize)
	s.server.db.Exec("INSERT INTO public.festa_prizes_accepted VALUES ($1, $2)", pkt.PrizeID, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

type Prize struct {
	ID       uint32 `db:"id"`
	Tier     uint32 `db:"tier"`
	SoulsReq uint32 `db:"souls_req"`
	ItemID   uint32 `db:"item_id"`
	NumItem  uint32 `db:"num_item"`
	Claimed  int    `db:"claimed"`
}

func handleMsgMhfEnumerateFestaPersonalPrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateFestaPersonalPrize)
	rows, _ := s.server.db.Queryx(`SELECT id, tier, souls_req, item_id, num_item, (SELECT count(*) FROM festa_prizes_accepted fpa WHERE fp.id = fpa.prize_id AND fpa.character_id = $1) AS claimed FROM festa_prizes fp WHERE type='personal'`, s.charID)
	var count uint32
	prizeData := byteframe.NewByteFrame()
	for rows.Next() {
		prize := &Prize{}
		err := rows.StructScan(&prize)
		if err != nil {
			continue
		}
		count++
		prizeData.WriteUint32(prize.ID)
		prizeData.WriteUint32(prize.Tier)
		prizeData.WriteUint32(prize.SoulsReq)
		prizeData.WriteUint32(7) // Unk
		prizeData.WriteUint32(prize.ItemID)
		prizeData.WriteUint32(prize.NumItem)
		prizeData.WriteBool(prize.Claimed > 0)
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(count)
	bf.WriteBytes(prizeData.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnumerateFestaIntermediatePrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateFestaIntermediatePrize)
	rows, _ := s.server.db.Queryx(`SELECT id, tier, souls_req, item_id, num_item, (SELECT count(*) FROM festa_prizes_accepted fpa WHERE fp.id = fpa.prize_id AND fpa.character_id = $1) AS claimed FROM festa_prizes fp WHERE type='guild'`, s.charID)
	var count uint32
	prizeData := byteframe.NewByteFrame()
	for rows.Next() {
		prize := &Prize{}
		err := rows.StructScan(&prize)
		if err != nil {
			continue
		}
		count++
		prizeData.WriteUint32(prize.ID)
		prizeData.WriteUint32(prize.Tier)
		prizeData.WriteUint32(prize.SoulsReq)
		prizeData.WriteUint32(7) // Unk
		prizeData.WriteUint32(prize.ItemID)
		prizeData.WriteUint32(prize.NumItem)
		prizeData.WriteBool(prize.Claimed > 0)
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(count)
	bf.WriteBytes(prizeData.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
