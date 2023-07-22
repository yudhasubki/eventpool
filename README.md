# Eventpool

This library provides a simple and efficient implementation of a local publish-subscribe pattern with topics and queues using Golang channels. The publish-subscribe pattern is a widely used messaging paradigm that enables decoupling of components in an application. With this library, you can easily build event-driven systems and enable communication between different parts of your application without the components directly knowing about each other.

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

```go

func main() {
	eventPool := eventpool.New()
	eventPool.Submit("your_topic", exampleConsumer)
	eventPool.Run()

	for i := 0; i < 10; i++ {
		eventPool.Publish("test", eventpool.SendString("Hai 1"))
		eventPool.Publish("test", eventpool.SendString("Hai 2"))
		eventPool.Publish("test", eventpool.SendString("Hai 3"))
		eventPool.Publish("test", eventpool.SendString("Hai 4"))
		eventPool.Publish("test", eventpool.SendString("Hai 5"))
		eventPool.Publish("test", eventpool.SendString("Hai 6"))
	}
}

func exampleConsumer(message io.Reader) error {
	fmt.Println("PING")

	return nil
}
```

## Contributing
Contributions to this library are welcome! If you find any issues, have suggestions for improvements, or want to add new features, please submit a pull request or create an issue on the GitHub repository.

## License
[MIT](https://choosealicense.com/licenses/mit/)