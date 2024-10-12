package channelserver

import (
	"fmt"
	"time"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/utils/byteframe"
	"erupe-ce/utils/db"
	"erupe-ce/utils/gametime"
	"erupe-ce/utils/stringsupport"

	"go.uber.org/zap"
)

type GuildAdventure struct {
	ID          uint32 `db:"id"`
	Destination uint32 `db:"destination"`
	Charge      uint32 `db:"charge"`
	Depart      uint32 `db:"depart"`
	Return      uint32 `db:"return"`
	CollectedBy string `db:"collected_by"`
}

func HandleMsgMhfLoadGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadGuildAdventure)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	guild, _ := GetGuildInfoByCharacterId(s, s.CharID)
	data, err := database.Queryx("SELECT id, destination, charge, depart, return, collected_by FROM guild_adventures WHERE guild_id = $1", guild.ID)
	if err != nil {
		s.Logger.Error("Failed to get guild adventures from db", zap.Error(err))
		DoAckBufSucceed(s, pkt.AckHandle, make([]byte, 1))
		return
	}
	temp := byteframe.NewByteFrame()
	count := 0
	for data.Next() {
		count++
		adventureData := &GuildAdventure{}
		err = data.StructScan(&adventureData)
		if err != nil {
			continue
		}
		temp.WriteUint32(adventureData.ID)
		temp.WriteUint32(adventureData.Destination)
		temp.WriteUint32(adventureData.Charge)
		temp.WriteUint32(adventureData.Depart)
		temp.WriteUint32(adventureData.Return)
		temp.WriteBool(stringsupport.CSVContains(adventureData.CollectedBy, int(s.CharID)))
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(count))
	bf.WriteBytes(temp.Data())
	DoAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func HandleMsgMhfRegistGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildAdventure)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	guild, _ := GetGuildInfoByCharacterId(s, s.CharID)
	_, err = database.Exec("INSERT INTO guild_adventures (guild_id, destination, depart, return) VALUES ($1, $2, $3, $4)", guild.ID, pkt.Destination, gametime.TimeAdjusted().Unix(), gametime.TimeAdjusted().Add(6*time.Hour).Unix())
	if err != nil {
		s.Logger.Error("Failed to register guild adventure", zap.Error(err))
	}
	DoAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfAcquireGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildAdventure)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	var collectedBy string
	err = database.QueryRow("SELECT collected_by FROM guild_adventures WHERE id = $1", pkt.ID).Scan(&collectedBy)
	if err != nil {
		s.Logger.Error("Error parsing adventure collected by", zap.Error(err))
	} else {
		collectedBy = stringsupport.CSVAdd(collectedBy, int(s.CharID))
		_, err := database.Exec("UPDATE guild_adventures SET collected_by = $1 WHERE id = $2", collectedBy, pkt.ID)
		if err != nil {
			s.Logger.Error("Failed to collect adventure in db", zap.Error(err))
		}
	}
	DoAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfChargeGuildAdventure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfChargeGuildAdventure)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	_, err = database.Exec("UPDATE guild_adventures SET charge = charge + $1 WHERE id = $2", pkt.Amount, pkt.ID)
	if err != nil {
		s.Logger.Error("Failed to charge guild adventure", zap.Error(err))
	}
	DoAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func HandleMsgMhfRegistGuildAdventureDiva(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildAdventureDiva)
	database, err := db.GetDB()
	if err != nil {
		s.Logger.Fatal(fmt.Sprintf("Failed to get database instance: %s", err))
	}
	guild, _ := GetGuildInfoByCharacterId(s, s.CharID)
	_, err = database.Exec("INSERT INTO guild_adventures (guild_id, destination, charge, depart, return) VALUES ($1, $2, $3, $4, $5)", guild.ID, pkt.Destination, pkt.Charge, gametime.TimeAdjusted().Unix(), gametime.TimeAdjusted().Add(1*time.Hour).Unix())
	if err != nil {
		s.Logger.Error("Failed to register guild adventure", zap.Error(err))
	}
	DoAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
