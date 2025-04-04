package main

import (
	"fmt"

	"github.com/yudhasubki/eventpool"
)

func main() {
	eventPart := eventpool.NewPartition(3)
	listeners := []eventpool.EventpoolListener{
		{
			Name:       "groupA",
			Subscriber: SendMetrics,
		},
		{
			Name:       "groupB",
			Subscriber: SetCache,
		},
	}

	eventPart.Submit(listeners...)
	eventPart.Run()
}

func SendMetrics(name string, message []byte) error {
	panic("recover send metrics function")
}

func SetCache(name string, message []byte) error {

	fmt.Println(name, " receive message from publisher ", string(message))

	return nil
}

func SetWorkerOnFlight(name string, message []byte) error {

	fmt.Println(name, " receive message from publisher ", string(message))

	return nil
}
