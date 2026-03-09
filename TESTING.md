# Testing the Lowkey Signaling Handshake

This guide explains how to simulate a WebRTC signaling handshake between two peers using the `lowkey` CLI and gRPC server.

## Prerequisites

- **Redis**: Ensure Redis is running locally.
  ```bash
  docker run -p 6379:6379 redis
  ```

## Step 1: Start the Signaling Server

In your first terminal, build and run the gRPC server:

```bash
go build -o signaling-server ./server
./signaling-server
```

The server will listen on `[::]:50051`.

## Step 2: Start the Listening Peer (Terminal A)

In a second terminal, build the CLI and start the `listen` command. This peer will register a unique UUID and wait for signals.

```bash
go build -o lowkey .
./lowkey listen
```

**Note**: Copy the **My UUID** value printed in this terminal (e.g., `550e8400-e29b-41d4-a716-446655440000`).

## Step 3: Send a Signal (Terminal B)

In a third terminal, use the `send` command to fire a mock SDP Offer to the listener's UUID.

```bash
./lowkey send --target <SENDER_UUID_FROM_STEP_2>
```

## Expected Results

1. **Terminal B (Sender)**: Should print `SDP Offer sent successfully!`.
2. **Terminal A (Listener)**: Should display the `[SDP Received]` block containing the mock SDP data.
3. **Server Logs**: Should show `Client connected` and routing logs if implemented.

## Cleaning Up

Use `Ctrl+C` to stop the server and the listener.
