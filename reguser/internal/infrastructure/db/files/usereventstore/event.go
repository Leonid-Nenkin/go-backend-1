package usereventstore

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventCreate EventType = "create"
	EventDelete EventType = "delete"
)

type EventUser struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name,omitempty"`
	Data        string    `json:"data,omitempty"`
	Permissions int       `json:"perms,omitempty"`
}

type Event struct {
	TimeStamp time.Time  `json:"timestamp"`
	Type      EventType  `json:"eventType"`
	User      *EventUser `json:"user,omitempty"`
}
