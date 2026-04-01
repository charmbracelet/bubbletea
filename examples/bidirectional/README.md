# Bidirectional Goroutine Communication

This example demonstrates two-way communication between a Bubble Tea program
and a background goroutine.

The TUI sends questions (requests) to a background "oracle" worker via a
channel. The worker processes the request and sends a result back on a separate
channel. The TUI uses a Cmd to wait for the result without blocking the UI.

## Key patterns

- **Request channel**: The model holds a `chan<- request` to send work to the
  background goroutine.
- **Result channel**: The model holds a `<-chan resultMsg` to receive results.
- **waitForResult Cmd**: A command that blocks on the result channel. It is
  returned from Update only when a request is in flight, ensuring the program
  listens for the response.
- **Typed messages**: Both the request and result carry actual data (query
  string, answer, duration) rather than empty structs.

## Running

```bash
go run .
```

Type a question, press Enter, and the oracle will answer.
