package constants

import (
	"time"

	"github.com/griffnb/core/lib/log"
)

const (
	DEFAULT_LOCATION        = "America/New_York"
	DEFAULT_LOCATION_ABR    = "EST"
	DEFAULT_LOCATION_OFFSET = "-05:00"
)

const SYSTEM_LIMIT = 400

func GetDefaultLocation() *time.Location {
	loc, err := time.LoadLocation(DEFAULT_LOCATION)
	if err != nil {
		log.Error(err)
		return nil
	}

	return loc
}
