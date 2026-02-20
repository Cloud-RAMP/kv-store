package benchmark_test

// Code was written in companion with DeepSeek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"
)

type Operation int

const (
	GET Operation = iota
	SET
)

func (o Operation) String() string {
	if o == GET {
		return "GET"
	}
	return "POST" // SET operations typically use POST or PUT
}

type RequestTask struct {
	key   string
	value string
	op    Operation
}

type SetRequest struct {
	Value string `json:"value"`
}

type Metrics struct {
	mu       sync.Mutex
	duration time.Duration
	success  int
	failure  int
	wg       sync.WaitGroup
}

const SERVER_URL string = "http://localhost:3000"
const NUM_WORKERS = 3
const TOTAL_REQUESTS = 1000

// BenchmarkKVStore benchmarks the KV store with multiple workers
func BenchmarkKVStore(b *testing.B) {
	// Create task channel
	tasks := make(chan RequestTask)

	metrics := Metrics{}

	// Start worker goroutines
	for range NUM_WORKERS {
		metrics.wg.Add(1)
		go worker(b, tasks, &metrics)
	}

	// Generate and send tasks
	b.ResetTimer() // Reset timer to exclude setup time

	// looping over b.N becuase the benchmark automatically controls how many interations are done
	startTime := time.Now()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100) // Cycle through 100 keys
		value := fmt.Sprintf("value-%d", i%100)

		// Create a SET task
		tasks <- RequestTask{
			key:   key,
			value: value,
			op:    SET,
		}

		// Create a GET task
		tasks <- RequestTask{
			key:   key,
			value: value,
			op:    GET,
		}
	}
	close(tasks) // No more tasks to send, close the channel

	// Wait for all workers to complete
	metrics.wg.Wait()
	totalTime := time.Since(startTime)

	// Report metrics
	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	b.ReportMetric(float64(metrics.success)/totalTime.Seconds(), "req/sec")
	b.ReportMetric(float64(metrics.success)/float64(b.N)*100, "success%")
	b.ReportMetric(float64(metrics.duration.Microseconds())/float64(metrics.success), "us/req")

	// b.N*2 becuase we send 2 requests / benchmark
	b.Logf("Total requests: %d, Success: %d, Failures: %d, Time: %v",
		b.N*2, metrics.success, metrics.failure, totalTime)
}

// worker processes tasks from the tasks channel
func worker(b *testing.B, tasks <-chan RequestTask, metrics *Metrics) {
	defer metrics.wg.Done()

	// Create HTTP client with timeout (not strictly necessary)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// every goroutine waits on the task chan and picks up when it can
	for task := range tasks {
		start := time.Now()

		var req *http.Request
		var err error

		// Check out the task operation.
		// If SET, add a JSON body with the value
		switch task.op {
		case SET:
			setReq := SetRequest{Value: task.value}
			jsonData, err := json.Marshal(setReq)
			if err != nil {
				b.Logf("Failed to marshal JSON: %v", err)
				metrics.mu.Lock()
				metrics.failure++
				metrics.mu.Unlock()
				continue
			}

			req, err = http.NewRequest("POST", SERVER_URL+"/"+task.key, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

		case GET:
			req, err = http.NewRequest("GET", SERVER_URL+"/"+task.key, nil)
		}

		if err != nil {
			b.Logf("Failed to create request: %v", err)
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			b.Logf("Failed to send request: %v", err)
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		// Read response
		data, err := io.ReadAll(resp.Body)
		// _, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			b.Logf("Failed to read response: %v", err)
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		// Check status code
		if resp.StatusCode >= 400 {
			b.Logf("Request failed: %s", resp.Status)
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		if task.op == GET && string(data) != task.value {
			b.Logf("Request did not get expected value: %s, got %s", task.value, string(data))
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		// Record success
		elapsed := time.Since(start)
		metrics.mu.Lock()
		metrics.success++
		metrics.duration += elapsed
		metrics.mu.Unlock()
	}
}

// RunWithDifferentLoads demonstrates running benchmarks with different configurations
func RunWithDifferentLoads(b *testing.B) {
	// Test with different worker counts
	workerCounts := []int{5, 10, 20, 50, 100}

	for _, workers := range workerCounts {
		b.Run(fmt.Sprintf("workers-%d", workers), func(b *testing.B) {
			oldWorkers := NUM_WORKERS
			// This is just an example - in real code you'd need to make NUM_WORKERS configurable
			_ = oldWorkers
			BenchmarkKVStore(b)
		})
	}
}
