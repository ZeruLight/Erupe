package model

import "time"

type Event struct {
	EventType    uint16
	Unk1         uint16
	Unk2         uint16
	Unk3         uint16
	Unk4         uint16
	Unk5         uint32
	Unk6         uint32
	QuestFileIDs []uint16
}

type LoginBoost struct {
	WeekReq    uint8 `db:"week_req"`
	WeekCount  uint8
	Active     bool
	Expiration time.Time `db:"expiration"`
	Reset      time.Time `db:"reset"`
}

type ActiveFeature struct {
	StartTime      time.Time `db:"start_time"`
	ActiveFeatures uint32    `db:"featured"`
}
type TrendWeapon struct {
	WeaponType uint8
	WeaponID   uint16
}
