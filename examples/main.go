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
		eventpool.EventpoolListener{
			Name:       "set-log",
			Subscriber: SetLog,
		},
	)
	event.Submit("cart-delete",
		eventpool.EventpoolListener{
			Name:       "cart-delete",
			Subscriber: CartDelete,
		},
		eventpool.EventpoolListener{
			Name:       "cart-delete-counter",
			Subscriber: CartDeleteCounter,
		},
	)

	event.Run()

	for i := 0; i < 10; i++ {
		go event.Publish("order", eventpool.SendString(fmt.Sprintf("Order ID [%d] Received ", i)))
		go event.Publish("cart-delete", eventpool.SendString(fmt.Sprintf("Order ID [%d] Received ", i)))
	}
	time.Sleep(5 * time.Second)

	event.SubmitOnFlight("order-in-the-air", eventpool.EventpoolListener{
		Name:       "set-in-the-air",
		Subscriber: SetWorkerInTheAir,
	})

	event.SubmitOnFlight("order-in-the-air", eventpool.EventpoolListener{
		Name:       "set-in-the-air-2",
		Subscriber: SetWorkerInTheAir,
	})

	go event.CloseBy("order")
	for i := 0; i < 10; i++ {
		go func() {
			err := event.Publish("order", eventpool.SendString(fmt.Sprintf("Order ID [%d] Received ", i)))
			if err != nil {
				fmt.Println(err)
			}
		}()
		go event.Publish("order-in-the-air", eventpool.SendString(fmt.Sprintf("Order ID [%d] Received ", i)))
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

func SetLog(name string, message io.Reader) error {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, message)
	if err != nil {
		return err
	}

	fmt.Println(name, " receive message from publisher ", buf.String())

	return nil
}

func SetWorkerInTheAir(name string, message io.Reader) error {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, message)
	if err != nil {
		return err
	}

	fmt.Println(name, " receive message from publisher ", buf.String())

	return nil
}

func CartDelete(name string, message io.Reader) error {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, message)
	if err != nil {
		return err
	}

	fmt.Println(name, " receive message from publisher ", buf.String())

	return nil
}

func CartDeleteCounter(name string, message io.Reader) error {
	var buf bytes.Buffer

	_, err := io.Copy(&buf, message)
	if err != nil {
		return err
	}

	fmt.Println(name, " receive message from publisher ", buf.String())

	return nil
}
