package event

import (
	"akira/internal/entity"
	"context"
	"slices"
	"sync"
	"time"
)

var _ entity.EventService = (*Service)(nil)

type Service struct {
	subscribers     entity.Subscribers
	mu              sync.RWMutex
	logger          entity.Logger
	ctx             context.Context
	deadLetterQueue []entity.Event
	maxDeadLetters  int
	eventsub        *entity.Subscriber
	cancelFunc      context.CancelFunc
}

func NewService(ctx context.Context, logger entity.Logger) *Service {
	cancel, cancelFunc := context.WithCancel(ctx)
	service := &Service{
		subscribers:     make(entity.Subscribers),
		logger:          logger,
		ctx:             ctx,
		deadLetterQueue: make([]entity.Event, 0),
		maxDeadLetters:  100,
		cancelFunc:      cancelFunc,
	}
	service.eventsub = service.Subscribe(
		"event-service",
		cancelFunc,
		entity.EventSystemStarted,
		entity.EventSystemError,
		entity.EventSystemShutdown,
	)
	go service.consumeEvents(cancel)
	return service
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

func (s *Service) Publish(event entity.Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	deliveredCount := 0
	for id, sub := range s.subscribers {
		if s.shouldReceive(sub, event) {
			select {
			case sub.Ch <- event:
				deliveredCount++
			case <-time.After(100 * time.Millisecond):
				s.recordDeadLetter(event, "channel timeout", id)
			default:
				s.recordDeadLetter(event, "channel full", id)
				s.logger.Warn(s.ctx, "subscriber channel is full, skipping", map[string]any{
					"subscriber": id,
					"event_type": event.Type,
				})
			}
		}
	}
	s.logger.Debug(s.ctx, "event published", map[string]any{
		"event_type":   event.Type,
		"delivered_to": deliveredCount,
		"subscribers":  len(s.subscribers),
		"user_id":      event.UserID,
	})
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
	s.cancelFunc()
	return nil
}

func (s *Service) shouldReceive(sub *entity.Subscriber, event entity.Event) bool {
	if len(sub.Filter) == 0 {
		return true
	}
	return slices.Contains(sub.Filter, event.Type)
}

func (s *Service) recordDeadLetter(event entity.Event, reason string, subID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Warn(s.ctx, "event added to dead letter queue", map[string]any{
		"event_type":   event.Type,
		"reason":       reason,
		"subscriber":   subID,
		"queue_length": len(s.deadLetterQueue),
	})
	if len(s.deadLetterQueue) >= s.maxDeadLetters {
		s.deadLetterQueue = s.deadLetterQueue[1:]
	}
	s.deadLetterQueue = append(s.deadLetterQueue, event)
}

func (s *Service) consumeEvents(ctx context.Context) {
	for {
		select {
		case event, ok := <-s.eventsub.Ch:
			if !ok {
				s.logger.Warn(s.ctx, "event subscriber channel closed", nil)
				return
			}
			go func(event entity.Event) {
				defer func() {
					if r := recover(); r != nil {
						s.logger.Warn(s.ctx, "panic in event handler", map[string]any{
							"event_type": event.Type,
							"recover":    r,
						})
					}
				}()
				switch event.Type {
				case entity.EventSystemStarted:
					s.logger.Info(s.ctx, "system started", nil)
				case entity.EventSystemShutdown:
					s.logger.Info(s.ctx, "system shutdown", nil)
				case entity.EventSystemError:
					s.logger.Error(s.ctx, "system error", nil, nil)
				}
			}(event)
		case <-ctx.Done():
			s.logger.Info(s.ctx, "event service shutting down", nil)
			return
		}
	}
}
