FROM golang:1.15-alpine AS app-builder
WORKDIR /app/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/suiteserve/
RUN CGO_ENABLED=0 go install

FROM node:14-alpine AS ui-builder
WORKDIR /ui/
COPY ui/package.json ui/yarn.lock ./
RUN yarn
COPY ui/ ./
RUN yarn build

FROM scratch
WORKDIR /app/
COPY --from=app-builder /go/bin/suiteserve ./
COPY --from=ui-builder /ui/build/ ./ui/build/
EXPOSE 8080
VOLUME /app/data/
ENTRYPOINT ["./suiteserve"]
