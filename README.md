# Redis gRPC Cache Service

A high-performance gRPC-based cache service backed by Valkey (Redis-compatible), written in Go.

## Features

- **gRPC API**: Fast, efficient protocol buffer-based communication
- **Valkey Backend**: Uses Valkey as the underlying cache storage
- **Key Operations**: Set, Get, Delete, and Exists operations
- **TTL Support**: Optional time-to-live for cached entries
- **Graceful Shutdown**: Proper signal handling for clean termination

## Prerequisites

- Go 1.25.1 or higher
- Valkey or Redis instance running
- Protocol Buffers compiler (for development)

## Installation

```bash
git clone https://github.com/Ruggilock/redis.git
cd redis
go mod download
```

## Configuration

The service is configured using environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VALKEY_HOST` | No | `localhost` | Valkey server host |
| `VALKEY_PORT` | No | `6379` | Valkey server port |
| `VALKEY_PASSWORD` | **Yes** | - | Valkey authentication password |

## Running the Service

```bash
export VALKEY_PASSWORD="your-password"
export VALKEY_HOST="localhost"
export VALKEY_PORT="6379"

go run cmd/server/main.go
```

The gRPC server will start on port `50051`.

## API Reference

### Set
Store a key-value pair with optional TTL.

```protobuf
rpc Set(SetRequest) returns (SetResponse);

message SetRequest {
  string key = 1;
  string value = 2;
  int64 ttl_seconds = 3; // 0 = no expiration
}
```

### Get
Retrieve a value by key.

```protobuf
rpc Get(GetRequest) returns (GetResponse);

message GetResponse {
  bool found = 1;
  string value = 2;
}
```

### Delete
Remove a key from the cache.

```protobuf
rpc Delete(DeleteRequest) returns (DeleteResponse);
```

### Exists
Check if a key exists in the cache.

```protobuf
rpc Exists(ExistsRequest) returns (ExistsResponse);
```

## Project Structure

```
.
cmd/
  server/           # Main server entry point
internal/
  service/          # Cache service and repository
proto/              # Protocol buffer definitions
  cache.proto
go.mod
```

## Development

### Regenerate Protocol Buffers

```bash
protoc --go_out=. --go-grpc_out=. proto/cache.proto
```

## License

See LICENSE file for details.
