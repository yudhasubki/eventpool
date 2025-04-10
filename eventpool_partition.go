package eventpool

import (
	"math/rand"
	"sync"

	"github.com/zeebo/xxh3"
)

type EventpoolPartition struct {
	Partitions    Partitions
	numPartitions int
}

type Partition struct {
	workers map[string]*PartitionedSubscriber
	mtx     sync.RWMutex
}

type Partitions []*Partition

func NewPartition(numPartitions int) *EventpoolPartition {
	if numPartitions <= 0 {
		numPartitions = 1
	}

	partitions := make(Partitions, numPartitions)
	for i := 0; i < numPartitions; i++ {
		partitions[i] = &Partition{
			workers: make(map[string]*PartitionedSubscriber),
			mtx:     sync.RWMutex{},
		}
	}

	return &EventpoolPartition{
		numPartitions: numPartitions,
		Partitions:    partitions,
	}
}

func (ep *EventpoolPartition) Submit(consumerPartition int, eventpoolListeners ...EventpoolListener) {
	for i := 0; i < ep.numPartitions; i++ {
		for _, listener := range eventpoolListeners {
			ep.Partitions[i].mtx.Lock()
			ep.Partitions[i].workers[listener.Name] = NewPartitionedSubscriber(listener.Name, listener.Subscriber, consumerPartition, listener.Opts...)
			ep.Partitions[i].mtx.Unlock()
		}
	}
}

// SubmitOnFlight is receptionist that always waiting to the new member while worker already running
func (ep *EventpoolPartition) SubmitOnFlight(consumerPartition int, eventpoolListeners ...EventpoolListener) {
	for i := 0; i < ep.numPartitions; i++ {
		for _, listener := range eventpoolListeners {
			ep.Partitions[i].mtx.Lock()

			_, exist := ep.Partitions[i].workers[listener.Name]
			if !exist {
				consumer := NewPartitionedSubscriber(listener.Name, listener.Subscriber, consumerPartition, listener.Opts...)
				ep.Partitions[i].workers[listener.Name] = consumer
			}

			ep.Partitions[i].mtx.Unlock()
		}
	}
}

func (ep *EventpoolPartition) CloseBy(listenerName ...string) {
	for _, listener := range listenerName {
		for index, partition := range ep.Partitions {
			partition.mtx.Lock()
			delete(ep.Partitions[index].workers, listener)
			partition.mtx.Unlock()
		}
	}
}

// Close is function to stop all the worker until the jobs get done.
func (ep *EventpoolPartition) Close() {
	for index := range ep.Partitions {
		for _, worker := range ep.Partitions[index].workers {
			worker.Close()
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
	// number partitions
	num := getPartition(key, ep.numPartitions)
	block := ep.Partitions[num]

	msg, err := message()
	if err != nil {
		return
	}

	if consumerGroupName == "" || consumerGroupName == "*" {
		for _, consumers := range block.workers {
			consumers.Submit(key, msg)
		}
		return
	}

	if consumers, ok := block.workers[consumerGroupName]; ok {
		consumers.Submit(key, msg)
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
			for _, worker := range workers.partitions {
				worker.listenPartition()
			}
		}
	}
}

// Cap is function get total message by topic name.
func (ep *EventpoolPartition) Cap(listenerName string) int {
	cap := 0
	for _, partition := range ep.Partitions {
		workers, exist := partition.workers[listenerName]
		if !exist {
			continue
		}
		for _, worker := range workers.partitions {
			cap += worker.cap()
		}
	}

	return cap
}
