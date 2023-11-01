package eventpool

import "sync"

type Eventpool struct {
	mu      sync.Mutex
	workers map[string]map[string]*subscriber
}

type EventpoolListener struct {
	Name       string
	Subscriber SubscriberFunc
}

func New() *Eventpool {
	return &Eventpool{
		mu:      sync.Mutex{},
		workers: make(map[string]map[string]*subscriber),
	}
}

// Submit is receptionist to register topic and function to process message
func (w *Eventpool) Submit(topic string, eventpoolListeners []EventpoolListener, opts ...SubscriberConfigFunc) {
	w.workers[topic] = make(map[string]*subscriber)
	for _, listener := range eventpoolListeners {
		w.workers[topic][listener.Name] = newSubscriber(listener.Name, listener.Subscriber, opts...)
	}
}

// Publish is a mailman to publish message into the worker
func (w *Eventpool) Publish(topic string, message messageFunc) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, listener := range w.workers[topic] {
		msg, err := message()
		if err != nil {
			return err
		}

		listener.jobs <- msg
	}

	return nil
}

// Run is function for spawn worker to listen their jobs.
func (w *Eventpool) Run() {
	for _, worker := range w.workers {
		for _, listener := range worker {
			listener.listen()
		}
	}
}

// Cap is function get total message by topic name.
func (w *Eventpool) Cap(topic string, listenerName string) int {
	listener, exist := w.workers[topic][listenerName]
	if !exist {
		return 0
	}

	return listener.cap()
}

// Close is function to stop all the worker until the jobs get done.
func (w *Eventpool) Close() {
	for _, worker := range w.workers {
		for _, listener := range worker {
			listener.close()
		}
	}
}
