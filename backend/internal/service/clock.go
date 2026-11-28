package service

import "time"

// Clock – abstrakcja czasu systemowego (testowalność Workera i serwisów).
type Clock interface {
	Now() time.Time
}

// SystemClock – zegar rzeczywisty.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// FixedClock – zegar o stałym wskazaniu (testy).
type FixedClock struct{ T time.Time }

func (c FixedClock) Now() time.Time { return c.T }
