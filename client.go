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

var ClockStatusName = map[ClockStatus]string{
	ClockStatusUnknown:      "UNKNOWN",
	ClockStatusSynchronized: "SYNCHRONIZED",
	ClockStatusFreeRunning:  "FREE_RUNNING",
}

func (cs ClockStatus) String() string { return ClockStatusName[cs] }

// Now represents a range of bounded timestamp from ClockBound.
// The "true" time is somewhere within the range.
type Now struct {
	Earliest time.Time
	Latest   time.Time
	Status   ClockStatus
}

// Client represents a connection to ClockBound's shared memory file.
type Client struct {
	f   *os.File
	m   mmap.MMap
	err error
}

// Now gets a set range of bounded timestamps from ClockBound.
func (c *Client) Now() (Now, error) {
	if c.err != nil {
		return Now{}, c.err
	}

	log.Printf("magic: %X %X %X %X %X %X %X %X\n", c.m[0], c.m[1], c.m[2], c.m[3], c.m[4], c.m[5], c.m[6], c.m[7])

	mg1 := binary.LittleEndian.Uint32(c.m[:4])
	log.Printf("magic1: %X\n", mg1)

	asof_s := binary.LittleEndian.Uint64(c.m[16:24])
	asof_ns := binary.LittleEndian.Uint64(c.m[24:32])
	asof := time.Unix(int64(asof_s), int64(asof_ns))

	// t1_s := binary.BigEndian.Uint64(c.m[16:24])
	// t1_ns := binary.BigEndian.Uint64(c.m[24:32])
	// t1 := time.Unix(int64(t1_s), int64(t1_ns))
	// b := c.m[16:24]
	// tmp := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
	// 	uint64(b[7])<<32 | uint64(b[6])<<40 | uint64(b[5])<<48 | uint64(b[4])<<56
	// log.Printf("tmp: %v\n", tmp)

	t1 := uint64(binary.LittleEndian.Uint32(c.m[16:20]))
	t2 := uint64(binary.LittleEndian.Uint32(c.m[20:24]))
	log.Printf("t_1_2: %d %d %v\n", t1, t2, t1|t2)

	va_s := binary.LittleEndian.Uint64(c.m[32:40])
	va_ns := binary.LittleEndian.Uint64(c.m[40:48])
	voidAfter := time.Unix(int64(va_s), int64(va_ns))

	bound := binary.LittleEndian.Uint64(c.m[48:56])
	status := binary.LittleEndian.Uint32(c.m[64:68])
	earliest := asof.Add(-1 * (time.Nanosecond * time.Duration(bound)))
	latest := asof.Add(time.Nanosecond * time.Duration(bound))

	clockStatus := ClockStatus(status)
	if latest.After(voidAfter) {
		clockStatus = ClockStatusUnknown
	}

	return Now{
		Earliest: earliest,
		Latest:   latest,
		Status:   clockStatus,
	}, nil
}

// Error implements the `error` interface, returning the internal error.
func (c *Client) Error() string {
	switch c.err {
	case nil:
		return fmt.Sprintf("%v", nil)
	default:
		return c.err.Error()
	}
}

// Close releases the opened and memory-mapped file.
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

// New creates an instance of Client.
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
