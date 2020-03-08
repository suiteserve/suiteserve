FROM golang:1.14-alpine AS builder
WORKDIR /go/src/git.blazey.dev/tests/
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN CGO_ENABLED=0 go install

FROM scratch
WORKDIR /go/bin/
COPY --from=builder /go/bin/tests .
ENV HOST 0.0.0.0
ENV PORT 8080
EXPOSE $PORT
ENTRYPOINT ["./tests"]