package clockboundclient

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mmap "github.com/edsrzf/mmap-go"
)

const DefaultShmPath = "/var/run/clockbound/shm"

type ClockStatus int

const (
	ClockStatusUnknown ClockStatus = iota
	ClockStatusSynchronized
	ClockStatusFreeRunning
)

type Now struct {
	Earliest time.Time
	Latest   time.Time
	Status   ClockStatus
}

type Client struct {
	f   *os.File
	m   mmap.MMap
	err error
}

func (c *Client) Now() (Now, error) {
	log.Printf("len: %d\n", len(c.m))
	log.Printf("%X\n", c.m)

	size := binary.LittleEndian.Uint32(c.m[8:12])
	log.Printf("size: 0x%X %d\n", size, size)

	ver := binary.LittleEndian.Uint16(c.m[12:14])
	log.Printf("version: 0x%X %d\n", ver, ver)

	gen := binary.LittleEndian.Uint16(c.m[14:16])
	log.Printf("generation: 0x%X %d\n", gen, gen)

	// As-of-timestamp
	asof_s := binary.LittleEndian.Uint64(c.m[16:24])
	log.Printf("as-of-ts (s): 0x%X %d\n", asof_s, asof_s)
	asof_ns := binary.LittleEndian.Uint64(c.m[24:32])
	log.Printf("as-of-ts (ns): 0x%X %d\n", asof_ns, asof_ns)
	ts := time.Unix(int64(asof_s), int64(asof_ns))

	// Void-after-timestamp
	va_s := binary.LittleEndian.Uint64(c.m[32:40])
	log.Printf("void-after-ts (s): 0x%X %d\n", va_s, va_s)
	va_ns := binary.LittleEndian.Uint64(c.m[40:48])
	log.Printf("void-after-ts (ns): 0x%X %d\n", va_ns, va_ns)
	vts := time.Unix(int64(va_s), int64(va_ns))

	log.Printf("as_of_ts  : %v\n", ts.Format(time.RFC3339Nano))
	log.Printf("void_after: %v\n", vts.Format(time.RFC3339Nano))

	bound := binary.LittleEndian.Uint64(c.m[48:56])
	log.Printf("bound_ns: 0x%X %d\n", bound, bound)

	drift := binary.LittleEndian.Uint32(c.m[56:60])
	log.Printf("drift: 0x%X %d\n", drift, drift)

	reserved := binary.LittleEndian.Uint32(c.m[60:64])
	log.Printf("reserved: 0x%X %d\n", reserved, reserved)

	status := binary.LittleEndian.Uint32(c.m[64:68])
	log.Printf("clock_status: 0x%X %d\n", status, status)

	earliest := ts.Add(-1 * (time.Nanosecond * time.Duration(bound)))
	latest := ts.Add(time.Nanosecond * time.Duration(bound))

	unix_ns := latest.UnixNano() - (latest.UnixNano()-earliest.UnixNano())/2
	log.Printf("up : %v\n", latest)
	log.Printf("now: %v\n", fromUnixNano(uint64(unix_ns)))
	log.Printf("low: %v\n", earliest)
	log.Printf("range: %v\n", latest.Sub(earliest))

	return Now{
		Status: ClockStatus(status),
	}, nil
}

func (c *Client) Error() string {
	switch c.err {
	case nil:
		return fmt.Sprintf("%v", nil)
	default:
		return c.err.Error()
	}
}

func (c *Client) Close() error {
	if c.err != nil {
		return c.err
	}

	if err := c.m.Unmap(); err != nil {
		return fmt.Errorf("Unmap failed: %w", err)
	}

	if err := c.f.Close(); err != nil {
		return fmt.Errorf("Close failed: %w", err)
	}

	return nil
}

func New() (*Client, error) {
	c := &Client{}
	f, err := os.OpenFile(DefaultShmPath, os.O_RDONLY, 0755)
	if err != nil {
		c.err = fmt.Errorf("OpenFile failed: %w", err)
		return c, c.err
	}

	m, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		c.err = fmt.Errorf("Map failed: %w", err)
		return c, c.err
	}

	c.f = f
	c.m = m
	return c, nil
}

func main() {
	f, err := os.OpenFile("/var/run/clockbound/shm", os.O_RDONLY, 0755)
	if err != nil {
		log.Println("OpenFile failed:", err)
		return
	}

	defer f.Close()

	m, err := mmap.Map(f, mmap.RDONLY, 0)
	if err != nil {
		log.Println("Map failed:", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	ticker := time.NewTicker(time.Second * 3)

	go func() {
		for {
			select {
			case <-ctx.Done():
				done <- nil
				return
			case <-ticker.C:
			}

			log.Printf("len: %d\n", len(m))
			log.Printf("%X\n", m)

			size := binary.LittleEndian.Uint32(m[8:12])
			log.Printf("size: 0x%X %d\n", size, size)

			ver := binary.LittleEndian.Uint16(m[12:14])
			log.Printf("version: 0x%X %d\n", ver, ver)

			gen := binary.LittleEndian.Uint16(m[14:16])
			log.Printf("generation: 0x%X %d\n", gen, gen)

			// As-of-timestamp
			asof_s := binary.LittleEndian.Uint64(m[16:24])
			log.Printf("as-of-ts (s): 0x%X %d\n", asof_s, asof_s)
			asof_ns := binary.LittleEndian.Uint64(m[24:32])
			log.Printf("as-of-ts (ns): 0x%X %d\n", asof_ns, asof_ns)
			ts := time.Unix(int64(asof_s), int64(asof_ns))

			// Void-after-timestamp
			va_s := binary.LittleEndian.Uint64(m[32:40])
			log.Printf("void-after-ts (s): 0x%X %d\n", va_s, va_s)
			va_ns := binary.LittleEndian.Uint64(m[40:48])
			log.Printf("void-after-ts (ns): 0x%X %d\n", va_ns, va_ns)
			vts := time.Unix(int64(va_s), int64(va_ns))

			log.Printf("as_of_ts  : %v\n", ts.Format(time.RFC3339Nano))
			log.Printf("void_after: %v\n", vts.Format(time.RFC3339Nano))

			bound := binary.LittleEndian.Uint64(m[48:56])
			log.Printf("bound_ns: 0x%X %d\n", bound, bound)

			drift := binary.LittleEndian.Uint32(m[56:60])
			log.Printf("drift: 0x%X %d\n", drift, drift)

			reserved := binary.LittleEndian.Uint32(m[60:64])
			log.Printf("reserved: 0x%X %d\n", reserved, reserved)

			status := binary.LittleEndian.Uint32(m[64:68])
			log.Printf("clock_status: 0x%X %d\n", status, status)

			earliest := ts.Add(-1 * (time.Nanosecond * time.Duration(bound)))
			latest := ts.Add(time.Nanosecond * time.Duration(bound))

			unix_ns := latest.UnixNano() - (latest.UnixNano()-earliest.UnixNano())/2
			log.Printf("up : %v\n", latest)
			log.Printf("now: %v\n", fromUnixNano(uint64(unix_ns)))
			log.Printf("low: %v\n", earliest)
			log.Printf("range: %v\n", latest.Sub(earliest))

			// now, err := c.Now()
			// if err != nil {
			// 	log.Println(err)
			// 	continue
			// }

			// if now.Header.Unsynchronized {
			// 	log.Println("Unsynchronized")
			// } else {
			// 	log.Println("Synchronized")
			// }

			// log.Println("Current: ", now.Time)
			// log.Println("Earliest:", now.Bound.Earliest)
			// log.Println("Latest:  ", now.Bound.Latest)
			// log.Println("Range:   ", now.Bound.Latest.Sub(now.Bound.Earliest))
		}
	}()

	// Interrupt handler.
	go func() {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
		log.Println("signal:", <-sigch)
		cancel()
	}()

	<-done

	ticker.Stop()

	if err := m.Unmap(); err != nil {
		log.Println("Unmap failed:", err)
		return
	}
}

func fromUnixNano(nano uint64) time.Time {
	return time.Unix(int64(nano/1e9), int64(nano%1e9))
}
