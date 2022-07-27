package channelserver

import (
	"encoding/hex"
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network/mhfpacket"
	"math/rand"
	"time"
)

func handleMsgMhfSaveMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveMezfesData)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadMezfesData)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0) // Unk

	resp.WriteUint8(2) // Count of the next 2 uint32s
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	resp.WriteUint32(0) // Unk

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEnumerateRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRanking)
	bf := byteframe.NewByteFrame()
	state := s.server.erupeConfig.DevModeOptions.TournamentEvent
	// Unk
	// Unk
	// Start?
	// End?
	midnight := Time_Current_Midnight()
	switch state {
	case 1:
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(3 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(12 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(21 * 24 * time.Hour).Unix()))
	case 2:
		bf.WriteUint32(uint32(midnight.Add(-3 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(9 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(16 * 24 * time.Hour).Unix()))
	case 3:
		bf.WriteUint32(uint32(midnight.Add(-12 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(-9 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(7 * 24 * time.Hour).Unix()))
	default:
		bf.WriteBytes(make([]byte, 16))
		bf.WriteUint32(uint32(Time_Current_Adjusted().Unix())) // TS Current Time
		bf.WriteUint16(1)
		bf.WriteUint32(0)
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}
	bf.WriteUint32(uint32(Time_Current_Adjusted().Unix())) // TS Current Time
	d, _ := hex.DecodeString("031491E631353089F18CF68EAE8EEB97C291E589EF00001200000A54001000000000ED130D949A96B697B393A294B081490000000A55001000010000ED130D949A96B697B393A294B081490000000A56001000020000ED130D949A96B697B393A294B081490000000A57001000030000ED130D949A96B697B393A294B081490000000A58001000040000ED130D949A96B697B393A294B081490000000A59001000050000ED130D949A96B697B393A294B081490000000A5A001000060000ED130D949A96B697B393A294B081490000000A5B001000070000ED130D949A96B697B393A294B081490000000A5C001000080000ED130D949A96B697B393A294B081490000000A5D001000090000ED130D949A96B697B393A294B081490000000A5E0010000A0000ED130D949A96B697B393A294B081490000000A5F0010000B0000ED130D949A96B697B393A294B081490000000A600010000C0000ED130D949A96B697B393A294B081490000000A610010000D0000ED130D949A96B697B393A294B081490000000A620011FFFF0000ED121582DD82F182C882C5949A96B697B393A294B081490000000A63000600EA0000000009834C838C834183570000000A64000600ED000000000B836E838A837D834F838D0000000A65000600EF0000000011834A834E8354839383668381834C83930003000002390006000600000E8CC2906C208B9091E58B9B94740001617E43303581798BA38B5A93E09765817A0A7E433030834E83478358836782C592DE82C182BD8B9B82CC83548343835982F08BA382A40A7E433034817991CE8FDB8B9B817A0A7E433030834C838C8341835781410A836E838A837D834F838D8141834A834E8354839383668381834C83930A7E433037817993FC8FDC8FDC9569817A0A7E4330308B9B947482CC82B582E982B58141835E838B836C835290B68E598C9481410A834F815B834E90B68E598C948141834F815B834E91AB90B68E598C9481410A834F815B834E89F095FA8C94283181603388CA290A2F97C29263837C8343839383672831816031303088CA290A2F8FA08360835083628367817B836E815B8374836083508362836794920A2831816035303088CA290A7E43303381798A4A8DC38AFA8AD4817A0A7E43303032303139944E31318C8E323293FA2031343A303082A982E70A32303139944E31318C8E323593FA2031343A303082DC82C5000000023A0011000700001297C292632082668B89E8E891CA935694740000ED7E43303581798BA38B5A93E09765817A0A7E43303081E182DD82F182C882C5949A96B697B393A294B0814981E282F00A93AF82B697C2926382C98F8A91AE82B782E934906C82DC82C582CC0A97C2926388F582C582A282A982C9918182AD834E838A834182B782E982A90A82F08BA382A40A0A7E433037817993FC8FDC8FDC9569817A0A7E43303091E631343789F18EEB906C8DD582CC8DB02831816032303088CA290A0A7E43303381798A4A8DC38AFA8AD4817A0A7E43303032303139944E31318C8E323293FA2031343A303082A982E70A32303139944E31318C8E323593FA2031343A303082DC82C50A000000023B001000070000128CC2906C2082668B89E8E891CA935694740001497E43303581798BA38B5A93E09765817A0A7E43303081E1949A96B697B393A294B0814981E282F00A82A282A982C9918182AD834E838A834182B782E982A982F08BA382A40A0A7E433037817993FC8FDC8FDC9569817A0A7E43303089A48ED282CC8381835F838B283188CA290A2F8CF68EAE82CC82B582E982B58141835E838B836C835290B68E598C9481410A834F815B834E90B68E598C948141834F815B834E91AB90B68E598C9481410A834F815B834E89F095FA8C94283181603388CA290A2F97C29263837C8343839383672831816031303088CA290A2F8FA08360835083628367817B836E815B8374836083508362836794920A2831816035303088CA290A7E43303381798A4A8DC38AFA8AD4817A0A7E43303032303139944E31318C8E323293FA2031343A303082A982E70A32303139944E31318C8E323593FA2031343A303082DC82C500")
	bf.WriteBytes(d)
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
	midnight := Time_Current_Midnight()
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
	if start == 0 || Time_Current_Adjusted().Unix() > int64(start)+2977200 {
		cleanupFesta(s)
		// Generate a new festa, starting midnight tomorrow
		start = uint32(midnight.Add(24 * time.Hour).Unix())
		s.server.db.Exec("INSERT INTO events (event_type, start_time) VALUES ('festa', to_timestamp($1)::timestamp without time zone)", start)
	}
	timestamps[0] = start
	timestamps[1] = timestamps[0] + 604800
	timestamps[2] = timestamps[1] + 604800
	timestamps[3] = timestamps[2] + 9000
	timestamps[4] = timestamps[3] + 1240200
	return timestamps
}

type Trial struct {
	ID        uint32 `db:"id"`
	Objective uint8  `db:"objective"`
	GoalID    uint32 `db:"goal_id"`
	TimesReq  uint16 `db:"times_req"`
	Locale    uint16 `db:"locale_req"`
	Reward    uint16 `db:"reward"`
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

	if timestamps[0] > uint32(Time_Current_Adjusted().Unix()) {
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
	bf.WriteUint32(uint32(Time_Current_Adjusted().Unix()))
	bf.WriteUint8(4)
	ps.Uint8(bf, "", false)
	bf.WriteUint32(0)
	bf.WriteUint32(blueSouls)
	bf.WriteUint32(redSouls)

	rows, _ = s.server.db.Queryx("SELECT * FROM festa_trials")
	trialData := byteframe.NewByteFrame()
	var count uint16
	for rows.Next() {
		trial := &Trial{}
		err := rows.StructScan(&trial)
		if err != nil {
			continue
		}
		count++
		trialData.WriteUint32(trial.ID)
		trialData.WriteUint8(0) // Unk
		trialData.WriteUint8(trial.Objective)
		trialData.WriteUint32(trial.GoalID)
		trialData.WriteUint16(trial.TimesReq)
		trialData.WriteUint16(trial.Locale)
		trialData.WriteUint16(trial.Reward)
		trialData.WriteUint8(0xFF) // Unk
		trialData.WriteUint8(0xFF) // MonopolyState
		trialData.WriteUint16(0)   // Unk
	}
	bf.WriteUint16(count)
	bf.WriteBytes(trialData.Data())

	// Static bonus rewards
	rewards, _ := hex.DecodeString("001901000007015E05F000000000000100000703E81B6300000000010100000C03E8000000000000000100000D0000000000000000000100000100000000000000000002000007015E05F000000000000200000703E81B6300000000010200000C03E8000000000000000200000D0000000000000000000200000400000000000000000003000007015E05F000000000000300000703E81B6300000000010300000C03E8000000000000000300000D0000000000000000000300000100000000000000000004000007015E05F000000000000400000703E81B6300000000010400000C03E8000000000000000400000D0000000000000000000400000400000000000000000005000007015E05F000000000000500000703E81B6300000000010500000C03E8000000000000000500000D00000000000000000005000001000000000000000000")
	bf.WriteBytes(rewards)

	bf.WriteUint16(0x0001)
	bf.WriteUint32(0xD4C001F4)

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

	d, _ := hex.DecodeString("000000000000000100001388000007D0000003E800000064012C00C8009600640032")
	bf.WriteBytes(d)
	ps.Uint16(bf, "", false)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// state festa (U)ser
func handleMsgMhfStateFestaU(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateFestaU)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	var souls uint32
	s.server.db.QueryRow("SELECT souls FROM guild_characters WHERE character_id=$1", s.charID).Scan(&souls)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(souls)

	// This definitely isn't right, but it does stop you from claiming the festa infinitely.
	var claimed uint32
	s.server.db.QueryRow("SELECT count(*) FROM festa_prizes_accepted fpa WHERE fpa.prize_id=0 AND fpa.character_id=$1", s.charID).Scan(&claimed)
	if claimed > 0 {
		bf.WriteUint32(0) // unk
	} else {
		bf.WriteUint32(0x01000000) // unk
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// state festa (G)uild
func handleMsgMhfStateFestaG(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateFestaG)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	resp := byteframe.NewByteFrame()
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
	for _, member := range members {
		bf.WriteUint16(0)
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
	rand.Seed(time.Now().UnixNano())
	team := uint32(rand.Intn(2))
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
	rows, _ := s.server.db.Queryx("SELECT id, tier, souls_req, item_id, num_item, (SELECT count(*) FROM festa_prizes_accepted fpa WHERE fp.id = fpa.prize_id AND fpa.character_id = 4) AS claimed FROM festa_prizes fp WHERE type='personal'")
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
	rows, _ := s.server.db.Queryx("SELECT id, tier, souls_req, item_id, num_item, (SELECT count(*) FROM festa_prizes_accepted fpa WHERE fp.id = fpa.prize_id AND fpa.character_id = 4) AS claimed FROM festa_prizes fp WHERE type='guild'")
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
