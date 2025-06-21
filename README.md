# Rocket State Service
This project implements a RESTful service in Go that consumes messages from various entities (rockets) and exposes the current state of these rockets via an API. The application has been designed with a focus on modularity, testability, and concurrency.

## The Challenge
The main goals of this service are:

1. Consume Messages: Receive JSON messages from multiple rocket "radio channels," where each channel represents a unique rocket.
2. Manage States: Update and maintain the current state of each rocket, considering that messages may arrive out of order and with an at-least-once guarantee (meaning there might be duplicates).
3. Expose REST API: Provide REST endpoints to query the state of individual rockets or a list of all rockets in the system.

## Architecture and Design
The application follows a clean (or hexagonal/layered) architecture organized into several packages to achieve clear separation of concerns, dependency injection, and facilitate testability.

- main: The application's entry point. It is responsible for initializing dependencies (repository, service, controller), configuring the HTTP router, and launching worker goroutines for asynchronous message processing.

- model: Defines the data structures representing incoming messages and the internal state of the rockets. It also contains the logic to update a rocket's state based on different message types.

- repository: Abstracts data storage logic. It contains the RocketRepository interface and its in-memory implementation. Thanks to dependency injection, this layer is easily replaceable with an implementation that uses a real database.

- service: Contains the core business logic of the application. It is responsible for processing messages (handling order and duplicates), interacting with the repository to persist and retrieve rocket states. It also includes the worker logic for asynchronous processing.

- controller: Manages HTTP requests and responses. It receives incoming messages via a POST endpoint and enqueues them for asynchronous processing. It also exposes GET endpoints to query rocket states.

### Using Concurrency (Goroutines and Channels)
The service has been refactored to use Go goroutines and channels to process rocket messages asynchronously.

- When a message arrives at the /messages endpoint, the ReceiveMessageHandler in the controller validates it and sends it to an internal channel (messageChannel).
- Several worker goroutines (launched from main and managed in the service package) consume messages from this channel concurrently. This allows the HTTP handler to respond quickly (202 Accepted), without waiting for the message to be fully processed.
- If the channel is full, the service responds with 503 Service Unavailable, indicating temporary overload.

## Technologies Used 🛠️
- Go (Golang): The primary programming language.
- Gin-Gonic: A high-performance web framework for Go, used to build the REST API.
- Swaggo: A tool to automatically generate Swagger/OpenAPI documentation from Go code annotations.
- Testify (Mock & Assert): A Go library to facilitate unit testing and assertions, as well as to create mocks for dependencies.

## How to Run the Application 🚀
### Prerequisites
Make sure you have the following installed:

- Go (version 1.16 or higher)
- The Swag CLI tool: go install github.com/swaggo/swag/cmd/swag@latest

Project Structure
Ensure your project has the following directory structure:
```
go-rocket-service/
├── main.go
├── controller/
│   ├── controller.go
│   └── controller_test.go
├── model/
│   ├── model.go
│   └── model_test.go
├── repository/
│   ├── repository.go
│   └── repository_test.go
└── service/
    ├── service.go
    └── service_test.go
```

### Steps to Run
1. Navigate to the project root directory:

```
cd go-rocket-service/
```

2. Initialize the Go module and download dependencies:
```
go mod tidy
```
3. Generate Swagger documentation:
```
swag init
```
4. Start the service:
```
make run
```

The service will start and listen on http://localhost:8088.

### Accessing Swagger Documentation
Once the service is running, you can access the interactive Swagger UI in your browser:
http://localhost:8088/swagger/index.html

## Makefile Automation
The project includes a Makefile to automate common tasks:

- make run: Generates Swagger documentation and then starts the application.

- make test: Executes all unit tests in the project (go test -v ./...).

- make coverage: Executes tests, generates a coverage report, and prints it to the console. It automatically excludes _mock.go files from coverage calculation.
```
make coverage
```
If you want to generate an HTML report for detailed coverage:
```
go tool cover -html=coverage.out -o coverage.html
```
- make swag: Explicitly generates Swagger documentation. Useful if you've modified annotations or endpoints.

- make clean: Deletes generated files from tests and Swagger (coverage reports, docs folder).

## Design Decisions and Trade-offs 📝
1. In-Memory Storage
- Decision: Use a Go in-memory generic map (map[string]T) to store rocket states.
- Advantages: Simple to implement, no external database configuration required, very high read/write speed (for a single node).
- Disadvantages: No data persistence. If the service restarts, all rocket states are lost. Does not scale horizontally (multiple service instances would not share the same state).
- Trade-off: Simplicity and speed over persistence and horizontal scalability.
- Alternatives (with trade-offs): Integrate a database (e.g., PostgreSQL, MongoDB, Redis) for persistence and scalability. This would add complexity in configuration, connection management, and ORM/drivers.

2. Out-of-Order and Duplicate Message Handling (ProcessMessage)
- Decision: Messages are processed based on their messageNumber. If an incoming messageNumber is greater than the last processed for that rocket, the state is updated. If it's equal, the message is re-processed (idempotency for duplicates). If it's smaller, the message is ignored.
- Advantages: Robust against "at-least-once" deliveries and out-of-order messages. The rocket's state always reflects the latest known information. Simple to implement and resource-efficient (no buffering of future messages required).
- Disadvantages: Does not guarantee strict sequential event processing. If message N+1 arrives before N and is processed, a subsequent N will be ignored even if it contains a valid earlier state change.
- Trade-off: Simplicity and "current state" consistency over a perfectly ordered event history.
- Alternatives (with trade-offs): Implement a per-rocket message reordering buffer that waits for missing messages (sequential messageNumber) before processing "ahead" messages. This would add memory complexity, timeout logic, and handling of unrecoverable "gaps." For strict order guarantees in complex event streams, distributed message queues like Apache Kafka would be used.

3. Asynchronous Concurrency (Goroutines and Channels)
- Decision: The MessageHandler enqueues messages into a channel, and a pool of worker goroutines processes them in the background.

- Advantages:
- - Non-Blocking: The main HTTP handler does not block waiting for message processing to complete, improving API responsiveness and throughput.
- - Resilience: Allows the service to absorb traffic spikes.
- - Scalability: Workers can be easily added or removed to adjust processing capacity.
- - Immediate Feedback: The client receives a 202 Accepted as soon as the message is enqueued.
- Disadvantages:
- - Complexity: Adds a layer of complexity to the design and debugging.
- - Observability: Requires monitoring of worker health and queue size to detect issues.
- - Eventual Consistency: The state queried by GET /rockets might not instantaneously reflect the latest POST message received, as processing is asynchronous.

- Trade-off: Higher throughput and API responsiveness over immediate "strong" consistency for all operations.

4. Use of Gin-Gonic
- Decision: Use Gin as the web framework.
- Advantages: Facilitates routing, JSON validation, middleware management, and writing concise handlers. Offers good performance.
- Disadvantages: Introduces a third-party dependency and a slight learning curve compared to the pure standard Go net/http package.
- Trade-off: Productivity and framework features over minimal dependencies.

5. Documentation with Swaggo
- Decision: Generate OpenAPI/Swagger documentation from code annotations.
- Advantages: Interactive documentation that is always up-to-date with the code. Facilitates API consumption by other developers.
- Disadvantages: Requires adding annotations to the code, which can be a bit verbose.
- Trade-off: Convenience of automated documentation over absolute code cleanliness.

6. Dependency Injection
Decision: Layers (controller, service, repository) are constructed by receiving interfaces as dependencies.

- Advantages:
- - Testability: Greatly facilitates unit testing by allowing the use of mocks to isolate layers.
- - Modularity: Layers are independent of the concrete implementations of their dependencies.
- - Flexibility: Allows easy swapping of an implementation (e.g., from an in-memory repository to a database) without affecting upper layers.
- - Disadvantages: Can add some "boilerplate" (repetitive code) during application initialization.
- - Trade-off: Maintainability and testability over initial conciseness.

## Automated Tests ✅
The project includes a comprehensive set of unit tests for the controller, service, model, and repository layers.

- The controller tests use a mock of the service and a message channel to verify that HTTP requests are handled correctly and that messages are enqueued as expected.
- The service tests use a mock of the repository to verify the business logic related to out-of-order and duplicate message processing, and state updates.
- The model tests verify the behavior of data structures and the rocket's state update logic.

- The repository tests verify save and retrieve operations in the in-memory implementation, ensuring concurrency safety.

These tests ensure code correctness and robustness, facilitating future modifications and regression detection.

This README.md should provide you with a clear overview of the application and the decisions made during its development.