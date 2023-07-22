package eventpool

import "sync"

type Eventpool struct {
	mu      sync.Mutex
	workers map[string]*subscriber
}

func New() *Eventpool {
	return &Eventpool{
		mu:      sync.Mutex{},
		workers: make(map[string]*subscriber),
	}
}

// Submit is receptionist to register topic and function to process message
func (w *Eventpool) Submit(topic string, fn SubscriberFunc, opts ...SubscriberConfigFunc) {
	w.workers[topic] = newSubscriber(fn, opts...)
}

// Publish is a mailman to publish message into the worker
func (w *Eventpool) Publish(topic string, message messageFunc) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	msg, err := message()
	if err != nil {
		return err
	}
	w.workers[topic].jobs <- msg

	return nil
}

// Run is function for spawn worker to listen their jobs.
func (w *Eventpool) Run() {
	for _, worker := range w.workers {
		worker.listen()
	}
}

// Cap is function get total message by topic name.
func (w *Eventpool) Cap(topic string) int {
	return w.workers[topic].cap()
}

// Close is function to stop all the worker until the jobs get done.
func (w *Eventpool) Close() {
	for _, worker := range w.workers {
		worker.close()
	}
}
