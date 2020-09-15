suiteserve: ui/dist
	CGO_ENABLED=0 go build -o suiteserve cmd/suiteserve/main.go

.PHONY: test
test:
	go test -race -v ./...

.PHONY: dev-db
dev-db:
	cd db; docker-compose up -d

.PHONY: dev-db-migrate-up
dev-db-migrate-up: dev-db
	migrate -database mongodb://ssmigrate:pass@localhost:27017/suiteserve \
		-path db/migrate up

.PHONY: dev-db-migrate-down
dev-db-migrate-down:
	migrate -database mongodb://ssmigrate:pass@localhost:27017/suiteserve \
		-path db/migrate down

ui/node_modules:
	cd ui; npm i

ui/dist: ui/node_modules
	cd ui; npm run build

tls/ca.pem tls/cert.pem tls/key.pem:
	mkcert -install -cert-file tls/cert.pem -key-file tls/key.pem \
		localhost localhostusercontent 127.0.0.1 ::1
	cp "`mkcert -CAROOT`/rootCA.pem" tls/ca.pem

.PHONY: clean
clean:
	cd data; find . ! -name . ! -name .gitignore -exec rm -r {} +
	rm -rf ui/dist ui/node_modules
	rm -f tls/ca.pem tls/cert.pem tls/key.pem suiteserve
