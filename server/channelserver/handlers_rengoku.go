package channelserver

import (
	ps "erupe-ce/common/pascalstring"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgMhfSaveRengokuData(s *Session, p mhfpacket.MHFPacket) {
	// saved every floor on road, holds values such as floors progressed, points etc.
	// can be safely handled by the client
	pkt := p.(*mhfpacket.MsgMhfSaveRengokuData)
	dumpSaveData(s, pkt.RawDataPayload, "rengoku")
	_, err := s.server.db.Exec("UPDATE characters SET rengokudata=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update rengokudata savedata in db", zap.Error(err))
	}
	bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
	bf.Seek(71, 0)
	maxStageMp := bf.ReadUint32()
	maxScoreMp := bf.ReadUint32()
	bf.Seek(4, 1)
	maxStageSp := bf.ReadUint32()
	maxScoreSp := bf.ReadUint32()
	var t int
	err = s.server.db.QueryRow("SELECT character_id FROM rengoku_score WHERE character_id=$1", s.charID).Scan(&t)
	if err != nil {
		s.server.db.Exec("INSERT INTO rengoku_score (character_id) VALUES ($1)", s.charID)
	}
	s.server.db.Exec("UPDATE rengoku_score SET max_stages_mp=$1, max_points_mp=$2, max_stages_sp=$3, max_points_sp=$4 WHERE character_id=$5", maxStageMp, maxScoreMp, maxStageSp, maxScoreSp, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadRengokuData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadRengokuData)
	var data []byte
	err := s.server.db.QueryRow("SELECT rengokudata FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get rengokudata savedata from db", zap.Error(err))
	}
	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		resp := byteframe.NewByteFrame()
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint16(0)
		resp.WriteUint32(0)
		resp.WriteUint16(0)
		resp.WriteUint16(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0) // an extra 4 bytes were missing based on pcaps

		resp.WriteUint8(3) // Count of next 3
		resp.WriteUint16(0)
		resp.WriteUint16(0)
		resp.WriteUint16(0)

		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)

		resp.WriteUint8(3) // Count of next 3
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)

		resp.WriteUint8(3) // Count of next 3
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)

		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)
		resp.WriteUint32(0)

		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
	}
}

func handleMsgMhfGetRengokuBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRengokuBinary)
	// a (massively out of date) version resides in the game's /dat/ folder or up to date can be pulled from packets
	data, err := ioutil.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, "rengoku_data.bin"))
	if err != nil {
		panic(err)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

const rengokuScoreQuery = `
SELECT max_stages_mp, max_points_mp, max_stages_sp, max_points_sp, c.name, gc.guild_id
FROM rengoku_score rs
LEFT JOIN characters c ON c.id = rs.character_id
LEFT JOIN guild_characters gc ON gc.character_id = rs.character_id
`

type RengokuScore struct {
	Name        string `db:"name"`
	GuildID     int    `db:"guild_id"`
	MaxStagesMP uint32 `db:"max_stages_mp"`
	MaxPointsMP uint32 `db:"max_points_mp"`
	MaxStagesSP uint32 `db:"max_stages_sp"`
	MaxPointsSP uint32 `db:"max_points_sp"`
}

func handleMsgMhfEnumerateRengokuRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRengokuRanking)

	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	isApplicant, _ := guild.HasApplicationForCharID(s, s.charID)
	if isApplicant {
		guild = nil
	}

	var score RengokuScore
	i := uint32(1)
	bf := byteframe.NewByteFrame()
	scoreData := byteframe.NewByteFrame()
	switch pkt.Leaderboard {
	case 0: // Max stage overall MP
		rows, _ := s.server.db.Queryx(fmt.Sprintf("%s ORDER BY max_stages_mp DESC", rengokuScoreQuery))
		for rows.Next() {
			rows.StructScan(&score)
			if score.Name == s.Name {
				bf.WriteUint32(i)
				bf.WriteUint32(score.MaxStagesMP)
				ps.Uint8(bf, s.Name, true)
				ps.Uint8(bf, "", false)
			}
			scoreData.WriteUint32(i)
			scoreData.WriteUint32(score.MaxStagesMP)
			ps.Uint8(scoreData, score.Name, true)
			ps.Uint8(scoreData, "", false)
			i++
		}
	case 1: // Max RdP overall MP
		rows, _ := s.server.db.Queryx(fmt.Sprintf("%s ORDER BY max_points_mp DESC", rengokuScoreQuery))
		for rows.Next() {
			rows.StructScan(&score)
			if score.Name == s.Name {
				bf.WriteUint32(i)
				bf.WriteUint32(score.MaxPointsMP)
				ps.Uint8(bf, s.Name, true)
				ps.Uint8(bf, "", false)
			}
			scoreData.WriteUint32(i)
			scoreData.WriteUint32(score.MaxPointsMP)
			ps.Uint8(scoreData, score.Name, true)
			ps.Uint8(scoreData, "", false)
			i++
		}
	case 2: // Max stage guild MP
		if guild != nil {
			rows, _ := s.server.db.Queryx(fmt.Sprintf("%s WHERE guild_id=$1 ORDER BY max_stages_mp DESC", rengokuScoreQuery), guild.ID)
			for rows.Next() {
				rows.StructScan(&score)
				if score.Name == s.Name {
					bf.WriteUint32(i)
					bf.WriteUint32(score.MaxStagesMP)
					ps.Uint8(bf, s.Name, true)
					ps.Uint8(bf, "", false)
				}
				scoreData.WriteUint32(i)
				scoreData.WriteUint32(score.MaxStagesMP)
				ps.Uint8(scoreData, score.Name, true)
				ps.Uint8(scoreData, "", false)
				i++
			}
		} else {
			bf.WriteBytes(make([]byte, 11))
		}
	case 3: // Max RdP guild MP
		if guild != nil {
			rows, _ := s.server.db.Queryx(fmt.Sprintf("%s WHERE guild_id=$1 ORDER BY max_points_mp DESC", rengokuScoreQuery), guild.ID)
			for rows.Next() {
				rows.StructScan(&score)
				if score.Name == s.Name {
					bf.WriteUint32(i)
					bf.WriteUint32(score.MaxPointsMP)
					ps.Uint8(bf, s.Name, true)
					ps.Uint8(bf, "", false)
				}
				scoreData.WriteUint32(i)
				scoreData.WriteUint32(score.MaxPointsMP)
				ps.Uint8(scoreData, score.Name, true)
				ps.Uint8(scoreData, "", false)
				i++
			}
		} else {
			bf.WriteBytes(make([]byte, 11))
		}
	case 4: // Max stage overall SP
		rows, _ := s.server.db.Queryx(fmt.Sprintf("%s ORDER BY max_stages_sp DESC", rengokuScoreQuery))
		for rows.Next() {
			rows.StructScan(&score)
			if score.Name == s.Name {
				bf.WriteUint32(i)
				bf.WriteUint32(score.MaxStagesSP)
				ps.Uint8(bf, s.Name, true)
				ps.Uint8(bf, "", false)
			}
			scoreData.WriteUint32(i)
			scoreData.WriteUint32(score.MaxStagesSP)
			ps.Uint8(scoreData, score.Name, true)
			ps.Uint8(scoreData, "", false)
			i++
		}
	case 5: // Max RdP overall SP
		rows, _ := s.server.db.Queryx(fmt.Sprintf("%s ORDER BY max_points_sp DESC", rengokuScoreQuery))
		for rows.Next() {
			rows.StructScan(&score)
			if score.Name == s.Name {
				bf.WriteUint32(i)
				bf.WriteUint32(score.MaxPointsSP)
				ps.Uint8(bf, s.Name, true)
				ps.Uint8(bf, "", false)
			}
			scoreData.WriteUint32(i)
			scoreData.WriteUint32(score.MaxPointsSP)
			ps.Uint8(scoreData, score.Name, true)
			ps.Uint8(scoreData, "", false)
			i++
		}
	case 6: // Max stage guild SP
		if guild != nil {
			rows, _ := s.server.db.Queryx(fmt.Sprintf("%s WHERE guild_id=$1 ORDER BY max_stages_sp DESC", rengokuScoreQuery), guild.ID)
			for rows.Next() {
				rows.StructScan(&score)
				if score.Name == s.Name {
					bf.WriteUint32(i)
					bf.WriteUint32(score.MaxStagesSP)
					ps.Uint8(bf, s.Name, true)
					ps.Uint8(bf, "", false)
				}
				scoreData.WriteUint32(i)
				scoreData.WriteUint32(score.MaxStagesSP)
				ps.Uint8(scoreData, score.Name, true)
				ps.Uint8(scoreData, "", false)
				i++
			}
		} else {
			bf.WriteBytes(make([]byte, 11))
		}
	case 7: // Max RdP guild SP
		if guild != nil {
			rows, _ := s.server.db.Queryx(fmt.Sprintf("%s WHERE guild_id=$1 ORDER BY max_points_sp DESC", rengokuScoreQuery), guild.ID)
			for rows.Next() {
				rows.StructScan(&score)
				if score.Name == s.Name {
					bf.WriteUint32(i)
					bf.WriteUint32(score.MaxPointsSP)
					ps.Uint8(bf, s.Name, true)
					ps.Uint8(bf, "", false)
				}
				scoreData.WriteUint32(i)
				scoreData.WriteUint32(score.MaxPointsSP)
				ps.Uint8(scoreData, score.Name, true)
				ps.Uint8(scoreData, "", false)
				i++
			}
		} else {
			bf.WriteBytes(make([]byte, 11))
		}
	}
	bf.WriteUint8(uint8(i) - 1)
	bf.WriteBytes(scoreData.Data())
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetRengokuRankingRank(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetRengokuRankingRank)
	// What is this for?
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0) // Max stage overall MP rank
	bf.WriteUint32(0) // Max RdP overall MP rank
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
