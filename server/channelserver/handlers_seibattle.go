package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"

	"github.com/jmoiron/sqlx"
)

func handleMsgMhfGetBreakSeibatuLevelReward(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBreakSeibatuLevelReward)
	bf := byteframe.NewByteFrame()
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

type WeeklySeibatuRankingReward struct {
	Unk0 int32
	Unk1 int32
	Unk2 uint32
	Unk3 int32
	Unk4 int32
	Unk5 int32
}

func handleMsgMhfGetWeeklySeibatuRankingReward(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySeibatuRankingReward)
	var data []*byteframe.ByteFrame
	weeklySeibatuRankingRewards := []WeeklySeibatuRankingReward{
		{0, 0, 0, 0, 0, 0},
	}
	for _, reward := range weeklySeibatuRankingRewards {
		bf := byteframe.NewByteFrame()
		bf.WriteInt32(reward.Unk0)
		bf.WriteInt32(reward.Unk1)
		bf.WriteUint32(reward.Unk2)
		bf.WriteInt32(reward.Unk3)
		bf.WriteInt32(reward.Unk4)
		bf.WriteInt32(reward.Unk5)
		data = append(data, bf)
	}
	s.DoAckEarthSucceed(pkt.AckHandle, data)
}

func handleMsgMhfGetFixedSeibatuRankingTable(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetFixedSeibatuRankingTable)
	bf := byteframe.NewByteFrame()
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteBytes(make([]byte, 32))
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfReadBeatLevel(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadBeatLevel)

	// This response is fixed and will never change on JP,
	// but I've left it dynamic for possible other client differences.
	resp := byteframe.NewByteFrame()
	for i := 0; i < int(pkt.ValidIDCount); i++ {
		resp.WriteUint32(pkt.IDs[i])
		resp.WriteUint32(1)
		resp.WriteUint32(1)
		resp.WriteUint32(1)
	}

	s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
}

func handleMsgMhfReadLastWeekBeatRanking(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadLastWeekBeatRanking)
	bf := byteframe.NewByteFrame()
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateBeatLevel(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateBeatLevel)

	s.DoAckBufSucceed(pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfReadBeatLevelAllRanking(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadBeatLevelAllRanking)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)

	for i := 0; i < 100; i++ {
		bf.WriteUint32(0)
		bf.WriteUint32(0)
		bf.WriteBytes(make([]byte, 32))
	}
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}

func handleMsgMhfReadBeatLevelMyRanking(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadBeatLevelMyRanking)
	bf := byteframe.NewByteFrame()
	s.DoAckBufSucceed(pkt.AckHandle, bf.Data())
}
