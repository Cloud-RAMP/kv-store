package benchmark_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

type RouterOperation int

const (
	RouterGET RouterOperation = iota
	RouterSET
)

type RouterRequestTask struct {
	key   string
	value string
	op    RouterOperation
}

type RouterSetRequest struct {
	Value string `json:"value"`
}

type RouterMetrics struct {
	mu            sync.Mutex
	duration      time.Duration
	success       int
	failure       int
	nodeHitCounts map[string]int
	wg            sync.WaitGroup
}

const routerWorkers = 5

func routerURL() string {
	if v := os.Getenv("ROUTER_URL"); v != "" {
		return v
	}
	return "http://localhost:3000"
}

// BenchmarkRouterKVStore benchmarks requests sent through the router.
// Usage:
//   ROUTER_URL=http://localhost:3000 go test ./test -bench BenchmarkRouterKVStore -benchmem
func BenchmarkRouterKVStore(b *testing.B) {
	tasks := make(chan RouterRequestTask)
	metrics := RouterMetrics{nodeHitCounts: make(map[string]int)}

	for range routerWorkers {
		metrics.wg.Add(1)
		go routerWorker(b, tasks, &metrics)
	}

	b.ResetTimer()
	startTime := time.Now()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("router-key-%d", i%200)
		value := fmt.Sprintf("router-value-%d", i%200)

		tasks <- RouterRequestTask{key: key, value: value, op: RouterSET}
		tasks <- RouterRequestTask{key: key, value: value, op: RouterGET}
	}
	close(tasks)

	metrics.wg.Wait()
	totalTime := time.Since(startTime)

	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	if metrics.success > 0 {
		b.ReportMetric(float64(metrics.success)/totalTime.Seconds(), "req/sec")
		b.ReportMetric(float64(metrics.duration.Microseconds())/float64(metrics.success), "us/req")
	}
	b.ReportMetric(float64(metrics.success)/float64(2*b.N)*100, "success%")

	b.Logf("Router benchmark complete: requests=%d success=%d failures=%d elapsed=%v", 2*b.N, metrics.success, metrics.failure, totalTime)
	b.Logf("Router node hit distribution (from X-KV-Node): %+v", metrics.nodeHitCounts)
}

// TestRouterConsistentNodeSelection verifies that repeated requests for the same key
// are routed to the same backend node (based on X-KV-Node header).
func TestRouterConsistentNodeSelection(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := routerURL()
	key := "router-consistency-key"
	value := "router-consistency-value"

	setReqBody, err := json.Marshal(RouterSetRequest{Value: value})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	setReq, err := http.NewRequest("POST", baseURL+"/"+key, bytes.NewBuffer(setReqBody))
	if err != nil {
		t.Fatalf("failed to create set request: %v", err)
	}
	setReq.Header.Set("Content-Type", "application/json")

	setResp, err := client.Do(setReq)
	if err != nil {
		t.Fatalf("set request failed: %v", err)
	}
	_ = setResp.Body.Close()
	if setResp.StatusCode >= 400 {
		t.Fatalf("set request returned status %s", setResp.Status)
	}

	nodeSeen := ""
	for i := 0; i < 20; i++ {
		getReq, err := http.NewRequest("GET", baseURL+"/"+key, nil)
		if err != nil {
			t.Fatalf("failed to create get request: %v", err)
		}

		resp, err := client.Do(getReq)
		if err != nil {
			t.Fatalf("get request failed: %v", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			t.Fatalf("failed reading get response body: %v", readErr)
		}
		if resp.StatusCode >= 400 {
			t.Fatalf("get request returned status %s", resp.Status)
		}
		if string(body) != value {
			t.Fatalf("unexpected value for key %q: got %q want %q", key, string(body), value)
		}

		node := resp.Header.Get("X-KV-Node")
		if node == "" {
			t.Fatalf("missing X-KV-Node header in response")
		}
		if nodeSeen == "" {
			nodeSeen = node
			continue
		}
		if node != nodeSeen {
			t.Fatalf("inconsistent routing for key %q: first node=%q, later node=%q", key, nodeSeen, node)
		}
	}
}

func routerWorker(b *testing.B, tasks <-chan RouterRequestTask, metrics *RouterMetrics) {
	defer metrics.wg.Done()

	client := &http.Client{Timeout: 5 * time.Second}
	baseURL := routerURL()

	for task := range tasks {
		start := time.Now()

		var req *http.Request
		var err error

		switch task.op {
		case RouterSET:
			payload, marshalErr := json.Marshal(RouterSetRequest{Value: task.value})
			if marshalErr != nil {
				b.Logf("failed to marshal JSON: %v", marshalErr)
				metrics.mu.Lock()
				metrics.failure++
				metrics.mu.Unlock()
				continue
			}
			req, err = http.NewRequest("POST", baseURL+"/"+task.key, bytes.NewBuffer(payload))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
			}
		case RouterGET:
			req, err = http.NewRequest("GET", baseURL+"/"+task.key, nil)
		default:
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		if err != nil {
			b.Logf("failed to create request: %v", err)
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			b.Logf("failed to send request: %v", err)
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		data, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			b.Logf("failed to read response: %v", readErr)
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		if resp.StatusCode >= 400 {
			b.Logf("request failed: %s", resp.Status)
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		if task.op == RouterGET && string(data) != task.value {
			b.Logf("unexpected value for key %s: got=%s want=%s", task.key, string(data), task.value)
			metrics.mu.Lock()
			metrics.failure++
			metrics.mu.Unlock()
			continue
		}

		elapsed := time.Since(start)
		node := resp.Header.Get("X-KV-Node")
		metrics.mu.Lock()
		metrics.success++
		metrics.duration += elapsed
		if node != "" {
			metrics.nodeHitCounts[node]++
		}
		metrics.mu.Unlock()
	}
}
