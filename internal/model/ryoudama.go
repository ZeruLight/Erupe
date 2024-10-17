package model

import "time"

type RyoudamaReward struct {
	Unk0 uint8
	Unk1 uint8
	Unk2 uint16
	Unk3 uint16
	Unk4 uint16
	Unk5 uint16
}

type RyoudamaKeyScore struct {
	Unk0 uint8
	Unk1 int32
}

type RyoudamaCharInfo struct {
	CID  uint32
	Unk0 int32
	Name string
}

type RyoudamaBoostInfo struct {
	Start time.Time
	End   time.Time
}

type Ryoudama struct {
	Reward    []RyoudamaReward
	KeyScore  []RyoudamaKeyScore
	CharInfo  []RyoudamaCharInfo
	BoostInfo []RyoudamaBoostInfo
	Score     []int32
}
