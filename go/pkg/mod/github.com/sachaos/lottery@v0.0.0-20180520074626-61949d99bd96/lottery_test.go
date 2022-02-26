package lottery

import (
	"math"
	"testing"
)

type TestItem struct {
	Name string
	w    int
}

func (i TestItem) Weight() int {
	return i.w
}

func TestDraw(t *testing.T) {
	lottery := NewDefaultLottery()

	t.Run("Counting", func(t *testing.T) {
		counter := map[int]int{}
		items := []Weighter{
			&TestItem{Name: "item1", w: 50},
			&TestItem{Name: "item2", w: 50},
			&TestItem{Name: "item2", w: 20},
		}
		totalWeight := 120
		totalCount := 10000000

		for i := 0; i < totalCount; i++ {
			idx := lottery.Draw(items)
			counter[idx] += 1
		}

		for idx, item := range items {
			expectedCount := int(float64(item.Weight()) / float64(totalWeight) * float64(totalCount))
			actualCount := counter[idx]

			deflection := math.Abs(float64(expectedCount - actualCount))

			t.Logf("idx: %d, expectedCount: %d, actualCount: %d", idx, expectedCount, actualCount)
			if deflection >= float64(expectedCount)*0.1 {
				t.Errorf("Deflection over 0.1")
			}
		}
	})

	t.Run("Fail input", func(t *testing.T) {
		items := []Weighter{}
		idx := lottery.Draw(items)
		if idx != -1 {
			t.Errorf("Expected response: -1, but actual: %v", idx)
		}
	})
}
