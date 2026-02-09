# KV-Store

To run the server, run the command `go run cmd/kv-store/main.go`. If you want a slightly higher performance version, run `go build cmd/kv-store/main.go` and then `./main`. The compiler will do some magic to make it like ~5% faster

To test, make sure the server is running and then run `python3 test/benchmark.py`

### Future work
* Operate on TCP instead of HTTP for faster performance
  * Requires more manual request parsing
* Use a worker pool to avoid spinning up a new goroutine on each request (huge overhead), maybe limit to like 50
* Add incremental saving (extra credit!). Use `gob` package to serialize data easily. Could also probably save as json too
  * Load this on startup
