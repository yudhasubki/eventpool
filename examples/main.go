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
	event.Submit("order",
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
				}), // if needed
				eventpool.CloseHook(func(name string) {
					fmt.Printf("[Enter Gracefully Shutdown][%s]\n", name)
				}), // if needed
			},
		},
		eventpool.EventpoolListener{
			Name:       "set-cache",
			Subscriber: SetCache,
		},
		eventpool.EventpoolListener{
			Name:       "set-log",
			Subscriber: SetLog,
		},
	)

	event.Run()

	for i := 0; i < 10; i++ {
		event.Publish("order", eventpool.SendString(fmt.Sprintf("Order ID [%d] Received ", i)))
	}

	time.Sleep(5 * time.Second)
	event.Close()
	time.Sleep(5 * time.Second)
}

func SendMetrics(message io.Reader) error {
	panic("recover send metrics function")
}

func SetCache(message io.Reader) error {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, message)
	if err != nil {
		return err
	}

	fmt.Println("[SetCache] receive message from publisher ", buf.String())

	return nil
}

func SetLog(message io.Reader) error {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, message)
	if err != nil {
		return err
	}

	fmt.Println("[SetLog] receive message from publisher ", buf.String())

	return nil
}
