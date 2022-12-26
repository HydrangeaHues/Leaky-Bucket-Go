# Leaky-Bucket-Go
A leaky bucket rate limiter simulation in Go

### Functionality
This program simulates a server receiving and rate limiting requests using the leaking bucket rate limiter algorithm. The program starts with a worker pool of 3 workers that are ready to pull requests out of the bucket and process them. As time goes on, when the bucket fills up until it has less than 10% free space remaining, the worker pool will be automatically scaled up to process the requests more quickly. When the bucket is < 10% full of requests, the worker pool will be automatically scaled back down to 3 workers. When the bucket becomes entirely full, all incoming requests will be dropped. When the bucket becomes entirely empty, any active workers will sleep for 10 seconds to allow more requests to come in. Read more about the leaking bucket rate limiter algorithm below and about certain design decisions that were made for this demo.

### Background
#### Leaking Bucket Algorithm
  - Requests are placed in a queue of finite size and processed at a fixed rate. If a request comes and the queue is full, the request is rejected, otherwise it is added to the queue (accepted).
  - Pros: Memory efficient and can provide stable output of requests if that is desired.
  - Cons: Potentially hard to tune to get the best performance for your use case, and a burst of traffic could cause new requests to continually be dropped while old requests are processed.

### Design Decisions

#### Why Go?
I wanted to use Go's built-in concurrency constructs to simulate a server receiving and rate limiting requests. This was an excuse for me to work with Go and implement a rate limiting algorithm I hadn't before.

#### A single global bucket
I chose to just use a single leaky bucket that would be responsible for limiting all simulated traffic / requests. This type of approach could be useful in production if we wanted to ensure that our servers would never get overwhelmed by the amount of requests coming in, but there are likely better ways to control global request volume.

#### Workers capable of working on any bucket
I designed the `processRequests` function and the `worker` struct in such a way that any worker could process requests from any bucket. In a real world setting it might make more sense to have certain workers dedicated to certain buckets, depending on the type of work being done, the request load a bucket typically sees, etc. You could refactor the `worker` struct to have a `leakyBucket` attribute and remove the `leakyBucket` parameter from `processRequests` to accomplish this.

#### Buffered channel versus array or slice for request buffer
I chose to use a buffered channel to hold the requests coming into my bucket because it accomplishes the goals of being a fixed size and the requests being processed in the same order they were received. Realistically things could be refactored such that an array or slice was used instead, but I think a buffered channel does the job well. A buffered channel (as opposed to an unbuffered channel) allows the requests to come in asynchronously of the workers pulling requests out of the channel, which I feel simulates a real life system better.
