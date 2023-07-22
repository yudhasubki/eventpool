package eventpool

import (
	"context"
	"io"
	"time"
)

const (
	waitSleepClose = 1
)

type SubscriberFunc func(message io.Reader) error

type SubscriberConfigFunc func(c *subscriberConfig)

type subscriberConfig struct {
	bufferSize  int
	maxWorkers  int
	maxRetry    int
	errorHook   func(job io.Reader)
	recoverHook func()
	closeHook   func()
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
func RecoverHook(recoverHook func()) func(config *subscriberConfig) {
	return func(config *subscriberConfig) {
		config.recoverHook = recoverHook
	}
}

// CloseHook handling for close the eventpool
func CloseHook(closeHook func()) func(config *subscriberConfig) {
	return func(config *subscriberConfig) {
		config.closeHook = closeHook
	}
}

// ErrorHook handling for dead-letter queue
func ErrorHook(errorHook func(job io.Reader)) func(config *subscriberConfig) {
	return func(config *subscriberConfig) {
		config.errorHook = errorHook
	}
}

type subscriber struct {
	jobs   chan io.Reader
	fn     SubscriberFunc
	ctx    context.Context
	cancel context.CancelFunc
	config subscriberConfig
}

func newSubscriber(fn SubscriberFunc, opts ...SubscriberConfigFunc) *subscriber {
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
		jobs:   make(chan io.Reader, cfg.bufferSize),
		fn:     fn,
		ctx:    ctx,
		cancel: cancel,
		config: cfg,
	}
}

func (s *subscriber) listen() {
	for i := 0; i < s.config.maxWorkers; i++ {
		go s.spawn(i)
	}
}

func (s *subscriber) spawn(worker int) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case job := <-s.jobs:
			maxRetry := 0
		Retry:
			err := s.process(job) // create new func to handle panic recover
			if err != nil {
				maxRetry++
				if maxRetry <= s.config.maxRetry {
					goto Retry
				}

				if maxRetry > s.config.maxRetry && s.config.errorHook != nil {
					s.config.errorHook(job)
				}
			}
		}
	}
}

func (s *subscriber) process(job io.Reader) error {
	defer func() {
		if r := recover(); r != nil {
			if s.config.recoverHook != nil {
				s.config.recoverHook()
			}
		}
	}()

	err := s.fn(job)
	if err != nil {
		return err
	}
	return nil
}

func (s *subscriber) cap() int {
	return len(s.jobs)
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

	s.config.closeHook()
}
