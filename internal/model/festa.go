package model

type FestaTrial struct {
	ID        uint32        `db:"id"`
	Objective uint16        `db:"objective"`
	GoalID    uint32        `db:"goal_id"`
	TimesReq  uint16        `db:"times_req"`
	Locale    uint16        `db:"locale_req"`
	Reward    uint16        `db:"reward"`
	Monopoly  FestivalColor `db:"monopoly"`
	Unk       uint16
}

type FestaReward struct {
	Unk0     uint8
	Unk1     uint8
	ItemType uint16
	Quantity uint16
	ItemID   uint16
	Unk5     uint16
	Unk6     uint16
	Unk7     uint8
}

type FestaPrize struct {
	ID       uint32 `db:"id"`
	Tier     uint32 `db:"tier"`
	SoulsReq uint32 `db:"souls_req"`
	ItemID   uint32 `db:"item_id"`
	NumItem  uint32 `db:"num_item"`
	Claimed  int    `db:"claimed"`
}
