all: podcast-archiver
linux: podcast-archiver-linux

podcast-archiver: $(shell find . -name '*.go')
	cd cmd/podcast-archiver && go build -o ../../$@

podcast-archiver-linux: $(shell find . -name '*.go')
	cd cmd/podcast-archiver && GOOS=linux GOARCH=amd64 go build -o ../../$@

clean:
	rm -f podcast-archiver
	rm -f podcast-archiver-linux

.PHONY: all clean linux
