[![main](https://github.com/flowerinthenight/clockbound-client-go/actions/workflows/main.yml/badge.svg)](https://github.com/flowerinthenight/clockbound-client-go/actions/workflows/main.yml)

## clockbound-client-go

A Go client for [AWS ClockBound](https://github.com/aws/clock-bound) using the newer [ClockBound Shared Memory Protocol Version 1](https://github.com/aws/clock-bound/blob/main/docs/PROTOCOL.md#clockbound-shared-memory-protocol-version-1) (ClockBound version 1.0.0 and later). Pre-1.0.0 uses the [ClockBound Unix Datagram Socket Protocol Version 1](https://github.com/aws/clock-bound/blob/main/docs/PROTOCOL.md#clockbound-unix-datagram-socket-protocol-version-1). Only tested on Linux on little endian CPU architecture.

The [ClockBound daemon](https://github.com/aws/clock-bound/tree/main/clock-bound-d) must be running in order to use this library.

Usage looks something like this:

```go
import (
  ...
  clockboundclient "github.com/flowerinthenight/clockbound-client-go"
)

func main() {
  // error checks redacted
  client, _ := clockboundclient.New()
  now, _ := client.Now()
  ...
  client.Close()
}
```

Check out the provided [example](./example/main.go) code for a more complete reference on how to use the client.
