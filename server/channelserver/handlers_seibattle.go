package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
)

type BreakSeibatuLevelReward struct {
	Item  int32
	Value int32
	Level int32
	Unk   int32
}

func handleMsgMhfGetBreakSeibatuLevelReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetBreakSeibatuLevelReward)
	var data []*byteframe.ByteFrame
	var weeklySeibatuRankingRewards []BreakSeibatuLevelReward

	switch pkt.EarthMonster {
	case 116:
		weeklySeibatuRankingRewards = []BreakSeibatuLevelReward{
			{8, 3, 2, 0},
			{8, 3, 2, 0},
			{8, 3, 2, 0},
			{8, 3, 3, 0},
			{8, 3, 3, 0},
			{8, 3, 3, 0},
			{8, 3, 3, 0}}
	case 107:
		weeklySeibatuRankingRewards = []BreakSeibatuLevelReward{
			{4, 3, 1, 0},
			{4, 3, 2, 0},
			{4, 3, 3, 0},
			{4, 3, 4, 0},
			{4, 3, 5, 0}}
	case 2:
		weeklySeibatuRankingRewards = []BreakSeibatuLevelReward{
			{5, 3, 1, 0},
			{5, 3, 2, 0},
			{5, 3, 3, 0},
			{5, 3, 4, 0},
			{5, 3, 5, 0}}

	case 36:
		weeklySeibatuRankingRewards = []BreakSeibatuLevelReward{
			{7, 3, 1, 0},
			{7, 3, 2, 0},
			{7, 3, 3, 0},
			{7, 3, 4, 0},
			{7, 3, 5, 0}}

	default:
		weeklySeibatuRankingRewards = []BreakSeibatuLevelReward{
			{1, 3, 1, 0},
			{1, 3, 2, 0},
			{1, 3, 3, 0},
			{1, 3, 4, 0},
			{1, 3, 5, 0}}
	}

	for _, seibatuData := range weeklySeibatuRankingRewards {
		bf := byteframe.NewByteFrame()

		bf.WriteInt32(seibatuData.Item)  // Item
		bf.WriteInt32(seibatuData.Value) // Value
		bf.WriteInt32(seibatuData.Level) //Level
		bf.WriteInt32(seibatuData.Unk)
		data = append(data, bf)
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

type WeeklySeibatuRankingReward0 struct {
	Index0           int32  //Place Start
	Index1           int32  //Place Finish
	Index2           uint32 // UNK
	DistributionType int32  //Type 7201:Item 7202:N Points 7203:Guild Contribution Points
	ItemID           int32
	Amount           int32
}
type WeeklySeibatuRankingReward1 struct {
	Unk0      int32
	ItemID    int32
	Amount    uint32
	PlaceFrom int32
	PlaceTo   int32
}

func handleMsgMhfGetWeeklySeibatuRankingReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetWeeklySeibatuRankingReward)
	var data []*byteframe.ByteFrame
	var weeklySeibatuRankingRewards []WeeklySeibatuRankingReward0
	var weeklySeibatuRankingRewardsWeird []WeeklySeibatuRankingReward1

	switch pkt.Operation {
	case 1:
		switch pkt.ID { // Seems to align with EarthStatus 1 and 2 for Conquest
		case 1:
			switch pkt.EarthMonster {
			case 116:
				weeklySeibatuRankingRewardsWeird = []WeeklySeibatuRankingReward1{
					{0, 2, 3, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 25, 1, 100},

					{0, 2, 2, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 30, 101, 1000},

					{0, 2, 2, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 6, 1000, 1001},
					{0, 2, 6, 1000, 1001},
				}
			case 107:
				weeklySeibatuRankingRewardsWeird = []WeeklySeibatuRankingReward1{
					{0, 2, 3, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 25, 1, 100},

					{0, 2, 2, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 30, 101, 1000},

					{0, 2, 2, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 6, 1000, 1001},
					{0, 2, 6, 1000, 1001},
				}
			case 2:
				weeklySeibatuRankingRewardsWeird = []WeeklySeibatuRankingReward1{
					{0, 2, 3, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 25, 1, 100},

					{0, 2, 2, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 30, 101, 1000},

					{0, 2, 2, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 6, 1000, 1001},
					{0, 2, 6, 1000, 1001},
				}
			case 36:
				weeklySeibatuRankingRewardsWeird = []WeeklySeibatuRankingReward1{
					{0, 2, 3, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 25, 1, 100},

					{0, 2, 2, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 30, 101, 1000},

					{0, 2, 2, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 6, 1000, 1001},
					{0, 2, 6, 1000, 1001},
				}
			}

		case 2:
			switch pkt.EarthMonster {
			case 116:
				weeklySeibatuRankingRewardsWeird = []WeeklySeibatuRankingReward1{
					{0, 2, 3, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 25, 1, 100},

					{0, 2, 2, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 30, 101, 1000},

					{0, 2, 2, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 6, 1000, 1001},
					{0, 2, 6, 1000, 1001},
				}
			case 107:
				weeklySeibatuRankingRewardsWeird = []WeeklySeibatuRankingReward1{
					{0, 2, 3, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 25, 1, 100},

					{0, 2, 2, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 30, 101, 1000},

					{0, 2, 2, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 6, 1000, 1001},
					{0, 2, 6, 1000, 1001},
				}
			case 2:
				weeklySeibatuRankingRewardsWeird = []WeeklySeibatuRankingReward1{
					{0, 2, 3, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 25, 1, 100},

					{0, 2, 2, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 30, 101, 1000},

					{0, 2, 2, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 6, 1000, 1001},
					{0, 2, 6, 1000, 1001},
				}
			case 36:
				weeklySeibatuRankingRewardsWeird = []WeeklySeibatuRankingReward1{
					{0, 2, 3, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 6, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 15, 1, 100},
					{0, 2, 25, 1, 100},

					{0, 2, 2, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 4, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 9, 101, 1000},
					{0, 2, 30, 101, 1000},

					{0, 2, 2, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 4, 1000, 1001},
					{0, 2, 6, 1000, 1001},
					{0, 2, 6, 1000, 1001},
				}
			}

		}
	case 3:
		weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward0{

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
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward0{
				{1, 1, 1, 7201, 12068, 1}}
		case 240041:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward0{
				{0, 0, 1, 7201, 12068, 1}}
		case 240042:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward0{
				{0, 0, 2, 7201, 12068, 1}}
		case 240051:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward0{
				{0, 0, 1, 7201, 12068, 1}}
		case 240052:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward0{
				{1, 1, 1, 7201, 12068, 1},
			}
		//Tower 260001 260003
		case 260001:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward0{

				//Can only have 10 in each dist (It disapears otherwise) Looks like up to dist 4 is implemented
				//This is claimable for every Dure Kill, Make cliamable in bulk or mandatory claim per kill
				//{unk,unk,dist,seiabtuType,ItemID,Value}
				{0, 0, 1, 7201, 11463, 1},
				{0, 0, 1, 7201, 11464, 1},
				{0, 0, 1, 7201, 11163, 1},
				{0, 0, 1, 7201, 11159, 5},
				{0, 0, 1, 7201, 11160, 5},
				{0, 0, 1, 7201, 11161, 5},

				{0, 0, 2, 7201, 12506, 1},
				{0, 0, 2, 7201, 10355, 1},
				{0, 0, 2, 7201, 11163, 1},
				{0, 0, 2, 7201, 11159, 5},
				{0, 0, 2, 7201, 11160, 5},
				{0, 0, 2, 7201, 11161, 5},
			}
		case 260003:
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward0{
				//Adjust Floors done in database to make blue
				//This is claimable for every Floor Climbed across dist 1 and 2
				//{Floor,unk,unk,seiabtuType,ItemID,Value}

				{1, 0, 0, 7201, 11158, 1},
				{2, 0, 0, 7201, 11173, 1},
				{3, 0, 0, 7201, 10813, 3},
				{4, 0, 0, 7201, 11163, 1},
				{5, 0, 0, 7201, 11164, 1},
				{6, 0, 0, 7201, 11389, 3},
				{6, 0, 0, 7201, 11381, 1},
				{7, 0, 0, 7201, 11384, 1},
				{8, 0, 0, 7201, 11159, 10},
				{9, 0, 0, 7201, 11160, 10},
				{10, 0, 0, 7201, 11161, 10},
				{11, 0, 0, 7201, 11265, 2},
				{11, 0, 0, 7201, 7279, 2},
				{12, 0, 0, 7201, 11381, 1},
				{13, 0, 0, 7201, 11384, 1},
				{14, 0, 0, 7201, 11381, 1},
				{15, 0, 0, 7201, 11384, 1},
				{15, 0, 0, 7201, 11464, 1},
				{16, 0, 0, 7201, 11381, 1},
				{17, 0, 0, 7201, 11384, 1},
				{18, 0, 0, 7201, 11381, 1},
				{19, 0, 0, 7201, 11384, 1},
				{20, 0, 0, 7201, 10778, 3},
				{21, 0, 0, 7201, 11265, 2},
				{21, 0, 0, 7201, 7279, 2},
				{22, 0, 0, 7201, 11381, 1},
				{23, 0, 0, 7201, 11384, 1},
				{24, 0, 0, 7201, 11381, 1},
				{25, 0, 0, 7201, 11389, 3},
				{25, 0, 0, 7201, 11286, 4},
				{26, 0, 0, 7201, 11384, 1},
				{27, 0, 0, 7201, 11381, 1},
				{28, 0, 0, 7201, 11384, 1},
				{29, 0, 0, 7201, 11381, 1},
				{30, 0, 0, 7201, 11209, 3},
				{31, 0, 0, 7201, 11265, 2},
				{31, 0, 0, 7201, 7279, 2},
				{32, 0, 0, 7201, 11159, 10},
				{33, 0, 0, 7201, 11463, 1},
				{34, 0, 0, 7201, 11160, 10},
				{35, 0, 0, 7201, 11286, 4},
				{36, 0, 0, 7201, 11161, 10},
				{38, 0, 0, 7201, 11384, 1},
				{39, 0, 0, 7201, 11164, 1},
				{40, 0, 0, 7201, 10813, 3},
				{41, 0, 0, 7201, 11265, 2},
				{41, 0, 0, 7201, 7280, 2},
				{43, 0, 0, 7201, 11381, 1},
				{45, 0, 0, 7201, 11286, 4},
				{47, 0, 0, 7201, 11384, 1},
				{48, 0, 0, 7201, 11358, 1},
				{50, 0, 0, 7201, 11356, 1},
				{51, 0, 0, 7201, 11265, 2},
				{51, 0, 0, 7201, 7280, 2},
				{53, 0, 0, 7201, 11381, 2},
				{55, 0, 0, 7201, 11357, 1},
				{57, 0, 0, 7201, 11384, 1},
				{60, 0, 0, 7201, 11286, 4},
				{61, 0, 0, 7201, 11265, 2},
				{61, 0, 0, 7201, 7280, 2},
				{63, 0, 0, 7201, 11381, 2},
				{66, 0, 0, 7201, 11463, 1},
				{67, 0, 0, 7201, 11384, 1},
				{70, 0, 0, 7201, 11286, 4},
				{71, 0, 0, 7201, 11265, 2},
				{71, 0, 0, 7201, 7280, 2},
				{73, 0, 0, 7201, 11381, 2},
				{77, 0, 0, 7201, 11384, 1},
				{79, 0, 0, 7201, 11164, 1},
				{80, 0, 0, 7201, 11286, 6},
				{81, 0, 0, 7201, 11265, 2},
				{81, 0, 0, 7201, 7281, 1},
				{83, 0, 0, 7201, 11381, 2},
				{85, 0, 0, 7201, 11464, 1},
				{87, 0, 0, 7201, 11384, 1},
				{90, 0, 0, 7201, 11286, 6},
				{91, 0, 0, 7201, 11265, 2},
				{91, 0, 0, 7201, 7281, 1},
				{93, 0, 0, 7201, 11381, 2},
				{95, 0, 0, 7201, 10778, 3},
				{97, 0, 0, 7201, 11384, 1},
				{99, 0, 0, 7201, 11463, 1},
				{100, 0, 0, 7201, 11286, 6},
				{101, 0, 0, 7201, 11265, 2},
				{101, 0, 0, 7201, 7281, 1},
				{103, 0, 0, 7201, 11381, 2},
				{107, 0, 0, 7201, 11384, 1},
				{110, 0, 0, 7201, 11286, 6},
				{113, 0, 0, 7201, 11381, 2},
				{115, 0, 0, 7201, 11164, 1},
				{117, 0, 0, 7201, 11384, 1},
				{120, 0, 0, 7201, 11286, 12},
				{123, 0, 0, 7201, 11381, 2},
				{127, 0, 0, 7201, 11384, 1},
				{130, 0, 0, 7201, 11286, 12},
				{132, 0, 0, 7201, 11381, 2},
				{134, 0, 0, 7201, 11384, 1},
				{136, 0, 0, 7201, 11381, 2},
				{138, 0, 0, 7201, 11384, 1},
				{140, 0, 0, 7201, 11286, 12},
				{142, 0, 0, 7201, 11382, 1},
				{144, 0, 0, 7201, 11385, 1},
				{145, 0, 0, 7201, 11464, 1},
				{146, 0, 0, 7201, 11382, 1},
				{148, 0, 0, 7201, 11385, 1},
				{150, 0, 0, 7201, 11164, 1},
				{155, 0, 0, 7201, 11382, 1},
				{160, 0, 0, 7201, 11209, 3},
				{165, 0, 0, 7201, 11385, 1},
				{170, 0, 0, 7201, 11159, 10},
				{175, 0, 0, 7201, 11382, 1},
				{180, 0, 0, 7201, 11160, 10},
				{185, 0, 0, 7201, 11385, 1},
				{190, 0, 0, 7201, 11161, 10},
				{195, 0, 0, 7201, 11382, 1},
				{200, 0, 0, 7201, 11159, 15},
				{210, 0, 0, 7201, 11160, 15},
				{220, 0, 0, 7201, 11385, 1},
				{235, 0, 0, 7201, 11382, 2},
				{250, 0, 0, 7201, 11161, 15},
				{265, 0, 0, 7201, 11159, 20},
				{280, 0, 0, 7201, 11385, 1},
				{300, 0, 0, 7201, 11160, 20},
				{315, 0, 0, 7201, 11382, 2},
				{330, 0, 0, 7201, 11385, 1},
				{350, 0, 0, 7201, 11161, 20},
				{365, 0, 0, 7201, 11382, 2},
				{380, 0, 0, 7201, 11385, 1},
				{400, 0, 0, 7201, 11159, 25},
				{415, 0, 0, 7201, 11382, 2},
				{430, 0, 0, 7201, 11385, 1},
				{450, 0, 0, 7201, 11160, 25},
				{465, 0, 0, 7201, 11382, 2},
				{480, 0, 0, 7201, 11385, 1},
				{500, 0, 0, 7201, 11161, 25},
				{525, 0, 0, 7201, 11382, 2},
				{550, 0, 0, 7201, 11385, 1},
				{575, 0, 0, 7201, 11159, 25},
				{600, 0, 0, 7201, 11382, 2},
				{625, 0, 0, 7201, 11385, 1},
				{650, 0, 0, 7201, 11160, 25},
				{675, 0, 0, 7201, 11382, 2},
				{700, 0, 0, 7201, 11385, 1},
				{725, 0, 0, 7201, 11161, 25},
				{750, 0, 0, 7201, 11382, 2},
				{775, 0, 0, 7201, 11385, 1},
				{800, 0, 0, 7201, 11159, 25},
				{825, 0, 0, 7201, 11382, 2},
				{850, 0, 0, 7201, 11385, 1},
				{875, 0, 0, 7201, 11160, 25},
				{900, 0, 0, 7201, 11382, 2},
				{925, 0, 0, 7201, 11385, 1},
				{950, 0, 0, 7201, 11161, 25},
				{975, 0, 0, 7201, 11382, 2},
				{1000, 0, 0, 7201, 11385, 1},
				{1025, 0, 0, 7201, 11159, 25},
				{1050, 0, 0, 7201, 11382, 2},
				{1075, 0, 0, 7201, 11385, 1},
				{1100, 0, 0, 7201, 11160, 25},
				{1125, 0, 0, 7201, 11382, 2},
				{1150, 0, 0, 7201, 11385, 1},
				{1200, 0, 0, 7201, 11161, 25},
				{1235, 0, 0, 7201, 11382, 2},
				{1270, 0, 0, 7201, 11385, 1},
				{1305, 0, 0, 7201, 11159, 25},
				{1340, 0, 0, 7201, 11382, 2},
				{1375, 0, 0, 7201, 11385, 1},
				{1410, 0, 0, 7201, 11160, 25},
				{1445, 0, 0, 7201, 11382, 2},
				{1480, 0, 0, 7201, 11385, 1},
				{1500, 0, 0, 7201, 11161, 25},
			}
		default: //Covers all Pallone Requests... for now
			weeklySeibatuRankingRewards = []WeeklySeibatuRankingReward0{
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
	if pkt.Operation == 1 {
		for _, seibatuData := range weeklySeibatuRankingRewardsWeird {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(seibatuData.Unk0)
			bf.WriteInt32(seibatuData.ItemID)
			bf.WriteUint32(seibatuData.Amount)
			bf.WriteInt32(seibatuData.PlaceFrom)
			bf.WriteInt32(seibatuData.PlaceTo)
			data = append(data, bf)
		}
	} else {
		for _, seibatuData := range weeklySeibatuRankingRewards {
			bf := byteframe.NewByteFrame()

			bf.WriteInt32(seibatuData.Index0)
			bf.WriteInt32(seibatuData.Index1)
			bf.WriteUint32(seibatuData.Index2)
			bf.WriteInt32(seibatuData.DistributionType)
			bf.WriteInt32(seibatuData.ItemID)
			bf.WriteInt32(seibatuData.Amount)
			data = append(data, bf)
		}
	}

	doAckEarthSucceed(s, pkt.AckHandle, data)
}

type FixedSeibatuRankingTable struct {
	Rank     int32
	Level    int32
	UnkArray string
}

func handleMsgMhfGetFixedSeibatuRankingTable(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetFixedSeibatuRankingTable)
	var fixedSeibatuRankingTable []FixedSeibatuRankingTable
	//Interestingly doesn't trigger the pkt on EarthStatus 1 But menu option is there is this Seibatu instead?
	switch pkt.EarthMonster {

	case 116:
		fixedSeibatuRankingTable = []FixedSeibatuRankingTable{
			{1, 1, "Hunter 1"},
			{2, 1, "Hunter 2"},
			{3, 1, "Hunter 3"},
			{4, 1, "Hunter 4"},
			{5, 1, "Hunter 5"},
			{6, 1, "Hunter 6"},
			{7, 1, "Hunter 7"},
			{8, 1, "Hunter 8"},
			{9, 1, "Hunter 9"},
		}
	case 107:
		fixedSeibatuRankingTable = []FixedSeibatuRankingTable{
			{1, 2, "Hunter 1"},
			{2, 2, "Hunter 2"},
			{3, 2, "Hunter 3"},
			{4, 2, "Hunter 4"},
			{5, 2, "Hunter 5"},
			{6, 2, "Hunter 6"},
			{7, 2, "Hunter 7"},
			{8, 2, "Hunter 8"},
			{9, 2, "Hunter 9"},
		}
	case 2:
		fixedSeibatuRankingTable = []FixedSeibatuRankingTable{
			{1, 3, "Hunter 1"},
			{2, 3, "Hunter 2"},
			{3, 3, "Hunter 3"},
			{4, 3, "Hunter 4"},
			{5, 3, "Hunter 5"},
			{6, 3, "Hunter 6"},
			{7, 3, "Hunter 7"},
			{8, 3, "Hunter 8"},
			{9, 3, "Hunter 9"},
		}
	case 36:
		fixedSeibatuRankingTable = []FixedSeibatuRankingTable{
			{1, 4, "Hunter 1"},
			{2, 4, "Hunter 2"},
			{3, 4, "Hunter 3"},
			{4, 4, "Hunter 4"},
			{5, 4, "Hunter 5"},
			{6, 4, "Hunter 6"},
			{7, 4, "Hunter 7"},
			{8, 4, "Hunter 8"},
			{9, 4, "Hunter 9"},
		}
	default:
		fixedSeibatuRankingTable = []FixedSeibatuRankingTable{
			{1, 1, "Hunter 1"},
			{2, 1, "Hunter 2"},
			{3, 1, "Hunter 3"},
			{4, 1, "Hunter 4"},
			{5, 1, "Hunter 5"},
			{6, 1, "Hunter 6"},
			{7, 1, "Hunter 7"},
			{8, 1, "Hunter 8"},
			{9, 1, "Hunter 9"},
		}
	}

	bf := byteframe.NewByteFrame()

	for _, seibatuData := range fixedSeibatuRankingTable {

		bf.WriteInt32(seibatuData.Rank)
		bf.WriteInt32(seibatuData.Level)
		bf.WriteBytes(stringsupport.PaddedString(seibatuData.UnkArray, 32, true))

	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfReadBeatLevel(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadBeatLevel)

	// This response is fixed and will never change on JP,
	// but I've left it dynamic for possible other client differences.
	resp := byteframe.NewByteFrame()
	for i := 0; i < int(pkt.ValidIDCount); i++ {
		resp.WriteUint32(pkt.IDs[i])
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)
	}

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfReadLastWeekBeatRanking(s *Session, p mhfpacket.MHFPacket) {
	//Controls the monster headings for the other menus
	pkt := p.(*mhfpacket.MsgMhfReadLastWeekBeatRanking)
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(uint32(pkt.EarthMonster))
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
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
