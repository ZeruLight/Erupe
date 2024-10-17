package model

import "time"

type PaperMissionTimetable struct {
	Start time.Time
	End   time.Time
}

type PaperMissionData struct {
	Unk0            uint8
	Unk1            uint8
	Unk2            int16
	Reward1ID       uint16
	Reward1Quantity uint8
	Reward2ID       uint16
	Reward2Quantity uint8
}

type PaperMission struct {
	Timetables []PaperMissionTimetable
	Data       []PaperMissionData
}

type PaperData struct {
	Unk0 uint16
	Unk1 int16
	Unk2 int16
	Unk3 int16
	Unk4 int16
	Unk5 int16
	Unk6 int16
}

type PaperGift struct {
	Unk0 uint16
	Unk1 uint8
	Unk2 uint8
	Unk3 uint16
}
