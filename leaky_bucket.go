package main

import (
	"fmt"
	"time"
)

// request is a barebones modeling of a request a user might make to our theoretical servers.
type request struct {
	requestType string
	requestedAt time.Time
}

// leakyBucket simulates how a leaky bucket rate limiter might be modeled.
type leakyBucket struct {
	requestChannel chan request
	name           string
	// The maximum number of workers allowed to be operating at once on this bucket.
	workerCap int
	// The minimum number of workers that must always be on standby for a bucket.
	workerMin int
}

// worker is intended to be utilized with Go routines to simulate worker processes concurrently
// pulling jobs off the leakyBucket.
type worker struct {
	name string
	// quitChannel is used to send a signal to shut down a worker when scaling the worker pool.
	quitChannel chan bool
}

// receiveRequests simulates potentially what a server receiving traffic could look like.
// This function is intended to be run as a Go routine and contains an infinite loop to simulate
// a constant flow of traffic to the server.
func receiveRequests(bucket leakyBucket) {
	for {
		if len(bucket.requestChannel) < cap(bucket.requestChannel) {
			fmt.Println("New request received!")
			bucket.requestChannel <- request{requestType: "HTML Request", requestedAt: time.Now()}
			time.Sleep(100 * time.Millisecond)
		} else {
			fmt.Println("Request queue full! Dropping requests.")
			time.Sleep(3 * time.Second)
		}
	}
}

// processRequests handles the operations a worker can perform.
// These include pulling requests off the bucket, being killed,
// and sleeping for 10 seconds if no requests are present on the bucket.
// Intended to be run as a Go routine, this function contains an infinite loop
// to keep the worker operating until no longer needed.
func processRequests(worker worker, bucket leakyBucket) {
	for {
		select {
		case req := <-bucket.requestChannel:
			fmt.Printf("%s is processing a request of type %s\n", worker.name, req.requestType)
			time.Sleep(750 * time.Millisecond)
		case <-worker.quitChannel:
			fmt.Printf("Killing %s\n", worker.name)
			return
		default:
			fmt.Printf("All requests processed. %s will sleep for 10 seconds\n", worker.name)
			time.Sleep(10 * time.Second)
		}

	}
}

// workerPoolSizeAdjuster monitors the amount of requests on the bucket argument passed in
// and scales the number of workers operating on it accordingly.
// Workers are added if the bucket only has 10% of its capacity open for requests at a given time,
// and if there is room to add more workers to a buckets pool.
// Workers are removed if the bucket is using less than 10% of its capacity and there are more than 3 workers currently.
func workerPoolSizeAdjuster(pool []worker, bucket leakyBucket) {
	for {
		if len(bucket.requestChannel) > (cap(bucket.requestChannel)-(cap(bucket.requestChannel)/10)) && ((len(pool) + 1) <= bucket.workerCap) {
			fmt.Println("Additional worker being spawned to help process requests.")
			newWorkerName := fmt.Sprintf("Worker %d", len(pool)+1)
			newWorker := worker{name: newWorkerName, quitChannel: make(chan bool)}
			pool = append(pool, newWorker)
			go processRequests(newWorker, bucket)
		} else if (len(bucket.requestChannel) < (cap(bucket.requestChannel) / 10)) && (len(pool)-1 >= bucket.workerMin) {
			fmt.Println("Removing workers due to light request load.")
			pool[len(pool)-1].quitChannel <- true
			pool = pool[:len(pool)-1]
		}
	}
}

// initializeBucket initializes and returns a leakyBucket struct.
func initializeBucket(bucketName string, bucketCapacity int, workerCap int, workerMin int) leakyBucket {
	requests := make(chan request, bucketCapacity)
	return leakyBucket{requestChannel: requests, name: bucketName, workerCap: workerCap, workerMin: workerMin}
}

func main() {
	// Create a bucket to simulate a rate limit that applies to all traffic coming into a server,
	// regardless of origin or purpose.
	globalBucket := initializeBucket("Global Bucket", 20, 5, 3)

	// Initialize a worker pool and add 3 workers to start.
	workerPool := make([]worker, 0)
	workerPool = append(workerPool, worker{name: "Worker 1", quitChannel: make(chan bool)})
	workerPool = append(workerPool, worker{name: "Worker 2", quitChannel: make(chan bool)})
	workerPool = append(workerPool, worker{name: "Worker 3", quitChannel: make(chan bool)})

	// Fill the bucket halfway full with requests to start.
	// This is done just to showcase things more quickly in the demo.
	for i := 0; i < cap(globalBucket.requestChannel)/2; i++ {
		globalBucket.requestChannel <- request{requestType: "Login Attempt", requestedAt: time.Now()}
	}

	// Start simulating our leakyBucket receiving requests.
	go receiveRequests(globalBucket)

	// Spawn a Go routine for each worker in our pool to start processing requests from our leakyBucket.
	for _, item := range workerPool {
		go processRequests(item, globalBucket)
	}

	// Spawn a Go routine to monitor and adjust the worker pool size dynamically.
	go workerPoolSizeAdjuster(workerPool, globalBucket)

	// Empty infinite loop to allow the demo and Go routines to run.
	for {

	}
}
