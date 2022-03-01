package types

import "time"

type Application struct {
	Name string
}

type Application2 struct {
	Name string
}

type DateOnly time.Time

type TypeToEmbed struct {
	Embedded string
}

type ShouldSkip struct {
	Name string
}
