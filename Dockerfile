FROM golang:1.14-alpine AS app-builder
WORKDIR /app/
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/suiteserve/
RUN CGO_ENABLED=0 go install

FROM node:14-alpine AS ui-builder
WORKDIR /ui/
COPY ui/package*.json ./
RUN npm i
COPY ui/ ./
RUN npm run build

FROM scratch
WORKDIR /app/
COPY --from=app-builder /go/bin/suiteserve ./
COPY --from=ui-builder /ui/dist/ ui/dist/
EXPOSE 8080
VOLUME /app/data/
ENTRYPOINT ["./suiteserve"]
