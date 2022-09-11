# eventpool

Limiting Golang concurrency using the Queue and PubSub mechanism. Every function registered is needed to set the topic to deliver the message in the right pool.

## Installation

To install this package, you need to setup your Go workspace. The simplest way to install the library is to run:

```go
go get -u github.com/yudhasubki/eventpool
```

## Usage

```go

func main() {
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
}
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)