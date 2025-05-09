package subpub

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPublishSubscribeFIFO(t *testing.T) {
	b := NewSubPub()
	defer b.Close(context.Background())

	var mu sync.Mutex
	var got []int

	sub, err := b.Subscribe("topic", func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, msg.(int))
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		if err := b.Publish("topic", i); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(50 * time.Millisecond)
	sub.Unsubscribe()
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	for i := 0; i < 5; i++ {
		if got[i] != i {
			t.Errorf("expected %d, got %d", i, got[i])
		}
	}
}

func TestSlowSubscriberDoesNotBlock(t *testing.T) {
	b := NewSubPub()
	defer b.Close(context.Background())

	chFast := make(chan struct{})
	b.Subscribe("topic", func(msg interface{}) {
		chFast <- struct{}{}
	})

	b.Subscribe("topic", func(msg interface{}) {
		time.Sleep(100 * time.Millisecond)
	})

	go b.Publish("topic", "hello")

	select {
	case <-chFast:
	case <-time.After(20 * time.Millisecond):
		t.Error("fast subscriber was blocked by slow one")
	}
}

func TestCloseContextHonors(t *testing.T) {
	ctx := context.Background()
	b := NewSubPub()
	sub, _ := b.Subscribe("topic", func(msg interface{}) {
		<-ctx.Done()
	})

	_ = b.Publish("topic", "x")
	c, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := b.Close(c)
	if err == nil {
		t.Error("expected context deadline exceeded")
	}
	sub.Unsubscribe()
}
