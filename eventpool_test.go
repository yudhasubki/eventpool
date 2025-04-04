package eventpool

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddSubscriberEventPool(t *testing.T) {
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
		EventpoolListener{
			Name: "test2",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
	}
	e := New()
	e.Submit(listeners...)

	assert.Equal(t, 2, len(e.Subscribers()))
}

func TestAddSubscriberEventOnFlightPool(t *testing.T) {
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
		EventpoolListener{
			Name: "test2",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
	}
	e := New()
	e.Submit(listeners...)

	assert.Equal(t, 2, len(e.Subscribers()))

	e.Run()

	e.SubmitOnFlight(EventpoolListener{
		Name: "test3",
		Subscriber: func(name string, message []byte) error {
			return nil
		},
	})

	assert.Equal(t, 3, len(e.Subscribers()))

	e.Close()
}

func TestPublishMessage(t *testing.T) {
	var count uint64
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message []byte) error {
				atomic.AddUint64(&count, 1)

				return nil
			},
		},
	}
	e := New()
	e.Submit(listeners...)
	e.Run()
	time.Sleep(3 * time.Second)
	for i := 0; i < 10; i++ {
		e.Publish(SendString(fmt.Sprint(i)))
	}
	time.Sleep(1 * time.Second)
	assert.Equal(t, uint64(10), atomic.LoadUint64(&count))
}

func TestCapacitySubscriber(t *testing.T) {
	var messageCount = 10
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message []byte) error {
				time.Sleep(2 * time.Second)

				return nil
			},
		},
	}
	e := New()
	e.Submit(listeners...)
	e.Run()

	for i := 0; i < messageCount; i++ {
		e.Publish(SendString(fmt.Sprint(i)))
	}
	time.Sleep(1 * time.Second)
	assert.Equal(t, e.Cap("test1"), 10)
	time.Sleep(3 * time.Second)
	assert.Equal(t, e.Cap("test1"), 0)
	e.Close()
}

func TestDeleteSubscriber(t *testing.T) {
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
		EventpoolListener{
			Name: "test2",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
	}
	e := New()
	e.Submit(listeners...)
	e.Run()

	assert.Equal(t, 2, len(e.Subscribers()))
	e.CloseBy("test2")
	assert.Equal(t, 1, len(e.Subscribers()))
	e.Close()
}

func TestParitionAddSubscriberEventOnFlightPool(t *testing.T) {
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
		EventpoolListener{
			Name: "test2",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
	}
	e := NewPartition(3)
	e.Submit(listeners...)

	assert.Equal(t, 2, len(e.Subscribers()))

	e.Run()

	e.SubmitOnFlight(EventpoolListener{
		Name: "test3",
		Subscriber: func(name string, message []byte) error {
			return nil
		},
	})

	assert.Equal(t, 3, len(e.Subscribers()))

	e.Close()
}

func TestCapacitySubscriberPartition(t *testing.T) {
	var messageCount = 10
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message []byte) error {
				time.Sleep(2 * time.Second)

				return nil
			},
		},
	}
	e := NewPartition(64)
	e.Submit(listeners...)
	e.Run()
	for i := 0; i < messageCount; i++ {
		e.Publish("*", fmt.Sprint(i), SendString(fmt.Sprint(i)))
	}
	time.Sleep(1 * time.Second)
	assert.Equal(t, 10, e.Cap("test1"))
	time.Sleep(3 * time.Second)
	assert.Equal(t, e.Cap("test1"), 0)
	e.Close()
}

func TestDeleteSubscriberPartition(t *testing.T) {
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
		EventpoolListener{
			Name: "test2",
			Subscriber: func(name string, message []byte) error {
				return nil
			},
		},
	}
	e := NewPartition(1)
	e.Submit(listeners...)
	e.Run()

	assert.Equal(t, 2, len(e.Subscribers()))
	e.CloseBy("test2")
	assert.Equal(t, 1, len(e.Subscribers()))
	e.Close()
}

func testMessageFunc() messageFunc {
	return SendString("hello")
}

func dummySubscriber(name string, r []byte) error {

	return nil
}

func BenchmarkEventWildcardByPartition(b *testing.B) {
	ep := NewPartition(16)

	listeners := []EventpoolListener{
		{
			Name:       "groupA",
			Subscriber: dummySubscriber,
		},
		{
			Name:       "groupB",
			Subscriber: dummySubscriber,
		},
	}

	ep.Submit(listeners...)
	ep.Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ep.Publish("*", "", testMessageFunc())
	}
}

func BenchmarkEventSpecificGroupByPartition(b *testing.B) {
	ep := NewPartition(16)

	listeners := []EventpoolListener{
		{
			Name:       "groupA",
			Subscriber: dummySubscriber,
		},
		{
			Name:       "groupB",
			Subscriber: dummySubscriber,
		},
	}

	ep.Submit(listeners...)
	ep.Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ep.Publish("groupA", "", testMessageFunc())
	}
}

func BenchmarkMultipleEventByBroadcast(b *testing.B) {
	ep := New()

	listeners := []EventpoolListener{
		{
			Name:       "workerA",
			Subscriber: dummySubscriber,
		},
		{
			Name:       "workerB",
			Subscriber: dummySubscriber,
		},
	}

	ep.SubmitOnFlight(listeners...)
	ep.Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ep.Publish(testMessageFunc())
	}
}

func BenchmarkSingleEventByBroadcast(b *testing.B) {
	ep := New()

	listeners := []EventpoolListener{
		{
			Name:       "workerA",
			Subscriber: dummySubscriber,
		},
	}

	ep.SubmitOnFlight(listeners...)
	ep.Run()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ep.Publish(testMessageFunc())
	}
}
