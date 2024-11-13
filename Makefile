build:
	go build -o ./bin/gotrack

test:
	go test -v ./...

copy: build 
	cp bin/* $(COPY_DIR)
	chmod +x $(COPY_DIR)/gotrack

.PHONY: build test copy
