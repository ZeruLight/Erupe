package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"time"
)

type RyoudamaReward struct {
	Unk0 uint8
	Unk1 uint8
	Unk2 uint16
	Unk3 uint16
	Unk4 uint16
	Unk5 uint16
}

type RyoudamaKeyScore struct {
	Unk0 uint8
	Unk1 int32
}

type RyoudamaCharInfo struct {
	CID  uint32
	Unk0 int32
	Name string
}

type RyoudamaBoostInfo struct {
	Start time.Time
	End   time.Time
}

type Ryoudama struct {
	Reward    []RyoudamaReward
	KeyScore  []RyoudamaKeyScore
	CharInfo  []RyoudamaCharInfo
	BoostInfo []RyoudamaBoostInfo
	Score     []int32
}

type CaravanMyScore struct {
	Unk0    int32
	MyScore int32
	Unk2    int32
	Unk3    int32
}

type CaravanMyRank struct {
	Unk0 int32
	Unk1 int32
	Unk2 int32
}

type CaravanRyoudanRanking struct {
	Score          int32
	HuntingGroupId uint32
	Name           string
}
type CaravanPersonalRanking struct {
	Score int32
	Name  string
}

func handleMsgMhfGetRyoudama(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRyoudama)
	var data []*byteframe.ByteFrame
	ryoudama := Ryoudama{Score: []int32{0}}
	switch pkt.Request2 {
	case 4:
		for _, score := range ryoudama.Score {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(score)
			data = append(data, bf)
		}
	case 5:
		for _, info := range ryoudama.CharInfo {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(info.CID)
			bf.WriteInt32(info.Unk0)
			bf.WriteBytes(stringsupport.PaddedString(info.Name, 14, true))
			data = append(data, bf)
		}
	case 6:
		for _, info := range ryoudama.BoostInfo {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(uint32(info.Start.Unix()))
			bf.WriteUint32(uint32(info.End.Unix()))
			data = append(data, bf)
		}
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfPostRyoudama(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTinyBin(s *Session, p mhfpacket.MHFPacket) {
	//Works for Tower but not Conquest

	//Conquest: Unk0 0 Unk1 2 Unk2 1
	type TinyBinItem struct {
		ItemId uint16
		Amount uint8
		Unk2   uint8 //if 4 the Red message "There are some items and points that cannot be recieved." Shows
	}

	tinyBinItems := []TinyBinItem{{7, 2, 4}, {8, 1, 0}, {9, 1, 0}, {300, 4, 0}, {10, 1, 0}}

	pkt := p.(*mhfpacket.MsgMhfGetTinyBin)
	// requested after conquest quests
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	for _, items := range tinyBinItems {
		bf.WriteUint16(items.ItemId)
		bf.WriteUint8(items.Amount)
		bf.WriteUint8(items.Unk2)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfPostTinyBin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTinyBin)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfCaravanMyScore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanMyScore)
	ryoudama := []CaravanMyScore{{6, 60900, 6, 6}}

	var data []*byteframe.ByteFrame
	for _, score := range ryoudama {
		bf := byteframe.NewByteFrame()
		bf.WriteInt32(score.Unk0)
		bf.WriteInt32(score.MyScore)
		bf.WriteInt32(score.Unk2)
		bf.WriteInt32(score.Unk3)
		data = append(data, bf)
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfCaravanRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanRanking)
	var data []*byteframe.ByteFrame

	// 1 = Top 100 when this Unk2 is the
	// 4 = Guild Score
	// 5 = Guild Team Individual Score
	// 2 = Personal Score
	switch pkt.Operation {
	case 1:
		personalRanking := []CaravanPersonalRanking{{60900, "Hunter 0"}, {20, "Hunter a"}, {4, "Hunter b"}, {4, "Hunter c"}, {2, "Hunter d"}, {1, "Hunter e"}}
		for _, score := range personalRanking {
			bf := byteframe.NewByteFrame()

			bf.WriteInt32(score.Score)
			bf.WriteBytes(stringsupport.PaddedString(score.Name, 14, true))
			data = append(data, bf)
		}
	case 2:
		personalRanking := []CaravanPersonalRanking{{60900, "Hunter 0"}, {20, "Hunter a"}, {4, "Hunter b"}, {4, "Hunter c"}, {2, "Hunter d"}, {1, "Hunter e"}}
		for _, score := range personalRanking {
			bf := byteframe.NewByteFrame()

			bf.WriteInt32(score.Score)
			bf.WriteBytes(stringsupport.PaddedString(score.Name, 14, true))
			data = append(data, bf)
		}
	case 3:
		ryoudama := []CaravanRyoudanRanking{{5, 1, "Clan a"}, {4, 2, "Clan b"}, {3, 3, "Clan c"}, {2, 4, "Clan d"}, {1, 5, "Clan e"}, {0, 6, "Clan f"}}
		for _, score := range ryoudama {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(score.Score)
			bf.WriteUint32(score.HuntingGroupId)
			bf.WriteBytes(stringsupport.PaddedString(score.Name, 26, true))
			data = append(data, bf)
		}
	case 4:
		ryoudama := []CaravanRyoudanRanking{{5, 1, "Clan a"}, {4, 2, "Clan b"}, {3, 3, "Clan c"}, {2, 4, "Clan d"}, {1, 5, "Clan e"}, {0, 6, "Clan f"}}
		for _, score := range ryoudama {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(score.Score)
			bf.WriteUint32(score.HuntingGroupId)
			bf.WriteBytes(stringsupport.PaddedString(score.Name, 26, true))
			data = append(data, bf)
		}
	case 5:
		//Unk2 is Hunting Team ID
		//Having more than 5 in array stops loading
		// Do Select ... Where HunterTeamID = pkt.unk2
		personalRanking := []CaravanPersonalRanking{{10, "Clan Hunter 1"}, {9, "Clan Hunter 2"}, {8, "Clan Hunter 3"}, {7, "Clan Hunter 4"}, {6, "Clan Hunter 5"}}

		for _, score := range personalRanking {
			bf := byteframe.NewByteFrame()
			bf.WriteInt32(score.Score)
			bf.WriteBytes(stringsupport.PaddedString(score.Name, 14, true))
			data = append(data, bf)
		}
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfCaravanMyRank(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanMyRank)
	// if value 1 is 0 ! on the route
	unk := []CaravanMyRank{{6, 6, 6}}
	//Personal - General Unk 1 : 1
	var data []*byteframe.ByteFrame
	for _, unkData := range unk {
		bf := byteframe.NewByteFrame()
		bf.WriteInt32(unkData.Unk0)
		bf.WriteInt32(unkData.Unk1)
		bf.WriteInt32(unkData.Unk2)
		data = append(data, bf)
	}
	doAckEarthSucceed(s, pkt.AckHandle, data)
}
