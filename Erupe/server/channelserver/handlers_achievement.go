package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

var achievementCurves = [][]uint32{
	// 0: HR weapon use, Class use, Tore dailies
	{5, 15, 30, 50, 100, 150, 200, 250, 300},
	// 1: Weapon collector, G wep enhances
	{1, 3, 5, 15, 30, 50, 75, 100, 150},
	// 2: Festa wins
	{1, 2, 3, 4, 5, 6, 7, 8, 9},
	// 3: GR weapon use
	{10, 50, 100, 200, 350, 500, 750, 1000, 1500},
	// 4: Armor refinement
	{0, 5, 5, 5, 5, 5, 5, 5, 5},
	// 5: Sigil crafts
	{0, 50, 50, 50, 50, 50, 50, 50, 50},
}

var achievementCurveMap = map[uint8][]uint32{
	0: achievementCurves[0], 1: achievementCurves[0], 2: achievementCurves[0], 3: achievementCurves[0],
	4: achievementCurves[0], 5: achievementCurves[0], 6: achievementCurves[0], 7: achievementCurves[1],
	8: achievementCurves[2], 9: achievementCurves[0], 10: achievementCurves[0], 11: achievementCurves[0],
	12: achievementCurves[0], 13: achievementCurves[0], 14: achievementCurves[0], 15: achievementCurves[0],
	16: achievementCurves[3], 17: achievementCurves[3], 18: achievementCurves[3], 19: achievementCurves[3],
	20: achievementCurves[3], 21: achievementCurves[3], 22: achievementCurves[3], 23: achievementCurves[3],
	24: achievementCurves[3], 25: achievementCurves[3], 26: achievementCurves[3], 27: achievementCurves[1],
	28: achievementCurves[4], 29: achievementCurves[5], 30: achievementCurves[3], 31: achievementCurves[3],
	32: achievementCurves[3],
}

func handleMsgMhfGetAchievement(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetAchievement)

	err := s.server.db.QueryRow("SELECT id FROM achievements WHERE id=$1", s.charID)
	if err != nil {
		s.server.db.Exec("INSERT INTO achievements (id) VALUES ($1)", s.charID)
	}

	scores := make([]int, 33)
	s.server.db.QueryRow("SELECT * FROM achievements WHERE id=$1", s.charID).Scan(&scores[0], &scores[0],
		&scores[1], &scores[2], &scores[3], &scores[4], &scores[5], &scores[6], &scores[7], &scores[8], &scores[9],
		&scores[10], &scores[11], &scores[12], &scores[13], &scores[14], &scores[15], &scores[16], &scores[17],
		&scores[18], &scores[19], &scores[20], &scores[21], &scores[22], &scores[23], &scores[24], &scores[25],
		&scores[26], &scores[27], &scores[28], &scores[29], &scores[30], &scores[31], &scores[32])

	resp := byteframe.NewByteFrame()
	points := uint32(69)
	resp.WriteUint32(points)
	resp.WriteUint32(points)
	resp.WriteUint32(points)
	resp.WriteUint32(points)
	resp.WriteBytes([]byte{0x02, 0x00, 0x00})

	entries := 34
	resp.WriteUint8(uint8(entries)) // Entry count
	for i := 0; i < entries; i++ {
		resp.WriteUint8(uint8(i)) // achievement id
		resp.WriteUint8(uint8(i)) // level
		resp.WriteUint16(20)      // point value
		resp.WriteUint32(100)     // required
		resp.WriteUint8(0)
		if i < 10 {
			resp.WriteUint16(0x7FFF)
		} else if i < 20 {
			resp.WriteUint16(0x3FFF)
		} else {
			resp.WriteUint16(0x1FFF)
		}
		//resp.WriteUint16(0x7F7F) // unk
		resp.WriteUint8(0)
		resp.WriteUint32(100) // progress
	}
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfSetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetCaAchievementHist)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfResetAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfPaymentAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfDisplayedAchievement(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetCaAchievementHist(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetCaAchievement(s *Session, p mhfpacket.MHFPacket) {}
