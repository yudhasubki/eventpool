package eventpool

import (
	"math/rand"
	"sync"

	"github.com/zeebo/xxh3"
)

type EventpoolPartition struct {
	Partitions    Partitions
	numPartitions int
	mtx           sync.Mutex
}

type Partition struct {
	workers       map[string][]*subscriber
	numPartitions int
}

type Partitions []*Partition

func NewPartition(numPartitions int) *EventpoolPartition {
	if numPartitions <= 0 {
		numPartitions = 1
	}

	partitions := make(Partitions, numPartitions)
	for i := 0; i < numPartitions; i++ {
		partitions[i] = &Partition{
			workers:       make(map[string][]*subscriber),
			numPartitions: numPartitions,
		}
	}

	return &EventpoolPartition{
		numPartitions: numPartitions,
		Partitions:    partitions,
		mtx:           sync.Mutex{},
	}
}

func (ep *EventpoolPartition) Submit(consumerPartition int, eventpoolListeners ...EventpoolListener) {
	for _, listener := range eventpoolListeners {
		consumers := make([]*subscriber, consumerPartition)
		for j := 0; j < consumerPartition; j++ {
			consumers[j] = newSubscriber(listener.Name, listener.Subscriber, listener.Opts...)
		}

		for i := 0; i < ep.numPartitions; i++ {
			ep.mtx.Lock()
			ep.Partitions[i].workers[listener.Name] = consumers
			ep.mtx.Unlock()
		}
	}
}

// SubmitOnFlight is receptionist that always waiting to the new member while worker already running
func (ep *EventpoolPartition) SubmitOnFlight(consumerPartition int, eventpoolListeners ...EventpoolListener) {
	for _, listener := range eventpoolListeners {
		consumers := make([]*subscriber, consumerPartition)
		for j := 0; j < consumerPartition; j++ {
			consumers[j] = newSubscriber(listener.Name, listener.Subscriber, listener.Opts...)
		}

		for i := 0; i < ep.numPartitions; i++ {
			ep.mtx.Lock()
			_, exist := ep.Partitions[i].workers[listener.Name]
			if !exist {
				ep.Partitions[i].workers[listener.Name] = consumers

				for _, consumer := range ep.Partitions[i].workers[listener.Name] {
					consumer.listenPartition()
				}
			}
			ep.mtx.Unlock()
		}
	}
}

func (ep *EventpoolPartition) CloseBy(listenerName ...string) {
	ep.mtx.Lock()
	defer ep.mtx.Unlock()

	for _, listener := range listenerName {
		for index := range ep.Partitions {
			delete(ep.Partitions[index].workers, listener)
		}
	}
}

// Close is function to stop all the worker until the jobs get done.
func (ep *EventpoolPartition) Close() {
	for index := range ep.Partitions {
		for _, workers := range ep.Partitions[index].workers {
			for _, worker := range workers {
				worker.close()
			}
		}
	}
}

func (ep *EventpoolPartition) Subscribers() []string {
	subscribers := make([]string, 0)
	for _, partition := range ep.Partitions {
		for name := range partition.workers {
			subscribers = append(subscribers, name)
		}
		break
	}

	return subscribers
}

func (ep *EventpoolPartition) Publish(consumerGroupName string, key string, message messageFunc) {
	num := getPartition(key, ep.numPartitions)

	block := ep.Partitions[num]

	blockNum := getPartition(key, block.numPartitions)

	msg, err := message()
	if err != nil {
		return
	}

	if consumerGroupName == "" || consumerGroupName == "*" {
		for _, consumers := range block.workers {
			consumers[blockNum].jobs <- msg
		}
		return
	}

	if consumers, ok := block.workers[consumerGroupName]; ok {
		consumers[blockNum].jobs <- msg
	}
}

func getPartition(key string, numPartition int) int {
	if key == "" {
		return int(rand.Uint64() % uint64(numPartition))
	}

	return int(xxh3.HashString(key) % uint64(numPartition))
}

func (ep *EventpoolPartition) Run() {
	for _, partition := range ep.Partitions {
		for _, workers := range partition.workers {
			for _, worker := range workers {
				worker.listenPartition()
			}
		}
	}
}
