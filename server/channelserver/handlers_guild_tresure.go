package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
)

type TreasureHunt struct {
	HuntID      uint32 `db:"id"`
	HostID      uint32 `db:"host_id"`
	Destination uint32 `db:"destination"`
	Level       uint32 `db:"level"`
	Return      uint32 `db:"return"`
	Acquired    bool   `db:"acquired"`
	Claimed     bool   `db:"claimed"`
	Hunters     string `db:"hunters"`
	Treasure    string `db:"treasure"`
	HuntData    []byte `db:"hunt_data"`
}

func handleMsgMhfEnumerateGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildTresure)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil || guild == nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	var hunts []TreasureHunt
	var hunt TreasureHunt
	rows, err := s.server.db.Queryx(`SELECT id, host_id, destination, level, return, acquired, claimed, hunters, treasure, hunt_data FROM guild_hunts WHERE guild_id=$1 AND $2 < return+$3
		`, guild.ID, TimeAdjusted().Unix(), s.server.erupeConfig.GameplayOptions.TreasureHuntExpiry)
	if err != nil {
		rows.Close()
		return
	}
	for rows.Next() {
		err = rows.StructScan(&hunt)
		if err != nil {
			continue
		}
		// Remove self from other hunter count
		hunt.Hunters = stringsupport.CSVRemove(hunt.Hunters, int(s.charID))
		if pkt.MaxHunts == 1 {
			if hunt.HostID != s.charID || hunt.Acquired {
				continue
			}
			hunt.Claimed = false
			hunt.Treasure = ""
			hunts = append(hunts, hunt)
			break
		} else if pkt.MaxHunts == 30 && hunt.Acquired && hunt.Level == 2 {
			hunts = append(hunts, hunt)
		}
	}
	if len(hunts) > 30 {
		hunts = hunts[:30]
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(hunts)))
	bf.WriteUint16(uint16(len(hunts)))
	for _, h := range hunts {
		bf.WriteUint32(h.HuntID)
		bf.WriteUint32(h.Destination)
		bf.WriteUint32(h.Level)
		bf.WriteUint32(uint32(stringsupport.CSVLength(h.Hunters)))
		bf.WriteUint32(h.Return)
		bf.WriteBool(h.Claimed)
		bf.WriteBool(stringsupport.CSVContains(h.Treasure, int(s.charID)))
		bf.WriteBytes(h.HuntData)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfRegistGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildTresure)
	bf := byteframe.NewByteFrameFromBytes(pkt.Data)
	huntData := byteframe.NewByteFrame()
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	guildCats := getGuildAirouList(s)
	destination := bf.ReadUint32()
	level := bf.ReadUint32()
	huntData.WriteUint32(s.charID)
	huntData.WriteBytes(stringsupport.PaddedString(s.Name, 18, true))
	catsUsed := ""
	for i := 0; i < 5; i++ {
		catID := bf.ReadUint32()
		huntData.WriteUint32(catID)
		if catID > 0 {
			catsUsed = stringsupport.CSVAdd(catsUsed, int(catID))
			for _, cat := range guildCats {
				if cat.ID == catID {
					huntData.WriteBytes(cat.Name)
					break
				}
			}
			huntData.WriteBytes(bf.ReadBytes(9))
		}
	}
	s.server.db.Exec(`INSERT INTO guild_hunts (guild_id, host_id, destination, level, return, hunt_data, cats_used) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, guild.ID, s.charID, destination, level, TimeAdjusted().Unix(), huntData.Data(), catsUsed)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildTresure)
	s.server.db.Exec("UPDATE guild_hunts SET acquired=true WHERE id=$1", pkt.HuntID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func treasureHuntUnregister(s *Session) {
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil || guild == nil {
		return
	}
	var huntID int
	var hunters string
	rows, err := s.server.db.Queryx("SELECT id, hunters FROM guild_hunts WHERE guild_id=$1", guild.ID)
	if err != nil {
		rows.Close()
		return
	}
	for rows.Next() {
		rows.Scan(&huntID, &hunters)
		hunters = stringsupport.CSVRemove(hunters, int(s.charID))
		s.server.db.Exec("UPDATE guild_hunts SET hunters=$1 WHERE id=$2", hunters, huntID)
	}
}

func handleMsgMhfOperateGuildTresureReport(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuildTresureReport)
	var csv string
	switch pkt.State {
	case 0: // Report registration
		// Unregister from all other hunts
		treasureHuntUnregister(s)
		if pkt.HuntID != 0 {
			// Register to selected hunt
			err := s.server.db.QueryRow("SELECT hunters FROM guild_hunts WHERE id=$1", pkt.HuntID).Scan(&csv)
			if err != nil {
				doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
				return
			}
			csv = stringsupport.CSVAdd(csv, int(s.charID))
			s.server.db.Exec("UPDATE guild_hunts SET hunters=$1 WHERE id=$2", csv, pkt.HuntID)
		}
	case 1: // Collected by hunter
		s.server.db.Exec("UPDATE guild_hunts SET hunters='', claimed=true WHERE id=$1", pkt.HuntID)
	case 2: // Claim treasure
		err := s.server.db.QueryRow("SELECT treasure FROM guild_hunts WHERE id=$1", pkt.HuntID).Scan(&csv)
		if err != nil {
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		csv = stringsupport.CSVAdd(csv, int(s.charID))
		s.server.db.Exec("UPDATE guild_hunts SET treasure=$1 WHERE id=$2", csv, pkt.HuntID)
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildTresureSouvenir)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	bf.WriteUint16(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildTresureSouvenir)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
