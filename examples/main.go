package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	clockboundclient "github.com/flowerinthenight/clockbound-client-go"
)

func main() {
	client, err := clockboundclient.New()
	if err != nil {
		log.Println("New failed:", err)
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

			client.Now()
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
	client.Close()
}

func fromUnixNano(nano uint64) time.Time {
	return time.Unix(int64(nano/1e9), int64(nano%1e9))
}
