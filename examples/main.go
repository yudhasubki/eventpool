package main

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/yudhasubki/eventpool"
)

func main() {
	event := eventpool.New()
	event.Submit(
		eventpool.EventpoolListener{
			Name:       "send-metric",
			Subscriber: SendMetrics,
			Opts: []eventpool.SubscriberConfigFunc{
				eventpool.RecoverHook(func(name string, job io.Reader) {
					var buf bytes.Buffer

					_, err := io.Copy(&buf, job)
					if err != nil {
						return
					}

					fmt.Printf("[RecoverPanic][%s] message : %v \n", name, buf.String())
				}),
				eventpool.CloseHook(func(name string) {
					fmt.Printf("[Enter Gracefully Shutdown][%s]\n", name)
				}),
			},
		},
		eventpool.EventpoolListener{
			Name:       "set-cache",
			Subscriber: SetCache,
		},
	)
	event.Run()

	for i := 0; i < 10; i++ {
		go event.Publish(eventpool.SendString(fmt.Sprintf("Order ID [%d] Received ", i)))
	}
	time.Sleep(5 * time.Second)

	event.SubmitOnFlight(eventpool.EventpoolListener{
		Name:       "set-in-the-air",
		Subscriber: SetWorkerOnFlight,
	})

	event.CloseBy(
		"send-metric",
		"set-cache",
		"set-in-the-air",
		"set-in-the-air-2",
		"cart-delete-counter",
		"cart-delete",
	)

	for i := 0; i < 10; i++ {
		go func(i int) {
			event.Publish(eventpool.SendString(fmt.Sprintf("Order ID [%d] Received ", i)))
		}(i)
	}

	time.Sleep(5 * time.Second)
	event.Close()
	time.Sleep(5 * time.Second)
}

func SendMetrics(name string, message io.Reader) error {
	panic("recover send metrics function")
}

func SetCache(name string, message io.Reader) error {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, message)
	if err != nil {
		return err
	}

	fmt.Println(name, " receive message from publisher ", buf.String())

	return nil
}

func SetWorkerOnFlight(name string, message io.Reader) error {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, message)
	if err != nil {
		return err
	}

	fmt.Println(name, " receive message from publisher ", buf.String())

	return nil
}
