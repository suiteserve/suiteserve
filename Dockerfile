FROM golang:1.14-alpine AS api-builder
WORKDIR /app/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/suiteserve
RUN CGO_ENABLED=0 go install

FROM node:14-alpine AS ui-builder
WORKDIR /app/
# TODO: currently getting integrity check error
# COPY ui/package*.json ./
COPY ui/package.json ./
RUN npm i
COPY ui/ ./
RUN npm run build

FROM scratch
ARG CONFIG_FILE=config/config.json
WORKDIR /app/
COPY --from=api-builder /go/bin/suiteserve ./
COPY $CONFIG_FILE config/
COPY --from=ui-builder /app/dist/ ui/dist/
EXPOSE 8080
VOLUME /app/data/
ENTRYPOINT ["./suiteserve"]
