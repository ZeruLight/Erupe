package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func handleMsgMhfGetAdditionalBeatReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetAdditionalBeatReward)
	// Actual response in packet captures are all just giant batches of null bytes
	// I'm assuming this is because it used to be tied to an actual event and
	// they never bothered killing off the packet when they made it static
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 0x104))
}

func handleMsgMhfGetUdRankingRewardList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdRankingRewardList)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0) // Len
	// Format
	// uint8 Unk
	// uint16 Unk
	// uint16 Unk
	// uint8 Unk
	// uint32 Unk
	// uint32 Unk
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetRewardSong(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRewardSong)
	var activationTime = 0xFFFFFFFF
	var questCount = 0
	var activationCount = 0
	s.server.db.QueryRow(`SELECT (EXTRACT(epoch FROM activation_time)::int) as activation_time, quest_count, activation_count FROM diva_buffs WHERE character_id=$1`, s.charID).Scan(&activationTime, &questCount, &activationCount)

	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)            			 		// No error
	bf.WriteUint8(uint8(activationCount))           		// Activation Count
	bf.WriteUint32(0)           			 		// UKN
	bf.WriteUint32(uint32(activationTime))	// Prayer activation time
	for i := 0; i < 4; i++ {    					// Quest Count
		bf.WriteUint16(uint16(questCount))
		bf.WriteUint8(uint8(questCount))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUseRewardSong(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUseRewardSong)
	activationTime := TimeAdjusted()
	s.server.db.Exec(`INSERT INTO diva_buffs VALUES ($1, $2, 0, 1) ON CONFLICT (character_id) DO UPDATE SET activation_time = excluded.activation_time, quest_count = excluded.quest_count, activation_count = diva_buffs.activation_count + 1;`, s.charID, activationTime)
	doAckBufSucceed(s, pkt.AckHandle, []byte{0})
}

func handleMsgMhfAddRewardSongCount(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddRewardSongCount)
	s.server.db.Exec(`UPDATE diva_buffs SET quest_count=quest_count+1 WHERE character_id = $1`, s.charID);
	doAckBufSucceed(s, pkt.AckHandle, []byte{0})
}

func handleMsgMhfAcquireMonthlyReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireMonthlyReward)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfAcceptReadReward(s *Session, p mhfpacket.MHFPacket) {}
