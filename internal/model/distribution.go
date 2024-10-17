package model

import "time"

type Distribution struct {
	ID              uint32    `db:"id"`
	Deadline        time.Time `db:"deadline"`
	Rights          uint32    `db:"rights"`
	TimesAcceptable uint16    `db:"times_acceptable"`
	TimesAccepted   uint16    `db:"times_accepted"`
	MinHR           int16     `db:"min_hr"`
	MaxHR           int16     `db:"max_hr"`
	MinSR           int16     `db:"min_sr"`
	MaxSR           int16     `db:"max_sr"`
	MinGR           int16     `db:"min_gr"`
	MaxGR           int16     `db:"max_gr"`
	EventName       string    `db:"event_name"`
	Description     string    `db:"description"`
	Selection       bool      `db:"selection"`
}

type DistributionItem struct {
	ItemType uint8  `db:"item_type"`
	ID       uint32 `db:"id"`
	ItemID   uint32 `db:"item_id"`
	Quantity uint32 `db:"quantity"`
}
