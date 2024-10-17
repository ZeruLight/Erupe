package model

import "time"

type CampaignEvent struct {
	ID         uint32
	Unk0       uint32
	MinHR      int16
	MaxHR      int16
	MinSR      int16
	MaxSR      int16
	MinGR      int16
	MaxGR      int16
	Unk1       uint16
	Unk2       uint8
	Unk3       uint8
	Unk4       uint16
	Unk5       uint16
	Start      time.Time
	End        time.Time
	Unk6       uint8
	String0    string
	String1    string
	String2    string
	String3    string
	Link       string
	Prefix     string
	Categories []uint16
}

type CampaignCategory struct {
	ID          uint16
	Type        uint8
	Title       string
	Description string
}

type CampaignLink struct {
	CategoryID uint16
	CampaignID uint32
}
