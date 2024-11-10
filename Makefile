build:
	go build -o ./bin/gotrack

test:
	go test -v ./...

copy: build 
	cp bin/* $(COPY_DIR)

.PHONY: build test copy
