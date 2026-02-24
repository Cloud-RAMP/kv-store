# KV-Store

## Running the System

To run the server, fist make sure that go is installed on your system. Our KV store runs on go 1.24, so make sure you download that version or one later. To start the server, run the command `make server`. If you want a slightly higher performance version, run `make build` and then `./server`. The compiler will do some magic to make it like ~5% faster

To run the python test, make sure the server is running and then run `make test`. To run the go benchmark test, navigate to the `test/benchmark_test.go` file and open it in VSCode with go extensions installed. Then, hit the "run benchmark" button above the `BenchmarkKVStore` function. You should then see results output in your terminal.

> **Note**: The first time the benchmark is run, there will be many key errors due to the multi-threaded nature of the test. This is expected and should not happen if the test is rerun or if the server is restarted with the persistent saving file populated.

## System Design

The design of this system is realtively simple. There are two manin layers:
* The server that accepts user connections and calls backend functions accordingly
* The backend which stores data and exposes GET, PUT, and DEL functions

### Server

The server runs on local port 3000, and accepts requests for any route. When a request comes in, the server automatically spins up a new goroutine to handle the request. A user can send the following types of HTTP requests to denote the operation they want to perform:

* GET - returns the key corresponding to the path requested
  * Ex. GET /hello returns the value corresponding to the key "hello", or a 404 error if it does not exist
* POST - sets the value of the key corresponding to the path requested
  * The value that you want to set the key to should be defined in a JSON body, as such: `{"value":"some_value"}`
  * This does not return an error, only 200
* DEL - deletes the key corresponding to the path requested
  * This operation does not error, and only returns a 200
  * If the key does not exist, it will still return a 200

### Backend Storage

Data is stored in memory with nested map structs. Each index of the main map contains an inner map, which keys access based on a deterministic hash. This is done to reduce lock contention, as we can spread locking across multiple maps with this implementation, as opposed to a single lock for one single map.

The backend data store is saved to the file path `internal/store/save.gob` and is encoded with the `gob` library. `gob` defines its own protocol for encoding go objects, allowing for easy decoding into go structures when reading the data. This saving is done once every 5 seconds on a goroutine that is spawned on project initialization.

The backend exposes a small set of fucntions that are to be used to manipulate the data store:
* Get
  * Returns the value corresponding to the requested key
  * Returns an error if it doesn't exist
* Put
  * Adds the key/value pair to the internal store
  * Does not return anything
* Del
  * Removes the requested key from the internal store
  * Does not return anything

## Possible Improvements

One possible idea that we had for system improvement was to utilize a worker pool instead of spawning a goroutine for each user request. Although goroutines are known for being lightweight and easy to spin up, creating one for every request could result in more overhead. We tested this implementation in a separate git branch but could not find performance improvements over the default implementation. It is possible that this method could be done in a way to improve performance, but we did not find that.

Another improvement to our KV store would be to operate on TCP instead of HTTP. Accepting requests with a TCP server library would allow for more fine-grained parsing of requests, since most information parsed into the `net.Request` object is unutilized by our server. By using TCP, we could simply examine the necessary fields, like request method, path, and body. However, this would result in more security vulnerabilities. If this were production-grade software, we should operate on HTTP instead of TCP.

## Issues

The main issue that we have faced with our implementation is missing keys in testing. The first time a test is run, the user may experience many `404 NOT FOUND` errors. If the `benchmark_test.go` test is run with one thread on a clean store, there are no `404 NOT FOUND` errors. However, if the same is done with more than one thread, some keys are marked as not found.

We hypothesize that this is due to the multi-threaded nature of these tests, and specifically how threads are scheduled. Because some threads may be scheduled before others, even if their operations depend on each other, this reuslts in some `GET {key}` operations happening before the `PUT {key} {value}` operations. This can be circumvented by running the testing script one additional time to seed the values in the KV store.

## Assumptions

* The user has go installed on their system
* The user is making their requests in an order where if they expect a key to be present, a `PUT` request verifiably came before thier `GET`