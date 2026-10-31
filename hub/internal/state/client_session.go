package state

import "time"

type ClientSession struct {
	ConnectionId   string
	ClientId       string
	Token          string
	Tags           []string
	CreatedAt      time.Time
	LoggedInAt     time.Time
	LastActivityAt time.Time
}

// TODO constructor
