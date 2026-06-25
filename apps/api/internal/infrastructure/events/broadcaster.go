package events

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID][]chan string
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[uuid.UUID][]chan string),
	}
}

func (b *Broadcaster) SubscribeTodoChanges(userID uuid.UUID) (<-chan string, func()) {
	b.mu.Lock()
	ch := make(chan string, 10)
	b.subscribers[userID] = append(b.subscribers[userID], ch)
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subscribers[userID]
		for i, sub := range subs {
			if sub == ch {
				b.subscribers[userID] = append(subs[:i], subs[i+1:]...)
				close(ch)
				break
			}
		}
		if len(b.subscribers[userID]) == 0 {
			delete(b.subscribers, userID)
		}
	}

	return ch, unsubscribe
}

func (b *Broadcaster) PublishTodoChange(ctx context.Context, userID uuid.UUID) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	subs, ok := b.subscribers[userID]
	if !ok {
		return
	}
	for _, sub := range subs {
		select {
		case sub <- "todo_update":
		default:
			// Non-blocking write to avoid hanging if client falls behind
		}
	}
}
