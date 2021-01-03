package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

//Service A wrapper struct encapsulating the Producer and Consumer
type Service struct {
	// Represent a proxy between the callback and the jobs channel
	ingestChan chan int
	jobsChan   chan int
}

// callbackFunc is invoked each time the external lib passes an event to us.
func (sc Service) callbackFunc(event int) {
	sc.ingestChan <- event
}

// workerFunc starts a single worker function that will range on the jobsChan until that channel closes.
func (sc Service) workerFunc(wg *sync.WaitGroup, index int) {
	defer wg.Done()

	fmt.Printf("Worker %d starting\n", index)
	for eventIndex := range sc.jobsChan {
		// simulate work taking between 1-3 seconds
		fmt.Printf("Worker %d started job %d\n", index, eventIndex)
		time.Sleep(time.Millisecond * time.Duration(1000+rand.Intn(2000)))
		fmt.Printf("Worker %d finished processing job %d\n", index, eventIndex)
	}
	fmt.Printf("Worker %d interrupted\n", index)
}

// startService acts as the proxy between the ingestChan and jobsChan, with a select to support graceful shutdown.
func (sc Service) startService(ctx context.Context) {
	for {
		select {
		case job := <-sc.ingestChan:
			sc.jobsChan <- job
		case <-ctx.Done():
			fmt.Println("Consumer received cancellation signal, closing jobsChan!")
			close(sc.jobsChan)
			fmt.Println("Consumer closed jobsChan")
			return
		}
	}
}

// Producer simulates an external library that invokes the
// registered callback when it has new data for us once per 100ms
type Producer struct {
	callbackFunc func(event int)
}

func (p Producer) start() {
	eventIndex := 0
	for {
		p.callbackFunc(eventIndex)
		eventIndex++
		time.Sleep(time.Millisecond * 100)
	}
}
