package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func handleMsgMhfAddKouryouPoint(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	// hunting with both ranks maxed gets you these
	pkt := p.(*mhfpacket.MsgMhfAddKouryouPoint)

	var points int
	err := db.QueryRow("UPDATE characters SET kouryou_point=COALESCE(kouryou_point + $1, $1) WHERE id=$2 RETURNING kouryou_point", pkt.KouryouPoints, s.CharID).Scan(&points)
	if err != nil {
		s.Logger.Error("Failed to update KouryouPoint in db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(points))
	s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetKouryouPoint(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKouryouPoint)

	var points int
	err := db.QueryRow("SELECT COALESCE(kouryou_point, 0) FROM characters WHERE id = $1", s.CharID).Scan(&points)
	if err != nil {
		s.Logger.Error("Failed to get kouryou_point savedata from db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(points))
	s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
}

func handleMsgMhfExchangeKouryouPoint(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	// spent at the guildmaster, 10000 a roll
	var points int
	pkt := p.(*mhfpacket.MsgMhfExchangeKouryouPoint)

	err := db.QueryRow("UPDATE characters SET kouryou_point=kouryou_point - $1 WHERE id=$2 RETURNING kouryou_point", pkt.KouryouPoints, s.CharID).Scan(&points)
	if err != nil {
		s.Logger.Error("Failed to update platemyset savedata in db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(points))
	s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
}
