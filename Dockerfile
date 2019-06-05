FROM golang:1.12-alpine AS builder
ENV GO111MODULE "on"
RUN mkdir -p /go/src/github.com/zerok && apk add --no-cache git
WORKDIR /go/src/github.com/zerok/podcast-archiver
COPY . .
RUN cd /go/src/github.com/zerok/podcast-archiver/cmd/podcast-archiver && go build

FROM alpine:3.8
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/zerok/podcast-archiver/cmd/podcast-archiver/podcast-archiver /usr/local/bin
ENTRYPOINT ["/usr/local/bin/podcast-archiver"]
