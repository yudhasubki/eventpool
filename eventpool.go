package eventpool

import (
	"sync"
)

type Eventpool struct {
	workers *sync.Map
}

type EventpoolListener struct {
	Name       string
	Subscriber SubscriberFunc
	Opts       []SubscriberConfigFunc
}

func New() *Eventpool {
	return &Eventpool{
		workers: new(sync.Map),
	}
}

// Submit is receptionist to register topic and function to process message
func (w *Eventpool) Submit(eventpoolListeners ...EventpoolListener) {
	for _, listener := range eventpoolListeners {
		w.workers.Store(listener.Name, newSubscriber(listener.Name, listener.Subscriber, listener.Opts...))
	}
}

// SubmitOnFlight is receptionist that always waiting to the new member while worker already running
func (w *Eventpool) SubmitOnFlight(eventpoolListeners ...EventpoolListener) {
	for _, listener := range eventpoolListeners {
		sub, loaded := w.workers.LoadOrStore(listener.Name, newSubscriber(listener.Name, listener.Subscriber, listener.Opts...))
		if loaded {
			continue
		}
		sub.(*subscriber).listen()
	}
}

// Publish is a mailman to publish message into the worker
func (w *Eventpool) Publish(message messageFunc) {
	w.workers.Range(func(key, value any) bool {
		msg, err := message()
		if err != nil {
			return false
		}
		value.(*subscriber).jobs <- msg

		return true
	})
}

// Run is function for spawn worker to listen their jobs.
func (w *Eventpool) Run() {
	w.workers.Range(func(key, value any) bool {
		value.(*subscriber).listen()
		return true
	})
}

// Subscribers is function to get all listener name by topic name
func (w *Eventpool) Subscribers() []string {
	subscribers := make([]string, 0)
	w.workers.Range(func(key, value any) bool {
		subscribers = append(subscribers, value.(*subscriber).name)
		return true
	})

	return subscribers
}

// Cap is function get total message by topic name.
func (w *Eventpool) Cap(listenerName string) int {
	sub, ok := w.workers.Load(listenerName)
	if ok {
		return sub.(*subscriber).cap()
	}

	return 0
}

func (w *Eventpool) CloseBy(listenerName ...string) {
	for _, listener := range listenerName {
		subs, loaded := w.workers.LoadAndDelete(listener)
		if loaded {
			subs.(*subscriber).close()
		}
	}
}

// Close is function to stop all the worker until the jobs get done.
func (w *Eventpool) Close() {
	w.workers.Range(func(key, value any) bool {
		value.(*subscriber).close()
		return true
	})
}
