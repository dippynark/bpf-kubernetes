# Build the bpf binary
FROM golang:1.12.1 as builder

# Copy in the go src
WORKDIR /go/src/github.com/dippynark/bpf-kubernetes
COPY cmd/    cmd/
COPY pkg/    pkg/
COPY vendor/ vendor/
COPY bpf/include/ bpf/include/

# Build
RUN GOOS=linux GOARCH=amd64 go build -tags netgo -a -o bpf-kubernetes github.com/dippynark/bpf-kubernetes/cmd

# Copy binary into thin image
FROM alpine:3.6
RUN apk update && apk add --no-cache libc6-compat
WORKDIR /
COPY --from=builder /go/src/github.com/dippynark/bpf-kubernetes/bpf-kubernetes .
ENTRYPOINT ["/bpf-kubernetes"]
