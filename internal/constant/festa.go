package constant

import "erupe-ce/internal/model"

const (
	FestivalColorNone model.FestivalColor = "none"
	FestivalColorBlue model.FestivalColor = "blue"
	FestivalColorRed  model.FestivalColor = "red"
)

var FestivalColorCodes = map[model.FestivalColor]int16{
	FestivalColorNone: -1,
	FestivalColorBlue: 0,
	FestivalColorRed:  1,
}
