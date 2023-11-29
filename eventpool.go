package eventpool

import (
	"errors"
	"sync"
)

type Eventpool struct {
	workers *sync.Map
}

var (
	ErrorTopicNotFound = errors.New("topic not found")
)

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
func (w *Eventpool) Submit(topic string, eventpoolListeners ...EventpoolListener) {
	subscribers := make(map[string]*subscriber)
	for _, listener := range eventpoolListeners {
		subscribers[listener.Name] = newSubscriber(listener.Name, listener.Subscriber, listener.Opts...)
	}

	w.workers.Store(topic, subscribers)
}

// SubmitOnFlight is receptionist that always waiting to the new member while worker already running
func (w *Eventpool) SubmitOnFlight(topic string, eventpoolListeners ...EventpoolListener) {
	subscribers, ok := w.workers.Load(topic)
	if ok {
		for _, listener := range eventpoolListeners {
			if _, exist := subscribers.(map[string]*subscriber)[listener.Name]; !exist {
				sub := newSubscriber(listener.Name, listener.Subscriber, listener.Opts...)
				subscribers.(map[string]*subscriber)[listener.Name] = sub
				sub.listen()
			}
		}
		w.workers.Store(topic, subscribers)
		return
	}

	tempSubscribers := make(map[string]*subscriber)
	for _, listener := range eventpoolListeners {
		if _, exist := tempSubscribers[listener.Name]; !exist {
			sub := newSubscriber(listener.Name, listener.Subscriber, listener.Opts...)
			tempSubscribers[listener.Name] = sub
			sub.listen()
		}
	}

	w.workers.Store(topic, tempSubscribers)
}

// Publish is a mailman to publish message into the worker
func (w *Eventpool) Publish(topic string, message messageFunc) error {
	subscribers, ok := w.workers.Load(topic)
	if ok {
		for _, listener := range subscribers.(map[string]*subscriber) {
			msg, err := message()
			if err != nil {
				return err
			}

			listener.jobs <- msg
		}
	}

	return nil
}

// Run is function for spawn worker to listen their jobs.
func (w *Eventpool) Run() {
	w.workers.Range(func(key, value any) bool {
		for _, listener := range value.(map[string]*subscriber) {
			listener.listen()
		}
		return true
	})
}

// Cap is function get total message by topic name.
func (w *Eventpool) Cap(topic string, listenerName string) int {
	subscribers, ok := w.workers.Load(topic)
	if ok {
		listener, exist := subscribers.(map[string]*subscriber)[listenerName]
		if !exist {
			return 0
		}

		return listener.cap()
	}

	return 0
}

func (w *Eventpool) CloseBy(topic string) {
	w.workers.Delete(topic)
}

// Close is function to stop all the worker until the jobs get done.
func (w *Eventpool) Close() {
	w.workers.Range(func(key, value any) bool {
		for _, listener := range value.(map[string]*subscriber) {
			listener.close()
		}
		return true
	})
}
