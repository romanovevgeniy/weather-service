package domain

import "time"

type Reading struct {
	Name        string
	Timestamp   time.Time
	Temperature float64
}
