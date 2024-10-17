package model

import "time"

type GuildAdventure struct {
	ID          uint32 `db:"id"`
	Destination uint32 `db:"destination"`
	Charge      uint32 `db:"charge"`
	Depart      uint32 `db:"depart"`
	Return      uint32 `db:"return"`
	CollectedBy string `db:"collected_by"`
}

type GuildTreasureHunt struct {
	HuntID      uint32    `db:"id"`
	HostID      uint32    `db:"host_id"`
	Destination uint32    `db:"destination"`
	Level       uint32    `db:"level"`
	Start       time.Time `db:"start"`
	Acquired    bool      `db:"acquired"`
	Collected   bool      `db:"collected"`
	HuntData    []byte    `db:"hunt_data"`
	Hunters     uint32    `db:"hunters"`
	Claimed     bool      `db:"claimed"`
}
type GuildTreasureSouvenir struct {
	Destination uint32
	Quantity    uint32
}

type FestivalColor string
type GuildApplicationType string

type GuildIconPart struct {
	Index    uint16
	ID       uint16
	Page     uint8
	Size     uint8
	Rotation uint8
	Red      uint8
	Green    uint8
	Blue     uint8
	PosX     uint16
	PosY     uint16
}

type GuildApplication struct {
	ID              int                  `db:"id"`
	GuildID         uint32               `db:"guild_id"`
	CharID          uint32               `db:"character_id"`
	ActorID         uint32               `db:"actor_id"`
	ApplicationType GuildApplicationType `db:"application_type"`
	CreatedAt       time.Time            `db:"created_at"`
}

type GuildLeader struct {
	LeaderCharID uint32 `db:"leader_id"`
	LeaderName   string `db:"leader_name"`
}
type MessageBoardPost struct {
	ID        uint32    `db:"id"`
	StampID   uint32    `db:"stamp_id"`
	Title     string    `db:"title"`
	Body      string    `db:"body"`
	AuthorID  uint32    `db:"author_id"`
	Timestamp time.Time `db:"created_at"`
	LikedBy   string    `db:"liked_by"`
}

type GuildMeal struct {
	ID        uint32    `db:"id"`
	MealID    uint32    `db:"meal_id"`
	Level     uint32    `db:"level"`
	CreatedAt time.Time `db:"created_at"`
}

type GuildMission struct {
	ID          uint32
	Unk         uint32
	Type        uint16
	Goal        uint16
	Quantity    uint16
	SkipTickets uint16
	GR          bool
	RewardType  uint16
	RewardLevel uint16
}

type GuildAllianceInvite struct {
	GuildID    uint32
	LeaderID   uint32
	Unk0       uint16
	Unk1       uint16
	Members    uint16
	GuildName  string
	LeaderName string
}
type UnkGuildInfo struct {
	Unk0 uint8
	Unk1 uint8
	Unk2 uint8
}
