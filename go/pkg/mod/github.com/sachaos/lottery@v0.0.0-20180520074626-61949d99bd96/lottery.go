package lottery

import (
	"math/rand"
	"time"
)

//go:generate mockgen -package lottery -source lottery.go -destination lottery_mock.go

type Weighter interface {
	Weight() int
}

type Lottery interface {
	Draw([]Weighter) int
}

type lottery struct {
	r *rand.Rand
}

func NewDefaultLottery() Lottery {
	return &lottery{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (l lottery) Draw(weighters []Weighter) int {
	if len(weighters) == 0 {
		return -1
	}

	totalWeight := 0
	for _, weighter := range weighters {
		totalWeight += weighter.Weight()
	}

	lot := l.r.Intn(totalWeight)

	tmp := 0
	for i, weighter := range weighters {
		tmp += weighter.Weight()
		if lot < tmp {
			return i
		}
	}
	return -1
}
