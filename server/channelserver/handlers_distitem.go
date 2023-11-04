package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network/mhfpacket"
	"time"

	"go.uber.org/zap"
)

type ItemDist struct {
	ID              uint32    `db:"id"`
	Deadline        time.Time `db:"deadline"`
	TimesAcceptable uint16    `db:"times_acceptable"`
	TimesAccepted   uint16    `db:"times_accepted"`
	MinHR           uint16    `db:"min_hr"`
	MaxHR           uint16    `db:"max_hr"`
	MinSR           uint16    `db:"min_sr"`
	MaxSR           uint16    `db:"max_sr"`
	MinGR           uint16    `db:"min_gr"`
	MaxGR           uint16    `db:"max_gr"`
	EventName       string    `db:"event_name"`
	Description     string    `db:"description"`
	Data            []byte    `db:"data"`
}

func handleMsgMhfEnumerateDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateDistItem)
	bf := byteframe.NewByteFrame()
	distCount := 0
	dists, err := s.server.db.Queryx(`
		SELECT d.id, event_name, description, times_acceptable,
		min_hr, max_hr, min_sr, max_sr, min_gr, max_gr,
		(
    	SELECT count(*)
    	FROM distributions_accepted da
    	WHERE d.id = da.distribution_id
    	AND da.character_id = $1
		) AS times_accepted,
		COALESCE(deadline, TO_TIMESTAMP(0)) AS deadline
		FROM distribution d
		WHERE character_id = $1 AND type = $2 OR character_id IS NULL AND type = $2 ORDER BY id DESC;
	`, s.charID, pkt.DistType)
	if err != nil {
		s.logger.Error("Error getting distribution data from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
	} else {
		for dists.Next() {
			distCount++
			distData := &ItemDist{}
			err = dists.StructScan(&distData)
			if err != nil {
				s.logger.Error("Error parsing item distribution data", zap.Error(err))
			}
			bf.WriteUint32(distData.ID)
			bf.WriteUint32(uint32(distData.Deadline.Unix()))
			bf.WriteUint32(0) // Unk
			bf.WriteUint16(distData.TimesAcceptable)
			bf.WriteUint16(distData.TimesAccepted)
			bf.WriteUint16(0) // Unk
			bf.WriteUint16(distData.MinHR)
			bf.WriteUint16(distData.MaxHR)
			bf.WriteUint16(distData.MinSR)
			bf.WriteUint16(distData.MaxSR)
			bf.WriteUint16(distData.MinGR)
			bf.WriteUint16(distData.MaxGR)
			bf.WriteUint8(0)
			bf.WriteUint16(0)
			bf.WriteUint8(0)
			bf.WriteUint16(0)
			bf.WriteUint16(0)
			bf.WriteUint8(0)
			ps.Uint8(bf, distData.EventName, true)
			for i := 0; i < 6; i++ {
				for j := 0; j < 13; j++ {
					bf.WriteUint8(0)
					bf.WriteUint32(0)
				}
			}
			i := uint8(0)
			bf.WriteUint8(i)
			if i <= 10 {
				for j := uint8(0); j < i; j++ {
					bf.WriteUint32(0)
					bf.WriteUint32(0)
					bf.WriteUint32(0)
				}
			}
		}
		resp := byteframe.NewByteFrame()
		resp.WriteUint16(uint16(distCount))
		resp.WriteBytes(bf.Data())
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	}
}

type DistributionItem struct {
	ItemType uint8  `db:"item_type"`
	ID       uint32 `db:"id"`
	ItemID   uint32 `db:"item_id"`
	Quantity uint32 `db:"quantity"`
}

func handleMsgMhfApplyDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfApplyDistItem)

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(pkt.DistributionID)
	var distItems []DistributionItem
	rows, err := s.server.db.Queryx(`SELECT id, item_id, item_type, quantity FROM distribution_items WHERE distribution_id=$1`, pkt.DistributionID)
	if err == nil {
		var distItem DistributionItem
		for rows.Next() {
			err = rows.StructScan(&distItem)
			if err != nil {
				continue
			}
			distItems = append(distItems, distItem)
		}
	}
	bf.WriteUint16(uint16(len(distItems)))
	for _, item := range distItems {
		bf.WriteUint8(item.ItemType)
		bf.WriteUint32(item.ItemID)
		bf.WriteUint32(item.Quantity)
		bf.WriteUint32(item.ID)
		switch item.ItemType {
		case 17:
			_ = addPointNetcafe(s, int(item.Quantity))
		case 19:
			s.server.db.Exec("UPDATE users u SET gacha_premium=gacha_premium+$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", item.Quantity, s.charID)
		case 20:
			s.server.db.Exec("UPDATE users u SET gacha_trial=gacha_trial+$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", item.Quantity, s.charID)
		case 21:
			s.server.db.Exec("UPDATE users u SET frontier_points=frontier_points+$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", item.Quantity, s.charID)
		case 23:
			saveData, err := GetCharacterSaveData(s, s.charID)
			if err == nil {
				saveData.RP += uint16(item.Quantity)
				saveData.Save(s)
			}
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())

	if pkt.DistributionID > 0 {
		_, err = s.server.db.Exec(`INSERT INTO public.distributions_accepted VALUES ($1, $2)`, pkt.DistributionID, s.charID)
	}
}

func handleMsgMhfAcquireDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireDistItem)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetDistDescription(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetDistDescription)
	var desc string
	err := s.server.db.QueryRow("SELECT description FROM distribution WHERE id = $1", pkt.DistributionID).Scan(&desc)
	if err != nil {
		s.logger.Error("Error parsing item distribution description", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf := byteframe.NewByteFrame()
	ps.Uint16(bf, desc, true)
	ps.Uint16(bf, "", false)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
