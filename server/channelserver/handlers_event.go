package channelserver

import (
	"erupe-ce/config"
	"erupe-ce/internal/model"
	"erupe-ce/utils/gametime"
	"erupe-ce/utils/token"
	"math"
	"time"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"

	"github.com/jmoiron/sqlx"
)

func handleMsgMhfEnumerateEvent(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateEvent)
	bf := byteframe.NewByteFrame()

	events := []model.Event{}

	bf.WriteUint8(uint8(len(events)))
	for _, event := range events {
		bf.WriteUint16(event.EventType)
		bf.WriteUint16(event.Unk1)
		bf.WriteUint16(event.Unk2)
		bf.WriteUint16(event.Unk3)
		bf.WriteUint16(event.Unk4)
		bf.WriteUint32(event.Unk5)
		bf.WriteUint32(event.Unk6)
		if event.EventType == 2 {
			bf.WriteUint8(uint8(len(event.QuestFileIDs)))
			for _, qf := range event.QuestFileIDs {
				bf.WriteUint16(qf)
			}
		}
	}

	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetWeeklySchedule(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySchedule)

	var features []model.ActiveFeature
	times := []time.Time{
		gametime.TimeMidnight().Add(-24 * time.Hour),
		gametime.TimeMidnight(),
		gametime.TimeMidnight().Add(24 * time.Hour),
	}

	for _, t := range times {
		var temp model.ActiveFeature
		err := db.QueryRowx(`SELECT start_time, featured FROM feature_weapon WHERE start_time=$1`, t).StructScan(&temp)
		if err != nil || temp.StartTime.IsZero() {
			weapons := token.RNG.Intn(config.GetConfig().GameplayOptions.MaxFeatureWeapons-config.GetConfig().GameplayOptions.MinFeatureWeapons+1) + config.GetConfig().GameplayOptions.MinFeatureWeapons
			temp = generateFeatureWeapons(weapons)
			temp.StartTime = t
			db.Exec(`INSERT INTO feature_weapon VALUES ($1, $2)`, temp.StartTime, temp.ActiveFeatures)
		}
		features = append(features, temp)
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(features)))
	bf.WriteUint32(uint32(gametime.TimeAdjusted().Add(-5 * time.Minute).Unix()))
	for _, feature := range features {
		bf.WriteUint32(uint32(feature.StartTime.Unix()))
		bf.WriteUint32(feature.ActiveFeatures)
		bf.WriteUint16(0)
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func generateFeatureWeapons(count int) model.ActiveFeature {
	_max := 14
	if config.GetConfig().ClientID < config.ZZ {
		_max = 13
	}
	if config.GetConfig().ClientID < config.G10 {
		_max = 12
	}
	if config.GetConfig().ClientID < config.GG {
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
	return model.ActiveFeature{ActiveFeatures: uint32(result)}
}

func handleMsgMhfGetKeepLoginBoostStatus(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKeepLoginBoostStatus)

	bf := byteframe.NewByteFrame()

	var loginBoosts []model.LoginBoost
	rows, err := db.Queryx("SELECT week_req, expiration, reset FROM login_boost WHERE char_id=$1 ORDER BY week_req", s.CharID)
	if err != nil || config.GetConfig().GameplayOptions.DisableLoginBoost {
		rows.Close()
		s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 35))
		return
	}
	for rows.Next() {
		var temp model.LoginBoost
		rows.StructScan(&temp)
		loginBoosts = append(loginBoosts, temp)
	}
	if len(loginBoosts) == 0 {
		temp := gametime.TimeWeekStart()
		loginBoosts = []model.LoginBoost{
			{WeekReq: 1, Expiration: temp},
			{WeekReq: 2, Expiration: temp},
			{WeekReq: 3, Expiration: temp},
			{WeekReq: 4, Expiration: temp},
			{WeekReq: 5, Expiration: temp},
		}
		for _, boost := range loginBoosts {
			db.Exec(`INSERT INTO login_boost VALUES ($1, $2, $3, $4)`, s.CharID, boost.WeekReq, boost.Expiration, time.Time{})
		}
	}

	for _, boost := range loginBoosts {
		// Reset if next week
		if !boost.Reset.IsZero() && boost.Reset.Before(gametime.TimeAdjusted()) {
			boost.Expiration = gametime.TimeWeekStart()
			boost.Reset = time.Time{}
			db.Exec(`UPDATE login_boost SET expiration=$1, reset=$2 WHERE char_id=$3 AND week_req=$4`, boost.Expiration, boost.Reset, s.CharID, boost.WeekReq)
		}

		boost.WeekCount = uint8((gametime.TimeAdjusted().Unix()-boost.Expiration.Unix())/604800 + 1)

		if boost.WeekCount >= boost.WeekReq {
			boost.Active = true
			boost.WeekCount = boost.WeekReq
		}

		// Show reset timer on expired boosts
		if boost.Reset.After(gametime.TimeAdjusted()) {
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

	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfUseKeepLoginBoost(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUseKeepLoginBoost)
	var expiration time.Time
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)
	switch pkt.BoostWeekUsed {
	case 1, 3:
		expiration = gametime.TimeAdjusted().Add(120 * time.Minute)
	case 4:
		expiration = gametime.TimeAdjusted().Add(180 * time.Minute)
	case 2, 5:
		expiration = gametime.TimeAdjusted().Add(240 * time.Minute)
	}
	bf.WriteUint32(uint32(expiration.Unix()))

	db.Exec(`UPDATE login_boost SET expiration=$1, reset=$2 WHERE char_id=$3 AND week_req=$4`, expiration, gametime.TimeWeekNext(), s.CharID, pkt.BoostWeekUsed)
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetRestrictionEvent(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetRestrictionEvent(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetRestrictionEvent)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}
