package model

type Achievement struct {
	Level     uint8
	Value     uint32
	NextValue uint16
	Required  uint32
	Updated   bool
	Progress  uint32
	Trophy    uint8
}
