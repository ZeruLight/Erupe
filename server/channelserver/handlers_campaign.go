package channelserver

import (
	"database/sql"
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"log"
	"time"
)

type CampaignEvent struct {
	ID           uint32    `db:"id"`
	MinHR        int16     `db:"min_hr"`
	MaxHR        int16     `db:"max_hr"`
	MinSR        int16     `db:"min_sr"`
	MaxSR        int16     `db:"max_sr"`
	MinGR        int16     `db:"min_gr"`
	MaxGR        int16     `db:"max_gr"`
	RecieveType  uint16    `db:"recieve_type"` //1 Item/Weapon //2 Event Quest //3 Item/Weapon
	StampAmount  uint8     `db:"stamp_amount"`
	Hide         uint8     `db:"hide"` //1 hides  // 10 and 0 seem to be visible  ?? is 10 overhang?
	BackgroundID uint16    `db:"background_id"`
	HideNPC      uint16    `db:"hide_npc"` //TODO: giving this the value 1 or above made the NPC glitch / hide
	Start        time.Time `db:"start_time"`
	End          time.Time `db:"end_time"`
	PeriodEnded  bool      `db:"period_ended"`
	String0      string    `db:"string0"`
	String1      string    `db:"string1"`
	String2      string    `db:"string2"`
	String3      string    `db:"string3"`
	Link         string    `db:"link"`
	Prefix       string    `db:"code_prefix"`
}

type CampaignCategory struct {
	ID          uint16 `db:"id"`
	Type        uint8  `db:"cat_type"`
	Title       string `db:"title"`
	Description string `db:"description_text"`
}

type CampaignLink struct {
	CategoryID uint16 `db:"category_id"`
	CampaignID uint32 `db:"campaign_id"`
}
type CampaignEntry struct {
	ID       uint32    `db:"id"`
	Hide     bool      `db:"hide"`
	ItemType uint8     `db:"item_type"`
	Amount   uint16    `db:"item_amount"`
	ItemNo   uint16    `db:"item_no"`
	Unk4     uint16    `db:"unk1"`
	Unk5     uint32    `db:"unk2"`
	DeadLine time.Time `db:"deadline"`
}

func handleMsgMhfEnumerateCampaign(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateCampaign)
	bf := byteframe.NewByteFrame()

	var events []CampaignEvent
	var categories []CampaignCategory

	var campaignLinks []CampaignLink

	err := s.server.db.Select(&events, "SELECT id,min_hr,max_hr,min_sr,max_sr,min_gr,max_gr,recieve_type,stamp_amount,hide,background_id,hide_npc,start_time,end_time,period_ended,string0,string1,string2,string3,link,code_prefix FROM campaigns")
	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	err = s.server.db.Select(&categories, "SELECT id, cat_type, title, description_text FROM campaign_categories")
	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	err = s.server.db.Select(&campaignLinks, "SELECT campaign_id, category_id FROM campaign_category_links")
	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if len(events) > 255 {
		bf.WriteUint8(255)
		bf.WriteUint16(uint16(len(events)))
	} else {
		bf.WriteUint8(uint8(len(events)))
	}
	for _, event := range events {
		bf.WriteUint32(event.ID)
		bf.WriteUint32(0) // always 0 in reference
		bf.WriteInt16(event.MinHR)
		bf.WriteInt16(event.MaxHR)
		bf.WriteInt16(event.MinSR)
		bf.WriteInt16(event.MaxSR)
		if _config.ErupeConfig.RealClientMode >= _config.G3 {
			bf.WriteInt16(event.MinGR)
			bf.WriteInt16(event.MaxGR)
		}
		bf.WriteUint16(event.RecieveType)
		bf.WriteUint8(event.StampAmount)
		bf.WriteUint8(event.Hide)
		bf.WriteUint16(event.BackgroundID)
		bf.WriteUint16(event.HideNPC)
		bf.WriteUint32(uint32(event.Start.Unix()))
		bf.WriteUint32(uint32(event.End.Unix()))
		bf.WriteBool(event.PeriodEnded)
		ps.Uint8(bf, event.String0, true)
		ps.Uint8(bf, event.String1, true)
		ps.Uint8(bf, event.String2, true)
		ps.Uint8(bf, event.String3, true)
		ps.Uint8(bf, event.Link, true)

	}

	if len(events) > 255 {
		bf.WriteUint8(255)
		bf.WriteUint16(uint16(len(events)))
	} else {
		bf.WriteUint8(uint8(len(events)))
	}
	for _, event := range events {
		bf.WriteUint32(event.ID)
		bf.WriteUint8(1) //StampAmount * This Amount = Stamps Shown
		bf.WriteBytes([]byte(event.Prefix))
	}

	if len(categories) > 255 {
		bf.WriteUint8(255)
		bf.WriteUint16(uint16(len(categories)))
	} else {
		bf.WriteUint8(uint8(len(categories)))
	}
	for _, category := range categories {
		bf.WriteUint16(category.ID)
		bf.WriteUint8(category.Type)
		xTitle := stringsupport.UTF8ToSJIS(category.Title)
		xDescription := stringsupport.UTF8ToSJIS(category.Description)
		bf.WriteUint8(uint8(len(xTitle)))
		bf.WriteUint8(uint8(len(xDescription)))
		bf.WriteBytes(xTitle)
		bf.WriteBytes(xDescription)
	}

	if len(campaignLinks) > 255 {
		bf.WriteUint8(255)
		bf.WriteUint16(uint16(len(campaignLinks)))
	} else {
		bf.WriteUint8(uint8(len(campaignLinks)))
	}
	for _, link := range campaignLinks {
		bf.WriteUint16(link.CategoryID)
		bf.WriteUint32(link.CampaignID)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfStateCampaign(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateCampaign)
	bf := byteframe.NewByteFrame()
	var stamps, amount int
	var period_ended bool
	err := s.server.db.QueryRow(`SELECT COUNT(*) FROM campaign_state WHERE campaign_id = $1 AND character_id = $2`, pkt.CampaignID, s.charID).Scan(&stamps)
	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	err = s.server.db.QueryRow(`SELECT stamp_amount,period_ended FROM campaigns WHERE id = $1`, pkt.CampaignID).Scan(&amount, &period_ended)
	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	var unkArray = []uint32{1, 2, 3, 4, 5, 6} //Not figured out what this does yet....?
	bf.WriteUint16(uint16(stamps + 1))        // game client seems to -1
	if amount == 1 {
		bf.WriteUint16(1)
	} else if stamps >= amount || period_ended {
		bf.WriteUint16(2)
	} else if amount > 1 {
		bf.WriteUint16(3)
	}
	for _, value := range unkArray {
		bf.WriteUint32(value)
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfApplyCampaign(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfApplyCampaign)

	var code string
	// Query to check if the event code exists in the database
	err := s.server.db.QueryRow("SELECT code FROM public.campaign_state WHERE code = $1", pkt.CodeString).Scan(&code)

	switch {
	case err == sql.ErrNoRows:
		fmt.Println("Event code does not exist in the database.")
		s.server.db.Exec(`INSERT INTO public.campaign_state ( code, campaign_id ,character_id) VALUES ($1,$2,$3)`, pkt.CodeString, pkt.CampaignID, s.charID)

		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Printf("Event code '%s' exists in the database.\n", code)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))

	}

}

func handleMsgMhfEnumerateItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateItem)
	bf := byteframe.NewByteFrame()

	var stamps, amount uint16
	err := s.server.db.QueryRow(`SELECT COUNT(*) FROM campaign_state WHERE campaign_id = $1 AND character_id = $2`, pkt.CampaignID, s.charID).Scan(&stamps)
	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	err = s.server.db.QueryRow(`SELECT stamp_amount FROM campaigns WHERE id = $1`, pkt.CampaignID).Scan(&amount)
	if err != nil {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if stamps >= amount {

		var items []CampaignEntry
		err = s.server.db.Select(&items, `SELECT id,hide,item_type,item_amount,item_no,unk1,unk2,deadline FROM campaign_entries WHERE campaign_id = $1`, pkt.CampaignID)
		if err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		bf.WriteUint16(uint16(len(items)))
		for _, item := range items {
			bf.WriteUint32(item.ID)
			bf.WriteBool(item.Hide)
			bf.WriteUint8(item.ItemType)
			bf.WriteUint16(item.Amount)
			bf.WriteUint16(item.ItemNo) //HACK:placed quest id in this field to fit with Item No pattern. however it could be another field... possibly the other unks.
			bf.WriteUint16(item.Unk4)
			bf.WriteUint32(item.Unk5)
			bf.WriteUint32(uint32(item.DeadLine.Unix()))
		}
		if len(items) == 0 {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))

		} else {

			doAckBufSucceed(s, pkt.AckHandle, bf.Data())

		}
	} else {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgMhfAcquireItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireItem)

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
