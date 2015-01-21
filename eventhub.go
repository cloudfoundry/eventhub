package eventhub

import "sync"

const MAX_PENDING_SUBSCRIBER_EVENTS = 1024

type Hub interface {
	Subscribe() (Source, error)
	Emit(Event)
	Close() error
	HasSubscribers() bool
}

type hub struct {
	subscribers []Source
	closed      bool
	lock        sync.Mutex
}

func New() Hub {
	return &hub{}
}

func (hub *hub) Subscribe() (Source, error) {
	hub.lock.Lock()
	defer hub.lock.Unlock()

	if hub.closed {
		return nil, ErrSubscribedToClosedHub
	}

	sub := newSource(MAX_PENDING_SUBSCRIBER_EVENTS)
	hub.subscribers = append(hub.subscribers, sub)

	return sub, nil
}

func (hub *hub) Emit(event Event) {
	hub.lock.Lock()
	defer hub.lock.Unlock()

	remainingSubscribers := make([]Source, 0, len(hub.subscribers))

	for _, sub := range hub.subscribers {
		err := sub.send(event)
		if err == nil {
			remainingSubscribers = append(remainingSubscribers, sub)
		}
	}

	hub.subscribers = remainingSubscribers
}

func (hub *hub) Close() error {
	hub.lock.Lock()
	defer hub.lock.Unlock()

	if hub.closed {
		return ErrHubAlreadyClosed
	}

	hub.closeSubscribers()
	hub.closed = true

	return nil
}

func (hub *hub) HasSubscribers() bool {
	hub.lock.Lock()
	defer hub.lock.Unlock()

	return len(hub.subscribers) != 0
}

func (hub *hub) closeSubscribers() {
	for _, sub := range hub.subscribers {
		_ = sub.Close()
	}
	hub.subscribers = nil
}
