[![main](https://github.com/flowerinthenight/clockbound-client-go/actions/workflows/main.yml/badge.svg)](https://github.com/flowerinthenight/clockbound-client-go/actions/workflows/main.yml)

## clockbound-client-go

A Go client for [AWS ClockBound](https://github.com/aws/clock-bound) using the newer ["ClockBound Shared Memory Protocol Version 1"](https://github.com/aws/clock-bound/blob/main/docs/PROTOCOL.md#clockbound-shared-memory-protocol-version-1). Only tested on Linux on little endian CPU architecture.

The [ClockBound daemon](https://github.com/aws/clock-bound/tree/main/clock-bound-d) must be running in order to use this library.

Check out the provided [example](./example/main.go) code on how to use the client.
