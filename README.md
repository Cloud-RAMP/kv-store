# KV-Store

To run the server, run the command `make server`. If you want a slightly higher performance version, run `make build` and then `./server`. The compiler will do some magic to make it like ~5% faster

To test, make sure the server is running and then run `make test`

> **Note**: The first time the benchmark is run, there will be many key errors due to the multi-threaded nature of the test. This is expected and should not happen if the test is rerun or if the server is restarted with the persistent saving file populated.

### Future work
* Operate on TCP instead of HTTP for faster performance
  * Requires more manual request parsing
* Use a worker pool to avoid spinning up a new goroutine on each request (huge overhead), maybe limit to like 50
* Add incremental saving (extra credit!). Use `gob` package to serialize data easily. Could also probably save as json too
  * Load this on startup
* Add throughput vs latency plot?
