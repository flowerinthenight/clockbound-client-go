**NOTE: Unusable at the moment. Still under investigation but it looks like the As-Of-Timestamp value read from SHM indicates elapsed time from boot, not from Jan 1 1970 UTC (Unix epoch). Could be the way SHM is being read.**

In the meantime, if you can use CGO, have a look at [clockbound-ffi-go](https://github.com/flowerinthenight/clockbound-ffi-go).

---

[![main](https://github.com/flowerinthenight/clockbound-client-go/actions/workflows/main.yml/badge.svg)](https://github.com/flowerinthenight/clockbound-client-go/actions/workflows/main.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/flowerinthenight/clockbound-client-go.svg)](https://pkg.go.dev/github.com/flowerinthenight/clockbound-client-go)

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

A cloud-init [startup script](./startup-aws-asg.sh) is also provided for spinning up an ASG with the ClockBound daemon already setup.

```sh
# Create a launch template. ImageId here is Amazon Linux, default VPC.
# (Added newlines for readability. Might not run when copied as is.)
$ aws ec2 create-launch-template \
  --launch-template-name cbclient-lt \
  --version-description version1 \
  --launch-template-data '
  {
    "UserData":"'"$(cat startup-aws-asg.sh | base64 -w 0)"'",
    "ImageId":"ami-0fb04413c9de69305",
    "InstanceType":"t2.micro"
  }'

# Create the ASG; update {target-zone} with actual value:
$ aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name cbclient-asg \
  --launch-template LaunchTemplateName=cbclient-lt,Version='1' \
  --min-size 1 \
  --max-size 1 \
  --availability-zones {target-zone}

# You can now SSH to the instance. Note that it might take some time before
# ClockBound is running due to the need to build it in Rust. You can wait
# for the `clockbound` process, or tail the startup script output, like so:
$ tail -f /var/log/cloud-init-output.log

# Run the sample code:
# Download the latest release sample from GitHub.
$ tar xvzf clockbound-client-sample-v{latest-version}-x86_64-linux.tar.gz
$ ./example
2025/01/23 03:03:59 earliest: 1970-01-01T00:04:03.189017214Z
2025/01/23 03:03:59 latest  : 1970-01-01T00:04:03.190960726Z
2025/01/23 03:03:59 range: 1.943512ms
2025/01/23 03:03:59 status: SYNCHRONIZED
...
```

## License

This library is licensed under the [MIT License](./LICENSE).
