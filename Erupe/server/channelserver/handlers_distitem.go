package channelserver

import (
	"time"

	"erupe-ce/network/mhfpacket"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/byteframe"
	"go.uber.org/zap"
)

type ItemDist struct {
	ID              uint32 `db:"id"`
	Deadline        uint32 `db:"deadline"`
	TimesAcceptable uint16 `db:"times_acceptable"`
	TimesAccepted   uint16 `db:"times_accepted"`
	MinHR           uint16 `db:"min_hr"`
	MaxHR           uint16 `db:"max_hr"`
	MinSR           uint16 `db:"min_sr"`
	MaxSR           uint16 `db:"max_sr"`
	MinGR           uint16 `db:"min_gr"`
	MaxGR           uint16 `db:"max_gr"`
	EventName       string `db:"event_name"`
	Description     string `db:"description"`
	Data            []byte `db:"data"`
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
		CASE
			WHEN (EXTRACT(epoch FROM deadline)::int) IS NULL THEN 0
			ELSE (EXTRACT(epoch FROM deadline)::int)
		END deadline
		FROM distribution d
		WHERE character_id = $1 AND type = $2 OR character_id IS NULL AND type = $2 ORDER BY id DESC;
	`, s.charID, pkt.Unk0)
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
			bf.WriteUint32(distData.Deadline)
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
			bf.WriteUint32(0) // Unk
			bf.WriteUint32(0) // Unk
			ps.Uint16(bf, distData.EventName, true)
			bf.WriteBytes(make([]byte, 391))
		}
		resp := byteframe.NewByteFrame()
		resp.WriteUint16(uint16(distCount))
		resp.WriteBytes(bf.Data())
		resp.WriteUint8(0)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	}
}

func handleMsgMhfApplyDistItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfApplyDistItem)

  if pkt.DistributionID == 0 {
    doAckBufSucceed(s, pkt.AckHandle, make([]byte, 6))
  } else {
		row := s.server.db.QueryRowx("SELECT data FROM distribution WHERE id = $1", pkt.DistributionID)
		dist := &ItemDist{}
		err := row.StructScan(dist)
		if err != nil {
			s.logger.Error("Error parsing item distribution data", zap.Error(err))
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 6))
			return
		}

		bf := byteframe.NewByteFrame()
		bf.WriteUint32(pkt.DistributionID)
		bf.WriteBytes(dist.Data)
		bf.WriteUint32(uint32(time.Now().Unix()))
    doAckBufSucceed(s, pkt.AckHandle, bf.Data())

		_, err = s.server.db.Exec(`
			INSERT INTO public.distributions_accepted
			VALUES ($1, $2)
		`, pkt.DistributionID, s.charID)
		if err != nil {
			s.logger.Error("Error updating accepted dist count", zap.Error(err))
		}
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
