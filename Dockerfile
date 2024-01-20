FROM golang:1.21-alpine AS builder
ARG VERSION
ARG COMMIT
RUN mkdir -p /go/src/github.com/zerok && apk add --no-cache git
WORKDIR /go/src/github.com/zerok/podcast-archiver
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    cd /go/src/github.com/zerok/podcast-archiver/cmd/podcast-archiver && \
    go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT"

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/zerok/podcast-archiver/cmd/podcast-archiver/podcast-archiver /usr/local/bin
ENTRYPOINT ["/usr/local/bin/podcast-archiver"]
