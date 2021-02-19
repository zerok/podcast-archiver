all: bin/podcast-archiver
linux: bin/podcast-archiver-linux

bin:
	mkdir -p bin

bin/podcast-archiver: $(shell find . -name '*.go')
	cd cmd/podcast-archiver && go build -o ../../$@

bin/podcast-archiver-linux: $(shell find . -name '*.go')
	cd cmd/podcast-archiver && GOOS=linux GOARCH=amd64 go build -o ../../$@

clean:
	rm -rf bin

test:
	go test ./... -v

.PHONY: all clean linux test
