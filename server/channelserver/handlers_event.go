package channelserver

import (
	"erupe-ce/common/token"
	_config "erupe-ce/config"
	"math"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

type Event struct {
	EventType    uint16 //0 Nothing //1 or 2 "An Acient Dragon has attacked the fort"
	Unk1         uint16
	Unk2         uint16
	Unk3         uint16
	Unk4         uint16
	StartTime    uint32
	EndTime      uint32
	QuestFileIDs []uint16
}

func cleanupEnumEvent(s *Session) {
	s.server.db.Exec("DELETE FROM events WHERE event_type='ancientdragon'")
}
func generateEnumEventTimestamps(s *Session, start uint32) []uint32 {
	timestamps := make([]uint32, 2)
	midnight := TimeMidnight()
	//if start is 0 or start is after a duration
	if start == 0 || TimeAdjusted().Unix() > int64(start)+int64(s.server.erupeConfig.GameplayOptions.EnumerateEvent.RestartAfter) {
		cleanupEnumEvent(s)
		// Generate a new diva defense, starting midnight tomorrow
		start = uint32(midnight.Add(24 * time.Hour).Unix())
		s.server.db.Exec("INSERT INTO events (event_type, start_time) VALUES ('ancientdragon', to_timestamp($1)::timestamp without time zone)", start)
	}
	timestamps[0] = start
	timestamps[1] = timestamps[0] + uint32(s.server.erupeConfig.GameplayOptions.EnumerateEvent.Duration)
	return timestamps
}

func handleMsgMhfEnumerateEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateEvent)
	bf := byteframe.NewByteFrame()
	id, start := uint32(0xCAFEBEEF), uint32(0)

	rows, _ := s.server.db.Queryx("SELECT id, (EXTRACT(epoch FROM start_time)::int) as start_time FROM events WHERE event_type='ancientdragon'")
	if !rows.Next() {
		for rows.Next() {
			rows.Scan(&id, &start)
		}
	}

	var timestamps []uint32
	events := []Event{}

	if s.server.erupeConfig.GameplayOptions.EnumerateEvent.Enabled {
		timestamps = generateEnumEventTimestamps(s, start)
		events = append(events, Event{s.server.erupeConfig.GameplayOptions.EnumerateEvent.EventID, 0, 0, 0, 0, timestamps[0], timestamps[1], s.server.erupeConfig.GameplayOptions.EnumerateEvent.QuestIDs})
	}

	bf.WriteUint8(uint8(len(events)))
	for _, event := range events {
		bf.WriteUint16(event.EventType)
		bf.WriteUint16(event.Unk1)
		bf.WriteUint16(event.Unk2)
		bf.WriteUint16(event.Unk3)
		bf.WriteUint16(event.Unk4)
		bf.WriteUint32(event.StartTime)
		bf.WriteUint32(event.EndTime)
		if event.EventType == 2 {
			bf.WriteUint8(uint8(len(event.QuestFileIDs)))
			for _, qf := range event.QuestFileIDs {
				bf.WriteUint16(qf)
			}
		}
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

type activeFeature struct {
	StartTime      time.Time `db:"start_time"`
	ActiveFeatures uint32    `db:"featured"`
}

func handleMsgMhfGetWeeklySchedule(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySchedule)

	var features []activeFeature
	times := []time.Time{
		TimeMidnight().Add(-24 * time.Hour),
		TimeMidnight(),
		TimeMidnight().Add(24 * time.Hour),
	}

	for _, t := range times {
		var temp activeFeature
		err := s.server.db.QueryRowx(`SELECT start_time, featured FROM feature_weapon WHERE start_time=$1`, t).StructScan(&temp)
		if err != nil || temp.StartTime.IsZero() {
			weapons := token.RNG.Intn(s.server.erupeConfig.GameplayOptions.MaxFeatureWeapons-s.server.erupeConfig.GameplayOptions.MinFeatureWeapons+1) + s.server.erupeConfig.GameplayOptions.MinFeatureWeapons
			temp = generateFeatureWeapons(weapons)
			temp.StartTime = t
			s.server.db.Exec(`INSERT INTO feature_weapon VALUES ($1, $2)`, temp.StartTime, temp.ActiveFeatures)
		}
		features = append(features, temp)
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(features)))
	bf.WriteUint32(uint32(TimeAdjusted().Add(-5 * time.Minute).Unix()))
	for _, feature := range features {
		bf.WriteUint32(uint32(feature.StartTime.Unix()))
		bf.WriteUint32(feature.ActiveFeatures)
		bf.WriteUint16(0)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func generateFeatureWeapons(count int) activeFeature {
	_max := 14
	if _config.ErupeConfig.RealClientMode < _config.ZZ {
		_max = 13
	}
	if _config.ErupeConfig.RealClientMode < _config.G10 {
		_max = 12
	}
	if _config.ErupeConfig.RealClientMode < _config.GG {
		_max = 11
	}
	if count > _max {
		count = _max
	}
	nums := make([]int, 0)
	var result int
	for len(nums) < count {
		num := token.RNG.Intn(_max)
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	for _, num := range nums {
		result += int(math.Pow(2, float64(num)))
	}
	return activeFeature{ActiveFeatures: uint32(result)}
}

type loginBoost struct {
	WeekReq    uint8 `db:"week_req"`
	WeekCount  uint8
	Active     bool
	Expiration time.Time `db:"expiration"`
	Reset      time.Time `db:"reset"`
}

func handleMsgMhfGetKeepLoginBoostStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKeepLoginBoostStatus)

	bf := byteframe.NewByteFrame()

	var loginBoosts []loginBoost
	rows, err := s.server.db.Queryx("SELECT week_req, expiration, reset FROM login_boost WHERE char_id=$1 ORDER BY week_req", s.charID)
	if err != nil || s.server.erupeConfig.GameplayOptions.DisableLoginBoost {
		rows.Close()
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 35))
		return
	}
	for rows.Next() {
		var temp loginBoost
		rows.StructScan(&temp)
		loginBoosts = append(loginBoosts, temp)
	}
	if len(loginBoosts) == 0 {
		temp := TimeWeekStart()
		loginBoosts = []loginBoost{
			{WeekReq: 1, Expiration: temp},
			{WeekReq: 2, Expiration: temp},
			{WeekReq: 3, Expiration: temp},
			{WeekReq: 4, Expiration: temp},
			{WeekReq: 5, Expiration: temp},
		}
		for _, boost := range loginBoosts {
			s.server.db.Exec(`INSERT INTO login_boost VALUES ($1, $2, $3, $4)`, s.charID, boost.WeekReq, boost.Expiration, time.Time{})
		}
	}

	for _, boost := range loginBoosts {
		// Reset if next week
		if !boost.Reset.IsZero() && boost.Reset.Before(TimeAdjusted()) {
			boost.Expiration = TimeWeekStart()
			boost.Reset = time.Time{}
			s.server.db.Exec(`UPDATE login_boost SET expiration=$1, reset=$2 WHERE char_id=$3 AND week_req=$4`, boost.Expiration, boost.Reset, s.charID, boost.WeekReq)
		}

		boost.WeekCount = uint8((TimeAdjusted().Unix()-boost.Expiration.Unix())/604800 + 1)

		if boost.WeekCount >= boost.WeekReq {
			boost.Active = true
			boost.WeekCount = boost.WeekReq
		}

		// Show reset timer on expired boosts
		if boost.Reset.After(TimeAdjusted()) {
			boost.Active = true
			boost.WeekCount = 0
		}

		bf.WriteUint8(boost.WeekReq)
		bf.WriteBool(boost.Active)
		bf.WriteUint8(boost.WeekCount)
		if !boost.Reset.IsZero() {
			bf.WriteUint32(uint32(boost.Expiration.Unix()))
		} else {
			bf.WriteUint32(0)
		}
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUseKeepLoginBoost(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUseKeepLoginBoost)
	var expiration time.Time
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	switch pkt.BoostWeekUsed {
	case 1, 3:
		expiration = TimeAdjusted().Add(120 * time.Minute)
	case 4:
		expiration = TimeAdjusted().Add(180 * time.Minute)
	case 2, 5:
		expiration = TimeAdjusted().Add(240 * time.Minute)
	}
	bf.WriteUint32(uint32(expiration.Unix()))
	s.server.db.Exec(`UPDATE login_boost SET expiration=$1, reset=$2 WHERE char_id=$3 AND week_req=$4`, expiration, TimeWeekNext(), s.charID, pkt.BoostWeekUsed)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetRestrictionEvent)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
