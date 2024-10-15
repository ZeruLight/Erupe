package channelserver

import (
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/db"
	"erupe-ce/utils/gametime"
	"erupe-ce/utils/mhfcourse"

	"erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	ps "erupe-ce/utils/pascalstring"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"
)

func handleMsgMhfAcquireCafeItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireCafeItem)
	var netcafePoints uint32
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("UPDATE characters SET netcafe_points = netcafe_points - $1 WHERE id = $2 RETURNING netcafe_points", pkt.PointCost, s.CharID).Scan(&netcafePoints)
	if err != nil {
		s.Logger.Error("Failed to get netcafe points from db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(netcafePoints)
	s.DoAckSimpleSucceed(pkt.AckHandle, resp.Data())
}

func handleMsgMhfUpdateCafepoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateCafepoint)
	var netcafePoints uint32
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("SELECT COALESCE(netcafe_points, 0) FROM characters WHERE id = $1", s.CharID).Scan(&netcafePoints)
	if err != nil {
		s.Logger.Error("Failed to get netcate points from db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(netcafePoints)
	s.DoAckSimpleSucceed(pkt.AckHandle, resp.Data())
}

func handleMsgMhfCheckDailyCafepoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckDailyCafepoint)

	midday := gametime.TimeMidnight().Add(12 * time.Hour)
	if gametime.TimeAdjusted().After(midday) {
		midday = midday.Add(24 * time.Hour)
	}

	// get time after which daily claiming would be valid from db
	var dailyTime time.Time
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("SELECT COALESCE(daily_time, $2) FROM characters WHERE id = $1", s.CharID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)).Scan(&dailyTime)
	if err != nil {
		s.Logger.Error("Failed to get daily_time savedata from db", zap.Error(err))
	}

	var bondBonus, bonusQuests, dailyQuests uint32
	bf := byteframe.NewByteFrame()
	if midday.After(dailyTime) {
		addPointNetcafe(s, 5)
		bondBonus = 5 // Bond point bonus quests
		bonusQuests = config.GetConfig().GameplayOptions.BonusQuestAllowance
		dailyQuests = config.GetConfig().GameplayOptions.DailyQuestAllowance
		database.Exec("UPDATE characters SET daily_time=$1, bonus_quests = $2, daily_quests = $3 WHERE id=$4", midday, bonusQuests, dailyQuests, s.CharID)
		bf.WriteBool(true) // Success?
	} else {
		bf.WriteBool(false)
	}
	bf.WriteUint32(bondBonus)
	bf.WriteUint32(bonusQuests)
	bf.WriteUint32(dailyQuests)
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetCafeDuration(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetCafeDuration)
	bf := byteframe.NewByteFrame()
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var cafeReset time.Time
	err = database.QueryRow(`SELECT cafe_reset FROM characters WHERE id=$1`, s.CharID).Scan(&cafeReset)
	if err != nil {
		cafeReset = gametime.TimeWeekNext()
		database.Exec(`UPDATE characters SET cafe_reset=$1 WHERE id=$2`, cafeReset, s.CharID)
	}
	if gametime.TimeAdjusted().After(cafeReset) {
		cafeReset = gametime.TimeWeekNext()
		database.Exec(`UPDATE characters SET cafe_time=0, cafe_reset=$1 WHERE id=$2`, cafeReset, s.CharID)
		database.Exec(`DELETE FROM cafe_accepted WHERE character_id=$1`, s.CharID)
	}

	var cafeTime uint32
	err = database.QueryRow("SELECT cafe_time FROM characters WHERE id = $1", s.CharID).Scan(&cafeTime)
	if err != nil {
		panic(err)
	}
	if mhfcourse.CourseExists(30, s.courses) {
		cafeTime = uint32(gametime.TimeAdjusted().Unix()) - uint32(s.sessionStart) + cafeTime
	}
	bf.WriteUint32(cafeTime)
	if config.GetConfig().ClientID >= config.ZZ {
		bf.WriteUint16(0)
		ps.Uint16(bf, fmt.Sprintf(s.Server.i18n.cafe.reset, int(cafeReset.Month()), cafeReset.Day()), true)
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

type CafeBonus struct {
	ID       uint32 `db:"id"`
	TimeReq  uint32 `db:"time_req"`
	ItemType uint32 `db:"item_type"`
	ItemID   uint32 `db:"item_id"`
	Quantity uint32 `db:"quantity"`
	Claimed  bool   `db:"claimed"`
}

func handleMsgMhfGetCafeDurationBonusInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetCafeDurationBonusInfo)
	bf := byteframe.NewByteFrame()
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var count uint32
	rows, err := database.Queryx(`
	SELECT cb.id, time_req, item_type, item_id, quantity,
	(
		SELECT count(*)
		FROM cafe_accepted ca
		WHERE cb.id = ca.cafe_id AND ca.character_id = $1
	)::int::bool AS claimed
	FROM cafebonus cb ORDER BY id ASC;`, s.CharID)
	if err != nil {
		s.Logger.Error("Error getting cafebonus", zap.Error(err))
		s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 4))
	} else {
		for rows.Next() {
			count++
			cafeBonus := &CafeBonus{}
			err = rows.StructScan(&cafeBonus)
			if err != nil {
				s.Logger.Error("Error scanning cafebonus", zap.Error(err))
			}
			bf.WriteUint32(cafeBonus.TimeReq)
			bf.WriteUint32(cafeBonus.ItemType)
			bf.WriteUint32(cafeBonus.ItemID)
			bf.WriteUint32(cafeBonus.Quantity)
			bf.WriteBool(cafeBonus.Claimed)
		}
		resp := byteframe.NewByteFrame()
		resp.WriteUint32(0)
		resp.WriteUint32(uint32(gametime.TimeAdjusted().Unix()))
		resp.WriteUint32(count)
		resp.WriteBytes(bf.Data())
		s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
	}
}

func handleMsgMhfReceiveCafeDurationBonus(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReceiveCafeDurationBonus)
	bf := byteframe.NewByteFrame()
	var count uint32
	bf.WriteUint32(0)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	rows, err := database.Queryx(`
	SELECT c.id, time_req, item_type, item_id, quantity
	FROM cafebonus c
	WHERE (
		SELECT count(*)
		FROM cafe_accepted ca
		WHERE c.id = ca.cafe_id AND ca.character_id = $1
	) < 1 AND (
		SELECT ch.cafe_time + $2
		FROM characters ch
		WHERE ch.id = $1 
	) >= time_req`, s.CharID, gametime.TimeAdjusted().Unix()-s.sessionStart)
	if err != nil {
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
	} else {
		for rows.Next() {
			cafeBonus := &CafeBonus{}
			err = rows.StructScan(cafeBonus)
			if err != nil {
				continue
			}
			count++
			bf.WriteUint32(cafeBonus.ID)
			bf.WriteUint32(cafeBonus.ItemType)
			bf.WriteUint32(cafeBonus.ItemID)
			bf.WriteUint32(cafeBonus.Quantity)
		}
		bf.Seek(0, io.SeekStart)
		bf.WriteUint32(count)
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
	}
}

func handleMsgMhfPostCafeDurationBonusReceived(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostCafeDurationBonusReceived)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var cafeBonus CafeBonus
	for _, cbID := range pkt.CafeBonusID {
		err := database.QueryRow(`
		SELECT cb.id, item_type, quantity FROM cafebonus cb WHERE cb.id=$1
		`, cbID).Scan(&cafeBonus.ID, &cafeBonus.ItemType, &cafeBonus.Quantity)
		if err == nil {
			if cafeBonus.ItemType == 17 {
				addPointNetcafe(s, int(cafeBonus.Quantity))
			}
		}
		database.Exec("INSERT INTO public.cafe_accepted VALUES ($1, $2)", cbID, s.CharID)
	}
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func addPointNetcafe(s *Session, p int) error {
	var points int
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("SELECT netcafe_points FROM characters WHERE id = $1", s.CharID).Scan(&points)
	if err != nil {
		return err
	}
	if points+p > config.GetConfig().GameplayOptions.MaximumNP {
		points = config.GetConfig().GameplayOptions.MaximumNP
	} else {
		points += p
	}
	database.Exec("UPDATE characters SET netcafe_points=$1 WHERE id=$2", points, s.CharID)
	return nil
}

func handleMsgMhfStartBoostTime(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStartBoostTime)
	bf := byteframe.NewByteFrame()
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	boostLimit := gametime.TimeAdjusted().Add(time.Duration(config.GetConfig().GameplayOptions.BoostTimeDuration) * time.Second)
	if config.GetConfig().GameplayOptions.DisableBoostTime {
		bf.WriteUint32(0)
		s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
		return
	}
	database.Exec("UPDATE characters SET boost_time=$1 WHERE id=$2", boostLimit, s.CharID)
	bf.WriteUint32(uint32(boostLimit.Unix()))
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetBoostTime(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostTime)
	s.DoAckBufSucceed(pkt.AckHandle, []byte{})
}

func handleMsgMhfGetBoostTimeLimit(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostTimeLimit)
	bf := byteframe.NewByteFrame()
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var boostLimit time.Time
	err = database.QueryRow("SELECT boost_time FROM characters WHERE id=$1", s.CharID).Scan(&boostLimit)
	if err != nil {
		bf.WriteUint32(0)
	} else {
		bf.WriteUint32(uint32(boostLimit.Unix()))
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetBoostRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBoostRight)
	var boostLimit time.Time
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("SELECT boost_time FROM characters WHERE id=$1", s.CharID).Scan(&boostLimit)
	if err != nil {
		s.DoAckBufSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
		return
	}
	if boostLimit.After(gametime.TimeAdjusted()) {
		s.DoAckBufSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
	} else {
		s.DoAckBufSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x02})
	}
}

func handleMsgMhfPostBoostTimeQuestReturn(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostBoostTimeQuestReturn)
	s.DoAckSimpleSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfPostBoostTime(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostBoostTime)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfPostBoostTimeLimit(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostBoostTimeLimit)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}
