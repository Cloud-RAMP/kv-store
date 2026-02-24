RUN_DIR=./cmd/kv-store
TARGET=server

.PHONY: test server build clean

server:
	go run ${RUN_DIR}/main.go

test:
	python3 ./test/benchmark.py

build:
	go build -o ${TARGET} ${RUN_DIR}/main.go

clean:
	rm internal/store/save.gob