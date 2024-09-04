package events

import (
	"time"
)

type GameType int

const (
	Hamster GameType = iota
	Casino
	GrowKid
	Subscribe
)

type Base struct {
	LastPlay        time.Time `bson:"last_play"`
	PlayCount       uint16    `bson:"play_count"`
	MaxPlayCount    uint16    `bson:"-"`
	BasePlayPower   uint64    `bson:"-"`
	PercentagePower float64   `bson:"-"`
	Luck            int       `bson:"-"`
}

type Buff interface {
	Type() GameType
	Apply(*Base)
	Description() string
}
