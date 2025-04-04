## EventPool

This library provides a high-performance implementation of publish-subscribe pattern in Go with two distinct models:

- **Broadcast Type (Pub-Sub)** - Deliver messages to all subscribers
- **Partition Type (Queue)** - Distributed message processing with minimal contention

### Partition Type Feature
#### Smart Partitioning
- Uses XXH3 hash algorithm for consistent message partitioning
	- Automatic key-based partition assignment
	- Empty keys use random distribution
	- Hashed keys ensure consistent routing

#### Concurrency Optimized
- Partition-level isolation minimizes lock contention
- Each partition operates independently

## Features
- **Topic-Based Pub-Sub**: The library allows publishers to send messages to specific topics. Subscribers can then listen to those topics of interest and receive messages accordingly.
- **Flexible Communication**: Decouple your application's components by using the publish-subscribe pattern, promoting a more maintainable and scalable architecture.
- **Efficient and Lightweight**: Built on top of Golang channels, this library is highly performant, making it suitable for resource-constrained environments.
- **Maximum Retry Limit**: Define a maximum number of retry attempts for a specific operation or task. When this limit is reached, the error hook is triggered to handle the error gracefully.
- **Error Hooks**: Register custom error hook functions to implement tailored actions when an error occurs. This can include logging the error, sending notifications, triggering fallback mechanisms, or performing any other appropriate response.
- **Graceful Shutdown**: Implement a reliable and efficient shutdown process, allowing your application to complete ongoing tasks and clean up resources before terminating.
- **Close Hooks**: Register custom close hooks to execute specific cleanup tasks during the shutdown process. This ensures that essential operations are completed before the application exits.
- **Panic Recovery**: Put in place a mechanism to recover from panics and prevent your application from crashing.
- **Recover Hooks**: Register custom recover hooks to execute specific actions when a panic occurs. This allows you to log errors, perform cleanup tasks, or gracefully terminate the application.
- **Dead Letter Queue**: Integrate a Dead Letter Queue that receives messages that have failed to be processed by subscribers through the Error Hook.

## Installation

To use this library, make sure you have Go installed and set up a Go workspace.

Use go get to fetch the library:

```bash
go get -u github.com/yudhasubki/eventpool
```

## Usage
Here's a quick example of how to use the library:

### Event Partition
```go
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

	eventPart.Submit(3, listeners...)
	eventPart.Run()
}

func SendMetrics(name string, message []byte) error {
	panic("recover send metrics function")
}

func SetCache(name string, message []byte) error {
	fmt.Println(name, " receive message from publisher ", string(message))

	return nil
}
```

### Event Broadcast
```go
package main

import (
	"fmt"
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
				eventpool.RecoverHook(func(name string, job []byte) {

					fmt.Printf("[RecoverPanic][%s] message : %v \n", name, string(job))
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

	event.CloseBy(
		"send-metric",
		"set-cache",
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

func SendMetrics(name string, message []byte) error {
	panic("recover send metrics function")
}

func SetCache(name string, message []byte) error {

	fmt.Println(name, " receive message from publisher ", string(message))

	return nil
}
```

if you want to add a new listener while the application is already running just do it this simple way:

```go
event.SubmitOnFlight(eventpool.EventpoolListener{
	Name:       "set-in-the-air",
	Subscriber: SetWorkerInTheAir,
})
```

If you want to handle multiple topics, you can use a simple approach with a struct. For example:

```go
type PubSub struct {
	topics map[string]*eventpool.Eventpool
}
```

## 🚀 Benchmark Performance

### System Specification
```text
OS: darwin (macOS)  
Arch: arm64 (Apple M1)  
CPU: 8-core (4 performance + 4 efficiency)  
Go Version: 1.21+

BenchmarkEventSpecificGroupByPartition-8   6710184   221.5 ns/op   8 B/op   1 allocs/op
BenchmarkSingleEventByBroadcast-8          3252388   386.6 ns/op   8 B/op   1 allocs/op
BenchmarkEventWildcardByPartition-8        3077424   345.2 ns/op   8 B/op   1 allocs/op  
BenchmarkMultipleEventByBroadcast-8        2266489   457.7 ns/op   8 B/op   1 allocs/op
```

### 📊 Throughput Comparison

| Benchmark Mode                  | Operations/sec | Latency       | Memory | Allocs |
|---------------------------------|----------------|---------------|--------|--------|
| `SpecificGroupByPartition`      | **6,710,184**  | 221.5 ns/op   | 8 B    | 1      |
| `SingleEventBroadcast`          | 3,252,388      | 386.6 ns/op   | 8 B    | 1      |
| `WildcardByPartition`           | 3,077,424      | 345.2 ns/op   | 8 B    | 1      |
| `MultipleEventBroadcast`        | 2,266,489      | 457.7 ns/op   | 8 B    | 1      |


## Contributing
Contributions to this library are welcome! If you find any issues, have suggestions for improvements, or want to add new features, please submit a pull request or create an issue on the GitHub repository.

## License
[MIT](https://choosealicense.com/licenses/mit/)