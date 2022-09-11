package eventpool_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/yudhasubki/eventpool"
)

func TestPublisher(t *testing.T) {
	eventPool := eventpool.New()
	eventPool.SubscriberRegistry("test", func(id string, message io.Reader) error {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, message)
		if err != nil {
			return err
		}
		defer buf.Reset()

		fmt.Println(id, " ", buf.String())
		time.Sleep(3 * time.Second)

		return nil
	}, eventpool.SetMinFlight(2),
		eventpool.SetMaxFlight(10),
		eventpool.SetPoolThreshold(2),
		eventpool.SetThreshold(40))
	eventPool.Start()

	for i := 0; i < 10; i++ {
		eventPool.Publish("test", eventpool.SendString("Hai 1"))
		eventPool.Publish("test", eventpool.SendString("Hai 2"))
		eventPool.Publish("test", eventpool.SendString("Hai 3"))
		eventPool.Publish("test", eventpool.SendString("Hai 4"))
		eventPool.Publish("test", eventpool.SendString("Hai 5"))
		eventPool.Publish("test", eventpool.SendString("Hai 6"))
	}

	time.Sleep(15 * time.Second)
	eventPool.Publish("test", eventpool.SendString("Hai 7"))
	eventPool.Publish("test", eventpool.SendString("Hai 8"))
	eventPool.Publish("test", eventpool.SendString("Hai 9"))
	eventPool.Publish("test", eventpool.SendString("Hai 10"))
	eventPool.Publish("test", eventpool.SendString("Hai 11"))
	eventPool.Publish("test", eventpool.SendString("Hai 12"))
	time.Sleep(20 * time.Second)
}
