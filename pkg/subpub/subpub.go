package subpub

import (
	"context"
	"errors"
	"sync"
)

var ErrClosed = errors.New("subpub: broker closed")

const messageBuffer = 16

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)
	Publish(subject string, msg interface{}) error
	Close(ctx context.Context) error
}

func NewSubPub() SubPub {
	return &broker{
		topics: make(map[string]map[*subscriber]struct{}),
		closed: make(chan struct{}),
	}
}

type subscriber struct {
	ch   chan interface{}
	once sync.Once
	done chan struct{}
}

func (s *subscriber) Unsubscribe() {
	s.once.Do(func() {
		close(s.done)
	})
}

type broker struct {
	mu      sync.RWMutex
	topics  map[string]map[*subscriber]struct{}
	closing bool          // вызвали close и больше не выываем publish или subscribe
	closed  chan struct{} // закрываем при close, чтобы все активные горутины тоже завершились
	wg      sync.WaitGroup
}

func (b *broker) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closing {
		return nil, ErrClosed
	}

	sub := &subscriber{
		done: make(chan struct{}),
		ch:   make(chan interface{}, messageBuffer),
	}

	if b.topics[subject] == nil {
		b.topics[subject] = make(map[*subscriber]struct{})
	}
	b.topics[subject][sub] = struct{}{}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			case msg, ok := <-sub.ch:
				if !ok {
					return
				}
				cb(msg)
			case <-sub.done:
				return
			case <-b.closed:
				return
			}
		}
	}()

	return sub, nil
}

func (b *broker) Publish(subject string, msg interface{}) error {
	b.mu.RLock()
	defer b.mu.Unlock()

	if b.closing {
		return ErrClosed
	}

	subs := b.topics[subject]

	for sub := range subs {
		select {
		case sub.ch <- msg:
		default:
			sub.ch <- msg
		}
	}

	return nil
}

func (b *broker) Close(ctx context.Context) error {
	b.mu.Lock()
	if b.closing {
		b.mu.Unlock()
		return ErrClosed
	}
	b.closing = true
	close(b.closed)
	for _, subs := range b.topics {
		for sub := range subs {
			close(sub.ch)
		}
	}
	b.mu.Unlock()

	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
