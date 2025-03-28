package event

import (
	"akira/internal/entity"
	"context"
	"sync"
	"slices"
)

var _ entity.EventService = (*Service)(nil)

type Service struct {
	subscribers entity.Subscribers
	mu          sync.RWMutex
	logger      entity.Logger
	ctx         context.Context
}

func NewService(ctx context.Context, logger entity.Logger) *Service {
	return &Service{
		subscribers: make(entity.Subscribers),
		logger:      logger,
		ctx:         ctx,
	}
}

func (s *Service) Subscribe(subID string, cancel context.CancelFunc, events ...entity.EventType) *entity.Subscriber {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub := &entity.Subscriber{
		SubID:  subID,
		Ch:     make(chan entity.Event, 100),
		Cancel: cancel,
		Filter: events,
	}
	s.subscribers[subID] = sub
	return sub
}

func (s *Service) Unsubscribe(subID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sub, ok := s.subscribers[subID]; ok {
		sub.Cancel()
		close(sub.Ch)
		delete(s.subscribers, subID)
	}
}

func (s *Service) shouldReceive(sub *entity.Subscriber, event entity.Event) bool {
	if len(sub.Filter) == 0 {
		return true
	}
	return slices.Contains(sub.Filter, event.Type)
}

func (s *Service) Publish(event entity.Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sub := range s.subscribers {
		if s.shouldReceive(sub, event) {
			select {
			case sub.Ch <- event:
			default:
				s.logger.Info(s.ctx, "subscriber channel is full, skipping", map[string]any{
					"subscriber": sub.SubID,
					"event_type": event.Type,
				})
			}
		}
	}
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, sub := range s.subscribers {
		sub.Cancel()
		close(sub.Ch)
		delete(s.subscribers, sub.SubID)
	}
	s.subscribers = make(entity.Subscribers)
	return nil
}
