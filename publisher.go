package eventpool

import (
	"sync"
)

type EventPool struct {
	subscribers map[string]*Subscriber
	mtx         map[string]*sync.Mutex
}

func (e *EventPool) Publish(topic string, message MessageFunc) error {
	e.mtx[topic].Lock()
	defer e.mtx[topic].Unlock()

	data, err := message()
	if err != nil {
		return err
	}

	e.subscribers[topic].message <- data

	return nil
}

func (e *EventPool) SubscriberRegistry(topic string, fn SubscriberFunc, opts ...SubscribeConfigFunc) {
	e.subscribers[topic] = NewSubscriber(fn, topic, opts...)
	e.mtx[topic] = new(sync.Mutex)
}

func New() *EventPool {
	return &EventPool{
		subscribers: make(map[string]*Subscriber),
		mtx:         make(map[string]*sync.Mutex),
	}
}

func (e EventPool) Start() {
	for _, s := range e.subscribers {
		go s.Dispatcher()
	}
}
