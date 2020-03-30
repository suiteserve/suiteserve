FROM golang:1.14-alpine AS builder
WORKDIR /go/src/testpass/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go install

FROM scratch
WORKDIR /app/
COPY --from=builder /go/bin/testpass .
ENV HOST 0.0.0.0
ENV PORT 8080
ENV MONGO_HOST mongo
ENV MONGO_USER root
ENV MONGO_PASS pass
EXPOSE $PORT
VOLUME /app/data/ /app/tls/
ENTRYPOINT ["./testpass"]