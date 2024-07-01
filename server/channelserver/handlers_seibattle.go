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
	Index0 int32  //Place Start
	Index1 int32  //Place Finish
	Index2 uint32 // UNK
	Type   int32  //Type //7201 Value  //7202 ??? Points  //7203 ??? Points Blue
	ID     int32  //ID
	Value  int32  // Value
}

func handleMsgMhfGetWeeklySeibatuRankingReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySeibatuRankingReward)
	var data []*byteframe.ByteFrame
	var weeklySeibatuRankingRewards []WeeklySeibatuRankingReward
	switch pkt.Operation {
	case 3:
		weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{

			//Unk0
			//Unk1
			//Unk2
			//Unk3,
			//ROUTE, (Crashes if it doesnt exist be careful with values )
			//Status 1 = Only Now !  2= Unk 3= Disabled}

			//Route 0
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 1
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
			//Route 2
			{0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0},
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
		switch pkt.ID {
		//243400 = Route 0
		//243401 = Route 1
		//I have a sneaky suspicion that the above massive array is feeding into this somehow....
		case 240031:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{1, 1, 1, 7201, 12068, 1}}
		case 240041:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{0, 0, 1, 7201, 12068, 1}}
		case 240042:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{0, 0, 2, 7201, 12068, 1}}
		case 240051:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{0, 0, 1, 7201, 12068, 1}}
		case 240052:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				{1, 1, 1, 7201, 12068, 1},
			}
		//Tower 260001 260003
		case 260001:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{

				//Can only have 10 in each dist (It disapears otherwise)
				//{unk,unk,dist,seiabtuType,ItemID,Value}
				{0, 0, 1, 7201, 12068, 1},
				{0, 0, 1, 7201, 12069, 1},
				{0, 0, 1, 7201, 12070, 1},
				{0, 0, 1, 7201, 12071, 1},
				{0, 0, 1, 7201, 12072, 1},
				{0, 0, 1, 7201, 12073, 1},
				{0, 0, 1, 7201, 12074, 1},
				{0, 0, 1, 7201, 12075, 1},
				{0, 0, 1, 7201, 12076, 1},
				{0, 0, 1, 7201, 12077, 1},

				{0, 0, 2, 7201, 12068, 1},
				{0, 0, 2, 7201, 12069, 1},
				{0, 0, 2, 7201, 12070, 1},
				{0, 0, 2, 7201, 12071, 1},
				{0, 0, 2, 7201, 12072, 1},
				{0, 0, 2, 7201, 12073, 1},
				{0, 0, 2, 7201, 12074, 1},
				{0, 0, 2, 7201, 12075, 1},
				{0, 0, 2, 7201, 12076, 1},
				{0, 0, 2, 7201, 12077, 1},
				// Left in because i think its funny the planned 4 and we got 2
				{0, 0, 3, 7201, 12068, 1},
				{0, 0, 3, 7201, 12069, 1},
				{0, 0, 3, 7201, 12070, 1},
				{0, 0, 3, 7201, 12071, 1},
				{0, 0, 3, 7201, 12072, 1},
				{0, 0, 3, 7201, 12073, 1},
				{0, 0, 3, 7201, 12074, 1},
				{0, 0, 3, 7201, 12075, 1},
				{0, 0, 3, 7201, 12076, 1},
				{0, 0, 3, 7201, 12077, 1},

				{0, 0, 4, 7201, 12068, 1},
				{0, 0, 4, 7201, 12069, 1},
				{0, 0, 4, 7201, 12070, 1},
				{0, 0, 4, 7201, 12071, 1},
				{0, 0, 4, 7201, 12072, 1},
				{0, 0, 4, 7201, 12073, 1},
				{0, 0, 4, 7201, 12074, 1},
				{0, 0, 4, 7201, 12075, 1},
				{0, 0, 4, 7201, 12076, 1},
				{0, 0, 4, 7201, 12077, 1},
			}
		case 260003:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				//Adjust Floors done in database to make blue ?? Possible value here for dist
				//{Floor,unk,unk,seiabtuType,ItemID,Value}

				{1, 0, 0, 7201, 12068, 1},
				{2, 0, 0, 7201, 12069, 3},
				{2, 0, 0, 7201, 12070, 1},
				{4, 0, 0, 7201, 12071, 3},
				{5, 0, 0, 7201, 12072, 6},
				{6, 0, 0, 7201, 12073, 1},
				{7, 0, 0, 7201, 12068, 1},
				{8, 0, 0, 7201, 12069, 1},
				{9, 0, 0, 7201, 12070, 2},
				{10, 0, 0, 7201, 12071, 1},
				{10, 0, 0, 7201, 12072, 1},
				{10, 0, 0, 7201, 12073, 6},

				{11, 0, 0, 7201, 12072, 1},
				{12, 0, 0, 7201, 12073, 1},
				{13, 0, 0, 7201, 12068, 6},
				{14, 0, 0, 7201, 12069, 6},
				{15, 0, 0, 7201, 12070, 1},
				{16, 0, 0, 7201, 12071, 1},
				{16, 0, 0, 7201, 12072, 2},
				{18, 0, 0, 7201, 12073, 1},
			}
		default: //Covers all Pallone Requests... for now
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward{
				// To do figure out values 3-5 its some sort of item structure
				//1st
				{1, 0, 0, 7202, 10, 10000},
				{1, 1, 0, 7201, 10, 30},
				{1, 1, 0, 7201, 10, 18},
				{1, 1, 0, 7201, 10, 18},
				//2nd - 3rd
				{2, 3, 0, 7202, 10, 6000},
				{2, 3, 0, 7201, 10, 15},
				{2, 3, 0, 7201, 10, 9},
				{2, 3, 0, 7201, 10, 9},
				//4th -10th
				{4, 10, 0, 7202, 10, 5500},
				{4, 10, 0, 7201, 10, 12},
				{4, 10, 0, 7201, 10, 9},
			}

		}

	}
	for _, seibatuData := range weeklySeibatuRankingRewards {
		bf := byteframe.NewByteFrame()
		bf.WriteInt32(seibatuData.Index0)
		bf.WriteInt32(seibatuData.Index1)
		bf.WriteUint32(seibatuData.Index2)
		bf.WriteInt32(seibatuData.Type)
		bf.WriteInt32(seibatuData.ID)
		bf.WriteInt32(seibatuData.Value)
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
