package message

import (
	"time"
)

type InternalMessage struct {
	ID        int
	Value     int
	Rx        string
	Type      string
	Timestamp time.Time
}
