package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func handleMsgMhfGetBreakSeibatuLevelReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBreakSeibatuLevelReward)
	bf := byteframe.NewByteFrame()
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

type WeeklySeibatuRankingReward struct {
	Unk0 int32  //Place Start
	Unk1 int32  //Place Finish
	Unk2 uint32 // UNK
	Unk3 int32  //Type
	Unk4 int32  //ID
	Unk5 int32  // Value
}

func handleMsgMhfGetWeeklySeibatuRankingReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySeibatuRankingReward)
	var data []*byteframe.ByteFrame
	var weeklySeibatuRankingRewards []WeeklySeibatuRankingReward
	switch pkt.Operation {
	case 3:
		weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
			//Route 0
			{0, 0, 0, 0, 1, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 1
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 2
			{0, 0, 0, 0, 5, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 3
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 4
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 5
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 6
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 7
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 8
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 9
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 10
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
		}

		// 0 = Max 7 Routes so value 6
		//ZZ looks like it only works up to Route 2

	case 5:
		// Unk1 = 5 and unk2 = 240001
		//unk2 = 243400 = Route 0
		//unk3 = 243401 = Route 1
		//Tower 260001 260003
		switch pkt.ID {
		case 240031:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{2, 5, 5, 5, 5, 5}, {0, 0, 0, 0, 0, 0}}
		case 240041:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{2, 5, 5, 5, 5, 5}, {0, 0, 0, 0, 0, 0}}

		case 240042:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{2, 5, 5, 5, 5, 5}, {0, 0, 0, 0, 0, 0}}

		case 240051:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{2, 5, 5, 5, 5, 5}, {0, 0, 0, 0, 0, 0}}

		case 260001:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{2, 5, 5, 5, 5, 5}, {0, 0, 0, 0, 0, 0}}

		case 260003:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{2, 5, 5, 5, 5, 5}, {0, 0, 0, 0, 0, 0}}

		default: //Covers all Pallone Requests... for now
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				// To do figure out values 3-5 its some sort of item structure
				{1, 0, 7, 7, 7, 10000}, {1, 1, 0, 0, 0, 30}, {1, 1, 0, 0, 0, 18}, {1, 1, 0, 0, 0, 18}, //1st
				{2, 3, 0, 0, 0, 6000}, {2, 3, 0, 0, 0, 15}, {2, 3, 0, 0, 0, 9}, {2, 3, 0, 0, 0, 9}, //2nd - 3rd
				{4, 10, 0, 0, 0, 5500}, {4, 10, 0, 0, 0, 12}, {4, 10, 0, 0, 0, 9}, //4th -10th
			}

		}

	}
	for _, rank := range weeklySeibatuRankingRewards {
		bf := byteframe.NewByteFrame()
		bf.WriteInt32(rank.Unk0)
		bf.WriteInt32(rank.Unk1)
		bf.WriteUint32(rank.Unk2)
		bf.WriteInt32(rank.Unk3)
		bf.WriteInt32(rank.Unk4)
		bf.WriteInt32(rank.Unk5)
		data = append(data, bf)
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfGetFixedSeibatuRankingTable(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetFixedSeibatuRankingTable)
	bf := byteframe.NewByteFrame()
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteBytes(make([]byte, 32))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfReadBeatLevel(s *Session, p mhfpacket.MHFPacket) {
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

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfReadLastWeekBeatRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadLastWeekBeatRanking)
	bf := byteframe.NewByteFrame()
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	bf.WriteInt32(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateBeatLevel(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateBeatLevel)

	doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfReadBeatLevelAllRanking(s *Session, p mhfpacket.MHFPacket) {
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
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfReadBeatLevelMyRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadBeatLevelMyRanking)
	bf := byteframe.NewByteFrame()
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
