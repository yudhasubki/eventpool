package eventpool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	waitSleepClose = 1
)

type SubscriberFunc func(name string, message []byte) error

type SubscriberConfigFunc func(c *subscriberConfig)

type subscriberConfig struct {
	bufferSize  int
	maxWorkers  int
	maxRetry    int
	errorHook   func(name string, job []byte)
	recoverHook func(name string, job []byte)
	closeHook   func(name string)
}

func BufferSize(bufferSize int) func(config *subscriberConfig) {
	return func(config *subscriberConfig) {
		if bufferSize == 0 {
			return
		}

		config.bufferSize = bufferSize
	}
}

func MaxWorker(max int) func(config *subscriberConfig) {
	return func(config *subscriberConfig) {
		if max == 0 {
			return
		}

		config.maxWorkers = max
	}
}

func MaxRetry(max int) func(config *subscriberConfig) {
	return func(config *subscriberConfig) {
		if max == 0 {
			return
		}
		config.maxRetry = max
	}
}

// RecoverHook handling if receive the signal panic
func RecoverHook(recoverHook func(name string, job []byte)) func(config *subscriberConfig) {
	return func(config *subscriberConfig) {
		config.recoverHook = recoverHook
	}
}

// CloseHook handling for close the eventpool
func CloseHook(closeHook func(name string)) func(config *subscriberConfig) {
	return func(config *subscriberConfig) {
		config.closeHook = closeHook
	}
}

// ErrorHook handling for dead-letter queue
func ErrorHook(errorHook func(name string, job []byte)) func(config *subscriberConfig) {
	return func(config *subscriberConfig) {
		config.errorHook = errorHook
	}
}

type subscriber struct {
	name              string
	jobs              chan []byte
	fn                SubscriberFunc
	ctx               context.Context
	cancel            context.CancelFunc
	config            subscriberConfig
	messageProcessing int
	mtx               sync.RWMutex
}

func newSubscriber(name string, fn SubscriberFunc, opts ...SubscriberConfigFunc) *subscriber {
	cfg := subscriberConfig{
		maxWorkers:  10,
		bufferSize:  100,
		maxRetry:    3,
		closeHook:   nil,
		recoverHook: nil,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &subscriber{
		name:              name,
		jobs:              make(chan []byte, cfg.bufferSize),
		fn:                fn,
		ctx:               ctx,
		cancel:            cancel,
		config:            cfg,
		messageProcessing: 0,
		mtx:               sync.RWMutex{},
	}
}

func (s *subscriber) listenPartition() {
	go s.spawn()
}

func (s *subscriber) listen() {
	for i := 0; i < s.config.maxWorkers; i++ {
		go s.spawn()
	}
}

func (s *subscriber) spawn() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case job := <-s.jobs:
			s.mtx.Lock()
			s.messageProcessing++
			s.mtx.Unlock()

			maxRetry := 0
		Retry:
			err := s.process(s.name, job) // create new func to handle panic recover
			if err != nil {
				maxRetry++
				if maxRetry <= s.config.maxRetry {
					goto Retry
				}

				if maxRetry > s.config.maxRetry && s.config.errorHook != nil {
					s.config.errorHook(s.name, job)
				}
			}

			s.mtx.Lock()
			s.messageProcessing--
			s.mtx.Unlock()
		}
	}
}

func (s *subscriber) process(name string, job []byte) error {
	defer func() {
		if r := recover(); r != nil {
			if s.config.recoverHook != nil {
				s.config.recoverHook(s.name, job)
			}
		}
	}()

	return s.fn(name, job)
}

func (s *subscriber) cap() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.messageProcessing
}

func (s *subscriber) close() {
Retry:
	if s.cap() > 0 {
		time.Sleep(waitSleepClose * time.Second)
		goto Retry
	}

	// cancel until the job getting done
	s.cancel()
	if s.config.closeHook == nil {
		return
	}

	s.config.closeHook(s.name)
}

type PartitionedSubscriber struct {
	partitions    []*subscriber
	numPartitions int
}

func NewPartitionedSubscriber(
	name string,
	handler SubscriberFunc,
	numPartitions int,
	opts ...SubscriberConfigFunc,
) *PartitionedSubscriber {
	partitions := make([]*subscriber, numPartitions)
	for i := 0; i < numPartitions; i++ {
		sub := newSubscriber(fmt.Sprintf("%s-%d", name, i), handler, opts...)
		partitions[i] = sub
	}

	return &PartitionedSubscriber{
		partitions:    partitions,
		numPartitions: numPartitions,
	}
}

func (ps *PartitionedSubscriber) Submit(key string, data []byte) {
	index := getPartition(key, ps.numPartitions)
	ps.partitions[index].jobs <- data
}

func (ps *PartitionedSubscriber) Close() {
	for _, sub := range ps.partitions {
		sub.close()
	}
}
