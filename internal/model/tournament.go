package model

import "time"

type TournamentInfo0 struct {
	ID             uint32
	MaxPlayers     uint32
	CurrentPlayers uint32
	Unk1           uint16
	TextColor      uint16
	Unk2           uint32
	Time1          time.Time
	Time2          time.Time
	Time3          time.Time
	Time4          time.Time
	Time5          time.Time
	Time6          time.Time
	Unk3           uint8
	Unk4           uint8
	MinHR          uint32
	MaxHR          uint32
	Unk5           string
	Unk6           string
}

type TournamentInfo21 struct {
	Unk0 uint32
	Unk1 uint32
	Unk2 uint32
	Unk3 uint8
}

type TournamentInfo22 struct {
	Unk0 uint32
	Unk1 uint32
	Unk2 uint32
	Unk3 uint8
	Unk4 string
}

type TournamentReward struct {
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
}
