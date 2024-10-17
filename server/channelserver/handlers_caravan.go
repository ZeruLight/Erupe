package channelserver

import (
	"erupe-ce/internal/model"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/stringsupport"

	"github.com/jmoiron/sqlx"
)

func handleMsgMhfGetRyoudama(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRyoudama)
	var data []*byteframe.ByteFrame
	ryoudama := model.Ryoudama{Score: []int32{0}}
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
	s.DoAckEarthSucceed(pkt.AckHandle, data)
}

func handleMsgMhfPostRyoudama(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetTinyBin(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetTinyBin)
	// requested after conquest quests
	s.DoAckBufSucceed(pkt.AckHandle, []byte{})
}

func handleMsgMhfPostTinyBin(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfPostTinyBin)
	s.DoAckSimpleSucceed(pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfCaravanMyScore(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanMyScore)
	var data []*byteframe.ByteFrame
	/*
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
	*/
	s.DoAckEarthSucceed(pkt.AckHandle, data)
}

func handleMsgMhfCaravanRanking(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanRanking)
	var data []*byteframe.ByteFrame
	/* RYOUDAN
	bf.WriteInt32(1)
	bf.WriteUint32(2)
	bf.WriteBytes(stringsupport.PaddedString("Test", 26, true))
	*/

	/* PERSONAL
	bf.WriteInt32(1)
	bf.WriteBytes(stringsupport.PaddedString("Test", 14, true))
	*/
	s.DoAckEarthSucceed(pkt.AckHandle, data)
}

func handleMsgMhfCaravanMyRank(s *Session, db *sqlx.DB, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCaravanMyRank)
	var data []*byteframe.ByteFrame
	/*
		bf.WriteInt32(0)
		bf.WriteInt32(0)
		bf.WriteInt32(0)
	*/
	s.DoAckEarthSucceed(pkt.AckHandle, data)
}
