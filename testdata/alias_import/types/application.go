package types

import "time"

type Application struct {
	Name string
}

type DateOnly time.Time

type TypeToEmbed struct {
	Embedded string
}
