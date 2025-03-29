package entity

import (
	"context"
	"time"
)

type EventType string

const (
	EventCollectionCreated       EventType = "collection:created"
	EventCollectionUpdated       EventType = "collection:updated"
	EventCollectionDeleted       EventType = "collection:deleted"
	EventCollectionSyncFetching  EventType = "collection:sync-fetching"
	EventCollectionSyncCompleted EventType = "collection:sync-completed"
	EventCollectionSyncFailed    EventType = "collection:sync-failed"
	EventSystemStarted           EventType = "system:started"
	EventSystemShutdown          EventType = "system:shutdown"
	EventSystemError             EventType = "system:error"
)

type Event struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id,omitempty"`
	Data      any       `json:"data"`
}

func NewEvent(eventType EventType, userID string, data any) Event {
	return Event{
		Type:      eventType,
		Timestamp: time.Now().UTC(),
		UserID:    userID,
		Data:      data,
	}
}

type Subscriber struct {
	SubID  string
	Ch     chan Event
	Cancel context.CancelFunc
	Filter []EventType
}

type Subscribers map[string]*Subscriber

type EventService interface {
	Subscribe(subID string, cancel context.CancelFunc, events ...EventType) *Subscriber
	Unsubscribe(subID string)
	Publish(event Event)
	Shutdown(ctx context.Context) error
}
