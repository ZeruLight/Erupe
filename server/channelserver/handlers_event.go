package channelserver

import (
	"math"
	"math/rand"
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

func handleMsgMhfEnumerateEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateEvent)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

type activeFeature struct {
	StartTime      time.Time `db:"start_time"`
	ActiveFeatures uint32    `db:"featured"`
}

func handleMsgMhfGetWeeklySchedule(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySchedule)

	var features []activeFeature
	rows, _ := s.server.db.Queryx(`SELECT start_time, featured FROM feature_weapon WHERE start_time=$1 OR start_time=$2`, Time_Current_Midnight().Add(-24*time.Hour), Time_Current_Midnight())
	for rows.Next() {
		var feature activeFeature
		rows.StructScan(&feature)
		features = append(features, feature)
	}

	if len(features) < 2 {
		if len(features) == 0 {
			feature := generateFeatureWeapons(s.server.erupeConfig.FeaturedWeapons)
			feature.StartTime = Time_Current_Midnight().Add(-24 * time.Hour)
			features = append(features, feature)
			s.server.db.Exec(`INSERT INTO feature_weapon VALUES ($1, $2)`, feature.StartTime, feature.ActiveFeatures)
		}
		feature := generateFeatureWeapons(s.server.erupeConfig.FeaturedWeapons)
		feature.StartTime = Time_Current_Midnight()
		features = append(features, feature)
		s.server.db.Exec(`INSERT INTO feature_weapon VALUES ($1, $2)`, feature.StartTime, feature.ActiveFeatures)
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(2)
	bf.WriteUint32(uint32(Time_Current_Adjusted().Add(-5 * time.Minute).Unix()))
	for _, feature := range features {
		bf.WriteUint32(uint32(feature.StartTime.Unix()))
		bf.WriteUint32(feature.ActiveFeatures)
		bf.WriteUint16(0)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func generateFeatureWeapons(count int) activeFeature {
	nums := make([]int, 0)
	var result int
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		num := r.Intn(14)
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
	WeekReq, WeekCount uint8
	Available          bool
	Expiration         uint32
}

func handleMsgMhfGetKeepLoginBoostStatus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKeepLoginBoostStatus)

	var loginBoostStatus []loginBoost
	insert := false
	boostState, err := s.server.db.Query("SELECT week_req, week_count, available, end_time FROM login_boost_state WHERE char_id=$1 ORDER BY week_req ASC", s.charID)
	if err != nil {
		panic(err)
	}
	for boostState.Next() {
		var boost loginBoost
		err = boostState.Scan(&boost.WeekReq, &boost.WeekCount, &boost.Available, &boost.Expiration)
		if err != nil {
			panic(err)
		}
		loginBoostStatus = append(loginBoostStatus, boost)
	}
	if len(loginBoostStatus) == 0 {
		// create default Entries (should only been week 1 with )
		insert = true
		loginBoostStatus = []loginBoost{
			{
				WeekReq:    1,    // weeks needed
				WeekCount:  0,    // weeks passed
				Available:  true, // available
				Expiration: 0,    //uint32(t.Add(120 * time.Minute).Unix()), // uncomment to enable permanently
			},
			{
				WeekReq:    2,
				WeekCount:  0,
				Available:  true,
				Expiration: 0,
			},
			{
				WeekReq:    3,
				WeekCount:  0,
				Available:  true,
				Expiration: 0,
			},
			{
				WeekReq:    4,
				WeekCount:  0,
				Available:  true,
				Expiration: 0,
			},
			{
				WeekReq:    5,
				WeekCount:  0,
				Available:  true,
				Expiration: 0,
			},
		}
	}
	resp := byteframe.NewByteFrame()
	CurrentWeek := Time_Current_Week_uint8()
	for d := range loginBoostStatus {
		if CurrentWeek == 1 && loginBoostStatus[d].WeekCount <= 5 {
			loginBoostStatus[d].WeekCount = 0
		}
		if loginBoostStatus[d].WeekReq == CurrentWeek || loginBoostStatus[d].WeekCount != 0 {
			loginBoostStatus[d].WeekCount = CurrentWeek
		}
		if !loginBoostStatus[d].Available && loginBoostStatus[d].WeekCount >= loginBoostStatus[d].WeekReq && uint32(time.Now().In(time.FixedZone("UTC+1", 1*60*60)).Unix()) >= loginBoostStatus[d].Expiration {
			loginBoostStatus[d].Expiration = 1
		}
		if !insert {
			_, err := s.server.db.Exec(`UPDATE login_boost_state SET week_count=$1, end_time=$2 WHERE char_id=$3 AND week_req=$4`, loginBoostStatus[d].WeekCount, loginBoostStatus[d].Expiration, s.charID, loginBoostStatus[d].WeekReq)
			if err != nil {
				panic(err)
			}
		}
	}
	for _, v := range loginBoostStatus {
		if insert {
			_, err := s.server.db.Exec(`INSERT INTO login_boost_state (char_id, week_req, week_count, available, end_time) VALUES ($1,$2,$3,$4,$5)`, s.charID, v.WeekReq, v.WeekCount, v.Available, v.Expiration)
			if err != nil {
				panic(err)
			}
		}
		resp.WriteUint8(v.WeekReq)
		resp.WriteUint8(v.WeekCount)
		resp.WriteBool(v.Available)
		resp.WriteUint32(v.Expiration)
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfUseKeepLoginBoost(s *Session, p mhfpacket.MHFPacket) {
	// Directly interacts with MhfGetKeepLoginBoostStatus
	// TODO: make these states persistent on a per character basis
	pkt := p.(*mhfpacket.MsgMhfUseKeepLoginBoost)
	var t = time.Now().In(time.FixedZone("UTC+1", 1*60*60))
	resp := byteframe.NewByteFrame()
	resp.WriteUint8(0)

	// response is end timestamp based on input
	switch pkt.BoostWeekUsed {
	case 1:
		t = t.Add(120 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	case 2:
		t = t.Add(240 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	case 3:
		t = t.Add(120 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	case 4:
		t = t.Add(180 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	case 5:
		t = t.Add(240 * time.Minute)
		resp.WriteUint32(uint32(t.Unix()))
	}
	_, err := s.server.db.Exec(`UPDATE login_boost_state SET available='false', end_time=$1 WHERE char_id=$2 AND week_req=$3`, uint32(t.Unix()), s.charID, pkt.BoostWeekUsed)
	if err != nil {
		panic(err)
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetRestrictionEvent(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetRestrictionEvent)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
