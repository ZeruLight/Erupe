package channelserver

import (
	"erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/broadcast"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/db"
	"erupe-ce/utils/gametime"
	"erupe-ce/utils/stringsupport"
	"fmt"
	"time"
)

type TreasureHunt struct {
	HuntID      uint32    `db:"id"`
	HostID      uint32    `db:"host_id"`
	Destination uint32    `db:"destination"`
	Level       uint32    `db:"level"`
	Start       time.Time `db:"start"`
	Acquired    bool      `db:"acquired"`
	Collected   bool      `db:"collected"`
	HuntData    []byte    `db:"hunt_data"`
	Hunters     uint32    `db:"hunters"`
	Claimed     bool      `db:"claimed"`
}

func HandleMsgMhfEnumerateGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildTresure)
	guild, err := GetGuildInfoByCharacterId(s, s.CharID)
	if err != nil || guild == nil {
		broadcast.DoAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	var hunts []TreasureHunt
	var hunt TreasureHunt
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	switch pkt.MaxHunts {
	case 1:
		err = database.QueryRowx(`SELECT id, host_id, destination, level, start, hunt_data FROM guild_hunts WHERE host_id=$1 AND acquired=FALSE`, s.CharID).StructScan(&hunt)
		if err == nil {
			hunts = append(hunts, hunt)
		}
	case 30:
		rows, err := database.Queryx(`SELECT gh.id, gh.host_id, gh.destination, gh.level, gh.start, gh.collected, gh.hunt_data,
			(SELECT COUNT(*) FROM guild_characters gc WHERE gc.treasure_hunt = gh.id AND gc.character_id <> $1) AS hunters,
			CASE
				WHEN ghc.character_id IS NOT NULL THEN true
				ELSE false
			END AS claimed
			FROM guild_hunts gh
			LEFT JOIN guild_hunts_claimed ghc ON gh.id = ghc.hunt_id AND ghc.character_id = $1
			WHERE gh.guild_id=$2 AND gh.level=2 AND gh.acquired=TRUE
		`, s.CharID, guild.ID)
		if err != nil {
			rows.Close()
			broadcast.DoAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		} else {
			for rows.Next() {
				err = rows.StructScan(&hunt)
				if err == nil && hunt.Start.Add(time.Second*time.Duration(config.GetConfig().GameplayOptions.TreasureHuntExpiry)).After(gametime.TimeAdjusted()) {
					hunts = append(hunts, hunt)
				}
			}
		}
		if len(hunts) > 30 {
			hunts = hunts[:30]
		}
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(hunts)))
	bf.WriteUint16(uint16(len(hunts)))
	for _, h := range hunts {
		bf.WriteUint32(h.HuntID)
		bf.WriteUint32(h.Destination)
		bf.WriteUint32(h.Level)
		bf.WriteUint32(h.Hunters)
		bf.WriteUint32(uint32(h.Start.Unix()))
		bf.WriteBool(h.Collected)
		bf.WriteBool(h.Claimed)
		bf.WriteBytes(h.HuntData)
	}
	broadcast.DoAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func HandleMsgMhfRegistGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildTresure)
	bf := byteframe.NewByteFrameFromBytes(pkt.Data)
	huntData := byteframe.NewByteFrame()
	guild, err := GetGuildInfoByCharacterId(s, s.CharID)
	if err != nil || guild == nil {
		broadcast.DoAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	guildCats := getGuildAirouList(s)
	destination := bf.ReadUint32()
	level := bf.ReadUint32()
	huntData.WriteUint32(s.CharID)
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
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	database.Exec(`INSERT INTO guild_hunts (guild_id, host_id, destination, level, hunt_data, cats_used) VALUES ($1, $2, $3, $4, $5, $6)
		`, guild.ID, s.CharID, destination, level, huntData.Data(), catsUsed)
	broadcast.DoAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfAcquireGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildTresure)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	database.Exec(`UPDATE guild_hunts SET acquired=true WHERE id=$1`, pkt.HuntID)
	broadcast.DoAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfOperateGuildTresureReport(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuildTresureReport)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	switch pkt.State {
	case 0: // Report registration
		database.Exec(`UPDATE guild_characters SET treasure_hunt=$1 WHERE character_id=$2`, pkt.HuntID, s.CharID)
	case 1: // Collected by hunter
		database.Exec(`UPDATE guild_hunts SET collected=true WHERE id=$1`, pkt.HuntID)
		database.Exec(`UPDATE guild_characters SET treasure_hunt=NULL WHERE treasure_hunt=$1`, pkt.HuntID)
	case 2: // Claim treasure
		database.Exec(`INSERT INTO guild_hunts_claimed VALUES ($1, $2)`, pkt.HuntID, s.CharID)
	}
	broadcast.DoAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

type TreasureSouvenir struct {
	Destination uint32
	Quantity    uint32
}

func HandleMsgMhfGetGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildTresureSouvenir)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	souvenirs := []TreasureSouvenir{}
	bf.WriteUint16(uint16(len(souvenirs)))
	for _, souvenir := range souvenirs {
		bf.WriteUint32(souvenir.Destination)
		bf.WriteUint32(souvenir.Quantity)
	}
	broadcast.DoAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func HandleMsgMhfAcquireGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildTresureSouvenir)
	broadcast.DoAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
