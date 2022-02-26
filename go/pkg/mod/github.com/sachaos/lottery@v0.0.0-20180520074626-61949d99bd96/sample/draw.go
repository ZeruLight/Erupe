package main

import (
	"fmt"

	"github.com/sachaos/lottery"
)

type Item struct {
	Name   string
	weight int
}

func (i Item) Weight() int {
	return i.weight
}

func main() {
	l := lottery.NewDefaultLottery()
	items := []lottery.Weighter{
		&Item{Name: "high rare item", weight: 10},
		&Item{Name: "rare item", weight: 100},
		&Item{Name: "normal item", weight: 1000},
	}
	idx := l.Draw(items)
	fmt.Printf("You got %s!\n", items[idx].(*Item).Name)
}
