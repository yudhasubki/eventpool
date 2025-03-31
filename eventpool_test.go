package eventpool

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddSubscriberEventPool(t *testing.T) {
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message io.Reader) error {
				return nil
			},
		},
		EventpoolListener{
			Name: "test2",
			Subscriber: func(name string, message io.Reader) error {
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
			Subscriber: func(name string, message io.Reader) error {
				return nil
			},
		},
		EventpoolListener{
			Name: "test2",
			Subscriber: func(name string, message io.Reader) error {
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
		Subscriber: func(name string, message io.Reader) error {
			return nil
		},
	})

	assert.Equal(t, 3, len(e.Subscribers()))

	e.Close()
}

func TestPublishMessage(t *testing.T) {
	var count = 0
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message io.Reader) error {
				count++
				return nil
			},
		},
	}
	e := New()
	e.Submit(listeners...)
	e.Run()

	for i := 0; i < 10; i++ {
		e.Publish(SendString(fmt.Sprint(i)))
	}

	time.Sleep(1 * time.Second)
	assert.Equal(t, 10, count)
}

func TestCapacitySubscriber(t *testing.T) {
	var messageCount = 10
	listeners := []EventpoolListener{
		EventpoolListener{
			Name: "test1",
			Subscriber: func(name string, message io.Reader) error {
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
			Subscriber: func(name string, message io.Reader) error {
				return nil
			},
		},
		EventpoolListener{
			Name: "test2",
			Subscriber: func(name string, message io.Reader) error {
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
