FROM node:14-alpine AS frontend-builder
WORKDIR /app/
COPY frontend/package*.json ./
RUN npm i
COPY frontend/ ./
RUN npm run build

FROM golang:1.14-alpine AS api-builder
WORKDIR /go/src/testpass/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go install

FROM scratch
WORKDIR /app/
COPY --from=api-builder /go/bin/testpass ./
COPY --from=frontend-builder /app/dist/ frontend/dist/
ENV HOST 0.0.0.0
ENV PORT 8080
ENV MONGO_HOST mongo
EXPOSE $PORT
VOLUME /app/storage/ /app/tls/
ENTRYPOINT ["./testpass"]
