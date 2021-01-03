package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const workerPoolSize = 4

func main() {
	service := Service{
		ingestChan: make(chan int, 1),
		jobsChan:   make(chan int, workerPoolSize),
	}
	// create the consumer

	// Simulate external lib sending us 10 events per second
	producer := Producer{callbackFunc: service.callbackFunc}

	go producer.start()

	// Set up cancellation context and waitgroup
	ctx, cancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Start service with cancellation context passed
	go service.startService(ctx)

	// Start workers and Add [workerPoolSize] to WaitGroup
	wg.Add(workerPoolSize)
	for i := 0; i < workerPoolSize; i++ {
		go service.workerFunc(wg, i)
	}

	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan // Blocks here until interrupted

	// Handle shutdown
	fmt.Println("*********************************\nShutdown signal received\n*********************************")
	cancelFunc() // Signal cancellation to context.Context
	wg.Wait()    // Block here until are workers are done

	fmt.Println("All workers done, shutting down!")
}
