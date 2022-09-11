package eventpool

import (
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SubscriberFunc func(id string, message io.Reader) error

type Subscriber struct {
	topic     string
	pools     SubscriberPools
	message   chan io.Reader
	fn        SubscriberFunc
	config    SubscribeConfig
	counter   float64
	counterCh chan float64
	mtx       *sync.RWMutex
}

func NewSubscriber(fn SubscriberFunc, topic string, opts ...SubscribeConfigFunc) *Subscriber {
	cfg := NewSubscriberConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	return &Subscriber{
		topic:     topic,
		message:   make(chan io.Reader),
		pools:     make(SubscriberPools, 0),
		fn:        fn,
		counterCh: make(chan float64),
		config:    *cfg,
		mtx:       new(sync.RWMutex),
	}
}

func (s *Subscriber) Dispatcher() {
	for i := 0; i < s.config.minFlight; i++ {
		s.Spawn()
	}

	// if counter > threshold and has space to spawn worker
	// spawner will spawn the worker
	go s.spawner()

	for {
		tick := time.NewTicker(3 * time.Second)
		select {
		case <-tick.C:
			s.cleanupResource()
		case msg := <-s.message:
			s.message <- msg
		}
	}
}

func (s *Subscriber) Increment(c float64) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.counter += c
}

func (s *Subscriber) Size() float64 {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.counter
}

func (s *Subscriber) spawner() {
	// counter is used for checking the threshold
	go func() {
		for c := range s.counterCh {
			s.Increment(c)
		}
	}()

	for {
		avg := (s.Size() / s.pools.Threshold()) * 100
		if avg >= s.config.threshold && len(s.pools) < s.config.maxFlight {
			s.Spawn()
		}
		time.Sleep(2 * time.Second)
	}
}

func (s *Subscriber) Spawn() {
	id := uuid.New().String()

	pool := &SubscriberPool{
		id:         id,
		message:    s.message,
		processor:  make(chan io.Reader),
		clean:      make(chan bool),
		fn:         s.fn,
		poolBuffer: s.config.poolBuffer,
		counterCh:  s.counterCh,
		mtx:        new(sync.RWMutex),
	}
	s.pools = append(s.pools, pool)

	go pool.Pool()
}

// cleanup is function to remove worker when idle reach
// the limit and the are no message to be processed.
// purpose to reduce the spawner never close / die
func (s *Subscriber) cleanupResource() {
	for {
		removed := false
		for i, pool := range s.pools {
			if pool.counter == 0 && len(s.pools) > s.config.minFlight && pool.idle > 2 {
				s.pools[i].clean <- true
				s.pools = append(s.pools[:i], s.pools[i+1:]...)
				removed = true
				break
			}
		}

		if !removed {
			break
		}
	}
}

type SubscriberPool struct {
	id         string
	idle       int
	counter    float64
	poolBuffer float64
	mtx        *sync.RWMutex
	clean      chan bool
	counterCh  chan float64
	message    chan io.Reader
	processor  chan io.Reader
	fn         SubscriberFunc
}

// pool is function to accommodate message from dispatcher
// to consuming it and several function to turn off the worker
// or how much the worker idling
func (s *SubscriberPool) Pool() {
	go func() {
		for {
			if s.Size() == s.poolBuffer {
				continue
			}

			ticker := time.NewTicker(3 * time.Second)

			select {
			case <-s.clean:
				return
			case <-ticker.C:
				s.idle++
			case msg := <-s.message:
				if s.idle > 0 {
					s.idle = 0
				}

				s.processor <- msg
			}
		}
	}()

	for message := range s.processor {
		s.increment()

		// run the process
		s.fn(s.id, message)

		s.decrement()
	}
	close(s.processor)
}

func (s *SubscriberPool) Size() float64 {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.counter
}

func (s *SubscriberPool) decrement() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.counter--
	s.counterCh <- -1
}

func (s *SubscriberPool) increment() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.counter++
	s.counterCh <- 1
}

type SubscriberPools []*SubscriberPool

func (ss SubscriberPools) Threshold() float64 {
	var total float64 = 0
	for _, s := range ss {
		total += s.poolBuffer
	}

	return total
}

type SubscribeConfig struct {
	minFlight  int
	maxFlight  int
	threshold  float64
	poolBuffer float64
}

func NewSubscriberConfig() *SubscribeConfig {
	return &SubscribeConfig{
		minFlight:  2,
		maxFlight:  3,
		threshold:  60,
		poolBuffer: 1,
	}
}

type SubscribeConfigFunc func(*SubscribeConfig)

// SetMinFlight is function to set minimum workers must be spawned
func SetMinFlight(min int) SubscribeConfigFunc {
	return func(sc *SubscribeConfig) {
		if min < 1 {
			min = 2
		}
		sc.minFlight = min
	}
}

// SetMinFlight is function to set maximum workers could be spawned
func SetMaxFlight(max int) SubscribeConfigFunc {
	return func(sc *SubscribeConfig) {
		sc.maxFlight = max
	}
}

// SetThreshold is function to set minimum threshold value to spawn the worker
// if reach the threshold worker will be spawned
func SetThreshold(threshold float64) SubscribeConfigFunc {
	return func(sc *SubscribeConfig) {
		sc.threshold = threshold
	}
}

// SetPoolThreshold is function to set value on each worker to accept the message
// if worker is fully of message, the worker will routing to another worker
// its similar like distributed message
func SetPoolThreshold(poolBuffer float64) SubscribeConfigFunc {
	return func(sc *SubscribeConfig) {
		sc.poolBuffer = poolBuffer
	}
}
