package model

import "time"

type HouseData struct {
	CharID        uint32 `db:"id"`
	HR            uint16 `db:"hr"`
	GR            uint16 `db:"gr"`
	Name          string `db:"name"`
	HouseState    uint8  `db:"house_state"`
	HousePassword string `db:"house_password"`
}
type Title struct {
	ID       uint16    `db:"id"`
	Acquired time.Time `db:"unlocked_at"`
	Updated  time.Time `db:"updated_at"`
}
