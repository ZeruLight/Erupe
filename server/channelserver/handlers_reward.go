package channelserver

import (
	"encoding/hex"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
)

func handleMsgMhfGetAdditionalBeatReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetAdditionalBeatReward)
	// Actual response in packet captures are all just giant batches of null bytes
	// I'm assuming this is because it used to be tied to an actual event and
	// they never bothered killing off the packet when they made it static
	s.DoAckBufSucceed(pkt.AckHandle, make([]byte, 0x104))
}

func handleMsgMhfGetUdRankingRewardList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdRankingRewardList)
	// Temporary canned response
	data, _ := hex.DecodeString("0100001600000A5397DF00000000000000000000000000000000")
	s.DoAckBufSucceed(pkt.AckHandle, data)
}

func handleMsgMhfGetRewardSong(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRewardSong)
	// Temporary canned response
	data, _ := hex.DecodeString("0100001600000A5397DF00000000000000000000000000000000")
	s.DoAckBufSucceed(pkt.AckHandle, data)
}

func handleMsgMhfUseRewardSong(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddRewardSongCount(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAcquireMonthlyReward(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireMonthlyReward)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)

	s.DoAckBufSucceed(pkt.AckHandle, resp.Data())
}

func handleMsgMhfAcceptReadReward(s *Session, p mhfpacket.MHFPacket) {}
