package service

import "time"

// Abstrakcja czasu systemowego.
type Clock interface {
	Now() time.Time
}

// Zegar rzeczywisty.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// Zegar o stałym wskazaniu (dla przyszłych testów).
type FixedClock struct{ T time.Time }

func (c FixedClock) Now() time.Time { return c.T }
