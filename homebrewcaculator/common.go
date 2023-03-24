package homebrewcaculator

import "time"

type Span int

const (
	ThirtyDays Span = 30
	NinetyDays Span = 90
	OneYear    Span = 365
)

func PossibleSpans() []Span {
	return []Span{ThirtyDays, NinetyDays, OneYear}
}

type Data struct {
	CountDate    time.Time
	TodayCount   int32 //today's count, set to -1 if not calculated yet.
	Count        int32 //total count
	PreviousData map[Span]int32
}

// Date to keep the date context are in same format (only keeps date part of time.Time)
type Date struct {
	time.Time
}

func (d Date) AddDays(days int) Date {
	return Date{d.Time.AddDate(0, 0, days)}
}
