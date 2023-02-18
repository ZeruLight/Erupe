package channelserver

import (
	"encoding/hex"
	"erupe-ce/network/mhfpacket"
)

func handleMsgMhfGetRyoudama(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRyoudama)
	// likely guild related
	// REQ: 00 04 13 53 8F 18 00
	// RSP: 0A 21 8E AD 00 00 00 00 00 00 00 00 00 00 00 01 00 01 FE 4E
	// REQ: 00 06 13 53 8F 18 00
	// RSP: 0A 21 8E AD 00 00 00 00 00 00 00 00 00 00 00 01 00 00 00 00 00 00 00 00
	// REQ: 00 05 13 53 8F 18 00
	// RSP: 0A 21 8E AD 00 00 00 00 00 00 00 00 00 00 00 0E 2A 15 9E CC 00 00 00 01 82 79 83 4E 83 8A 81 5B 83 69 00 00 00 00 1E 55 B0 2F 00 00 00 01 8D F7 00 00 00 00 00 00 00 00 00 00 00 00 2A 15 9E CC 00 00 00 02 82 79 83 4E 83 8A 81 5B 83 69 00 00 00 00 03 D5 30 56 00 00 00 02 95 BD 91 F2 97 42 00 00 00 00 00 00 00 00 3F 57 76 9F 00 00 00 03 93 56 92 6E 96 B3 97 70 00 00 00 00 00 00 38 D9 0E C4 00 00 00 03 87 64 83 78 83 42 00 00 00 00 00 00 00 00 23 F3 B9 77 00 00 00 04 82 B3 82 CC 82 DC 82 E9 81 99 00 00 00 00 3F 1B 17 9C 00 00 00 04 82 B1 82 A4 82 BD 00 00 00 00 00 00 00 00 00 B9 F9 C0 00 00 00 05 82 CD 82 E9 82 A9 00 00 00 00 00 00 00 00 23 9F 9A EA 00 00 00 05 83 70 83 62 83 4C 83 83 83 49 00 00 00 00 38 D9 0E C4 00 00 00 06 87 64 83 78 83 42 00 00 00 00 00 00 00 00 1E 55 B0 2F 00 00 00 06 8D F7 00 00 00 00 00 00 00 00 00 00 00 00 03 D5 30 56 00 00 00 07 95 BD 91 F2 97 42 00 00 00 00 00 00 00 00 02 D3 B8 77 00 00 00 07 6F 77 6C 32 35 32 35 00 00 00 00 00 00 00
	data, _ := hex.DecodeString("0A218EAD0000000000000000000000010000000000000000")
	doAckBufSucceed(s, pkt.AckHandle, data)
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

func handleMsgMhfCaravanMyScore(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCaravanRanking(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfCaravanMyRank(s *Session, p mhfpacket.MHFPacket) {}
