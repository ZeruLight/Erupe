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

func handleMsgMhfGetRyoudama(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRyoudama)

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(s.server.erupeConfig.DevModeOptions.EarthIDOverride))
	bf.WriteUint32(0)
	bf.WriteUint32(0)

	ryoudama := Ryoudama{Score: []int32{0}}
	switch pkt.Request2 {
	case 4:
		bf.WriteUint32(uint32(len(ryoudama.Score)))
		for _, score := range ryoudama.Score {
			bf.WriteInt32(score)
		}
	case 5:
		bf.WriteUint32(uint32(len(ryoudama.CharInfo)))
		for _, info := range ryoudama.CharInfo {
			bf.WriteUint32(info.CID)
			bf.WriteInt32(info.Unk0)
			bf.WriteBytes(stringsupport.PaddedString(info.Name, 14, true))
		}
	case 6:
		bf.WriteUint32(uint32(len(ryoudama.BoostInfo)))
		for _, info := range ryoudama.BoostInfo {
			bf.WriteUint32(uint32(info.Start.Unix()))
			bf.WriteUint32(uint32(info.End.Unix()))
		}
	default:
		bf.WriteUint32(0)
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfPostRyoudama(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTinyBin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTinyBin)
	// requested after conquest quests
	doAckBufSucceed(s, pkt.AckHandle, []byte{})
}

func handleMsgMhfPostTinyBin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTinyBin)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfCaravanMyScore(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanMyScore)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0) // Entries

	/*
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
	*/

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfCaravanRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanRanking)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0) // Entries

	/* RYOUDAN
	bf.WriteInt32(1)
	bf.WriteUint32(2)
	bf.WriteBytes(stringsupport.PaddedString("Test", 26, true))
	*/

	/* PERSONAL
	bf.WriteInt32(1)
	bf.WriteBytes(stringsupport.PaddedString("Test", 14, true))
	*/
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfCaravanMyRank(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanMyRank)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0)
	bf.WriteUint32(0) // Entries

	/*
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
	*/

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
