package model

import "time"

type TowerInfoTRP struct {
	TR  int32
	TRP int32
}

type TowerInfoSkill struct {
	TSP    int32
	Skills []int16 // 64
}

type TowerInfoHistory struct {
	Unk0 []int16 // 5
	Unk1 []int16 // 5
}

type TowerInfoLevel struct {
	Floors int32
	Unk1   int32
	Unk2   int32
	Unk3   int32
}

type GemInfo struct {
	Gem      uint16
	Quantity uint16
}

type GemHistory struct {
	Gem       uint16
	Message   uint16
	Timestamp time.Time
	Sender    string
}
type TenrouiraiProgress struct {
	Page     uint8
	Mission1 uint16
	Mission2 uint16
	Mission3 uint16
}

type TenrouiraiReward struct {
	Index    uint8
	Item     []uint16 // 5
	Quantity []uint8  // 5
}

type TenrouiraiKeyScore struct {
	Unk0 uint8
	Unk1 int32
}

type TenrouiraiData struct {
	Block   uint8
	Mission uint8
	// 1 = Floors climbed
	// 2 = Collect antiques
	// 3 = Open chests
	// 4 = Cats saved
	// 5 = TRP acquisition
	// 6 = Monster slays
	Goal   uint16
	Cost   uint16
	Skill1 uint8 // 80
	Skill2 uint8 // 40
	Skill3 uint8 // 40
	Skill4 uint8 // 20
	Skill5 uint8 // 40
	Skill6 uint8 // 50
}

type TenrouiraiCharScore struct {
	Score int32
	Name  string
}

type TenrouiraiTicket struct {
	Unk0 uint8
	RP   uint32
	Unk2 uint32
}

type Tenrouirai struct {
	Progress  []TenrouiraiProgress
	Reward    []TenrouiraiReward
	KeyScore  []TenrouiraiKeyScore
	Data      []TenrouiraiData
	CharScore []TenrouiraiCharScore
	Ticket    []TenrouiraiTicket
}
type TowerInfo struct {
	TRP     []TowerInfoTRP
	Skill   []TowerInfoSkill
	History []TowerInfoHistory
	Level   []TowerInfoLevel
}
