package channelserver

import (
	"time"
)

func TimeAdjusted() time.Time {
	baseTime := time.Now().In(time.FixedZone("UTC+9", 9*60*60))
	return time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), baseTime.Hour(), baseTime.Minute(), baseTime.Second(), baseTime.Nanosecond(), baseTime.Location())
}

func TimeMidnight() time.Time {
	baseTime := time.Now().In(time.FixedZone("UTC+9", 9*60*60))
	return time.Date(baseTime.Year(), baseTime.Month(), baseTime.Day(), 0, 0, 0, 0, baseTime.Location())
}

func TimeWeekStart() time.Time {
	midnight := TimeMidnight()
	offset := int(midnight.Weekday()) - int(time.Monday)
	if offset < 0 {
		offset += 7
	}
	return midnight.Add(-time.Duration(offset) * 24 * time.Hour)
}

func TimeWeekNext() time.Time {
	return TimeWeekStart().Add(time.Hour * 24 * 7)
}

func TimeGameAbsolute() uint32 {
	return uint32((TimeAdjusted().Unix() - 2160) % 5760)
}
