package channelserver

import (
	"erupe-ce/common/token"
	_config "erupe-ce/config"
	"math"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func handleMsgMhfRegisterEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegisterEvent)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(pkt.Unk2)
	bf.WriteUint8(pkt.Unk4)
	bf.WriteUint16(0x1142)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfReleaseEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReleaseEvent)

	// Do this ack manually because it uses a non-(0|1) error code
	/*
		_ACK_SUCCESS = 0
		_ACK_ERROR = 1

		_ACK_EINPROGRESS = 16
		_ACK_ENOENT = 17
		_ACK_ENOSPC = 18
		_ACK_ETIMEOUT = 19

		_ACK_EINVALID = 64
		_ACK_EFAILED = 65
		_ACK_ENOMEM = 66
		_ACK_ENOTEXIT = 67
		_ACK_ENOTREADY = 68
		_ACK_EALREADY = 69
		_ACK_DISABLE_WORK = 71
	*/
	s.QueueSendMHF(&mhfpacket.MsgSysAck{
		AckHandle:        pkt.AckHandle,
		IsBufferResponse: false,
		ErrorCode:        0x41,
		AckData:          []byte{0x00, 0x00, 0x00, 0x00},
	})
}

type Event struct {
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
	Unk4 uint16
	Unk5 uint32
	Unk6 uint32
	Unk7 []uint16
}

func handleMsgMhfEnumerateEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateEvent)
	bf := byteframe.NewByteFrame()

	events := []Event{}

	bf.WriteUint8(uint8(len(events)))
	for _, event := range events {
		bf.WriteUint16(event.Unk0)
		bf.WriteUint16(event.Unk1)
		bf.WriteUint16(event.Unk2)
		bf.WriteUint16(event.Unk3)
		bf.WriteUint16(event.Unk4)
		bf.WriteUint32(event.Unk5)
		bf.WriteUint32(event.Unk6)
		if event.Unk0 == 2 {
			bf.WriteUint8(uint8(len(event.Unk7)))
			for _, u := range event.Unk7 {
				bf.WriteUint16(u)
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
			temp = generateFeatureWeapons(s.server.erupeConfig.GameplayOptions.FeaturedWeapons)
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
	max := 14
	if _config.ErupeConfig.RealClientMode < _config.ZZ {
		max = 13
	}
	if _config.ErupeConfig.RealClientMode < _config.G10 {
		max = 12
	}
	if _config.ErupeConfig.RealClientMode < _config.GG {
		max = 11
	}
	if count > max {
		count = max
	}
	nums := make([]int, 0)
	var result int
	for len(nums) < count {
		rng := token.RNG()
		num := rng.Intn(max)
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
	case 1:
		fallthrough
	case 3:
		expiration = TimeAdjusted().Add(120 * time.Minute)
	case 4:
		expiration = TimeAdjusted().Add(180 * time.Minute)
	case 2:
		fallthrough
	case 5:
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
