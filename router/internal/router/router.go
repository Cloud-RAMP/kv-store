package router

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 5 * time.Second

// ParseNodes reads a comma-separated list of backend URLs.
func ParseNodes(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	nodes := make([]string, 0, len(parts))

	for _, part := range parts {
		node := strings.TrimSpace(part)
		if node == "" {
			continue
		}
		u, err := url.Parse(node)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("invalid node URL: %q", node)
		}
		nodes = append(nodes, strings.TrimRight(node, "/"))
	}

	if len(nodes) == 0 {
		return nil, fmt.Errorf("no backend nodes configured")
	}

	return nodes, nil
}

// Start runs the hash-based routing HTTP server.
func Start(address string, nodes []string) error {
	r := &Router{
		nodes: nodes,
		client: &http.Client{Timeout: defaultTimeout},
	}

	h := http.HandlerFunc(r.handle)
	fmt.Printf("Router listening on %s with %d nodes\n", address, len(nodes))
	return http.ListenAndServe(address, h)
}

type Router struct {
	nodes  []string
	client *http.Client
}

func (r *Router) handle(w http.ResponseWriter, req *http.Request) {
	key := strings.TrimPrefix(req.URL.Path, "/")
	if key == "" {
		http.Error(w, "key required in path", http.StatusBadRequest)
		return
	}

	node := pickNode(key, r.nodes)
	targetURL := node + req.URL.Path
	if req.URL.RawQuery != "" {
		targetURL += "?" + req.URL.RawQuery
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	upstreamReq, err := http.NewRequest(req.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		http.Error(w, "failed to build upstream request", http.StatusInternalServerError)
		return
	}

	copyHeaders(upstreamReq.Header, req.Header)
	upstreamReq.Header.Set("X-Forwarded-For", req.RemoteAddr)

	resp, err := r.client.Do(upstreamReq)
	if err != nil {
		http.Error(w, "backend unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.Header().Set("X-KV-Node", node)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func pickNode(key string, nodes []string) string {
	bestNode := nodes[0]
	bestScore := hashPair(key, bestNode)

	for i := 1; i < len(nodes); i++ {
		score := hashPair(key, nodes[i])
		if score > bestScore {
			bestNode = nodes[i]
			bestScore = score
		}
	}

	return bestNode
}

func hashPair(key string, node string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	_, _ = h.Write([]byte("|"))
	_, _ = h.Write([]byte(node))
	return h.Sum64()
}

func copyHeaders(dst http.Header, src http.Header) {
	for key, vals := range src {
		dst.Del(key)
		for _, v := range vals {
			dst.Add(key, v)
		}
	}
}
