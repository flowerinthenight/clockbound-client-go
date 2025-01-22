package clockboundclient

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
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
	// As-of-timestamp
	asof_s := binary.LittleEndian.Uint64(c.m[16:24])
	asof_ns := binary.LittleEndian.Uint64(c.m[24:32])
	asof := time.Unix(int64(asof_s), int64(asof_ns))

	// Void-after-timestamp
	va_s := binary.LittleEndian.Uint64(c.m[32:40])
	va_ns := binary.LittleEndian.Uint64(c.m[40:48])
	voidAfter := time.Unix(int64(va_s), int64(va_ns))

	log.Printf("as_of_ts  : %v\n", asof.Format(time.RFC3339Nano))
	log.Printf("void_after: %v\n", voidAfter.Format(time.RFC3339Nano))

	bound := binary.LittleEndian.Uint64(c.m[48:56])
	log.Printf("bound_ns: 0x%X %d\n", bound, bound)

	status := binary.LittleEndian.Uint32(c.m[64:68])
	log.Printf("clock_status: 0x%X %d\n", status, status)

	earliest := asof.Add(-1 * (time.Nanosecond * time.Duration(bound)))
	latest := asof.Add(time.Nanosecond * time.Duration(bound))

	log.Printf("earliest: %v\n", earliest.Format(time.RFC3339Nano))
	log.Printf("latest  : %v\n", latest.Format(time.RFC3339Nano))
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
