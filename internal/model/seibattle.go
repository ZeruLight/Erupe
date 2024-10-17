package model

import "time"

type SeibattleTimetable struct {
	Start time.Time
	End   time.Time
}

type SeibattleKeyScore struct {
	Unk0 uint8
	Unk1 int32
}

type SeibattleCareer struct {
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
}

type SeibattleOpponent struct {
	Unk0 int32
	Unk1 int8
}

type SeibattleConventionResult struct {
	Unk0 uint32
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
	Unk4 uint16
}

type SeibattleCharScore struct {
	Unk0 uint32
}

type SeibattleCurResult struct {
	Unk0 uint32
	Unk1 uint16
	Unk2 uint16
	Unk3 uint16
}

type Seibattle struct {
	Timetable        []SeibattleTimetable
	KeyScore         []SeibattleKeyScore
	Career           []SeibattleCareer
	Opponent         []SeibattleOpponent
	ConventionResult []SeibattleConventionResult
	CharScore        []SeibattleCharScore
	CurResult        []SeibattleCurResult
}
