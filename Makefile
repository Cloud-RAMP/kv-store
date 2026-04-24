BUILD_DIR=bin
OUTPUT_DIR=output

STORE_DIR=store/cmd
STORE_TARGET=$(BUILD_DIR)/store
ROUTER_DIR=router/cmd
ROUTER_TARGET=$(BUILD_DIR)/router

run: build
	mkdir -p $(OUTPUT_DIR)
	./$(ROUTER_TARGET) > $(OUTPUT_DIR)/router.txt & \
	./$(STORE_TARGET) -port 3001 > $(OUTPUT_DIR)/store1.txt & \
	./$(STORE_TARGET) -port 3002 > $(OUTPUT_DIR)/store2.txt & \
	./$(STORE_TARGET) -port 3003 > $(OUTPUT_DIR)/store3.txt & \
	wait

stop:
	lsof -ti :3000 -ti :3001 -ti :3002 -ti :3003 | xargs kill -9

bench:
	go test -bench=^BenchmarkKVStore$$ -benchmem -run=^$$ ./test

bench-python:
	python3 test/benchmark.py

build-store:
	mkdir -p $(BUILD_DIR)
	go build -o $(STORE_TARGET) $(STORE_DIR)/main.go

build-router:
	mkdir -p $(BUILD_DIR)
	go build -o $(ROUTER_TARGET) $(ROUTER_DIR)/main.go

build: build-store build-router

clean:
	rm -rf save.gob $(BUILD_DIR) $(OUTPUT_DIR)
