FROM golang:latest AS builder
RUN mkdir -p /go/src/pipeline
WORKDIR /go/src/pipeline
ADD main.go .
ADD go.mod .
RUN go install .

FROM alpine:latest

WORKDIR /root/
COPY --from=builder /go/bin/main .
ENTRYPOINT ./main