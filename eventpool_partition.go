package eventpool

import (
	"math/rand"

	"github.com/zeebo/xxh3"
)

type EventpoolPartition struct {
	Partitions    Partitions
	numPartitions int
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
	}
}

func (ep *EventpoolPartition) Submit(consumerPartition int, eventpoolListeners ...EventpoolListener) {
	for _, listener := range eventpoolListeners {
		consumers := make([]*subscriber, consumerPartition)
		for j := 0; j < consumerPartition; j++ {
			consumers[j] = newSubscriber(listener.Name, listener.Subscriber, listener.Opts...)
		}

		for i := 0; i < ep.numPartitions; i++ {
			ep.Partitions[i].workers[listener.Name] = consumers
		}
	}
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
