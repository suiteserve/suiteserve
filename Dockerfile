FROM golang:1.14-alpine AS api-builder
WORKDIR /app/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/suiteserve
RUN CGO_ENABLED=0 go install

FROM node:14-alpine AS frontend-builder
WORKDIR /app/
COPY frontend/package*.json ./
RUN npm i
COPY frontend/ ./
RUN npm run build

FROM scratch
ARG CONFIG_FILE=config/config.json
WORKDIR /app/
COPY --from=api-builder /go/bin/suiteserve ./
COPY $CONFIG_FILE config/
COPY --from=frontend-builder /app/dist/ frontend/dist/
EXPOSE 8080
VOLUME /app/data/
ENTRYPOINT ["./suiteserve"]
