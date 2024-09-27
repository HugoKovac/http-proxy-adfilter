package types

import (
	"time"
)

type DomainObject struct {
	category string
}

type DomainList struct {
	list map[string]DomainObject	`json:"list"`
	lastUpdate time.Time			`json:"last_update"`
}
