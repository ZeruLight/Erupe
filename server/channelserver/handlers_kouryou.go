package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/db"
	"fmt"

	"go.uber.org/zap"
)

func handleMsgMhfAddKouryouPoint(s *Session, p mhfpacket.MHFPacket) {
	// hunting with both ranks maxed gets you these
	pkt := p.(*mhfpacket.MsgMhfAddKouryouPoint)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var points int
	err = database.QueryRow("UPDATE characters SET kouryou_point=COALESCE(kouryou_point + $1, $1) WHERE id=$2 RETURNING kouryou_point", pkt.KouryouPoints, s.CharID).Scan(&points)
	if err != nil {
		s.Logger.Error("Failed to update KouryouPoint in db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(points))
	DoAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfGetKouryouPoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetKouryouPoint)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var points int
	err = database.QueryRow("SELECT COALESCE(kouryou_point, 0) FROM characters WHERE id = $1", s.CharID).Scan(&points)
	if err != nil {
		s.Logger.Error("Failed to get kouryou_point savedata from db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(points))
	DoAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfExchangeKouryouPoint(s *Session, p mhfpacket.MHFPacket) {
	// spent at the guildmaster, 10000 a roll
	var points int
	pkt := p.(*mhfpacket.MsgMhfExchangeKouryouPoint)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	err = database.QueryRow("UPDATE characters SET kouryou_point=kouryou_point - $1 WHERE id=$2 RETURNING kouryou_point", pkt.KouryouPoints, s.CharID).Scan(&points)
	if err != nil {
		s.Logger.Error("Failed to update platemyset savedata in db", zap.Error(err))
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(points))
	DoAckBufSucceed(s, pkt.AckHandle, resp.Data())
}
