suiteserve: test ui/dist
	CGO_ENABLED=0 go build -o suiteserve cmd/suiteserve/main.go

.PHONY: test
test:
	go test -race -v ./...

ui/node_modules:
	cd ui; npm i

ui/dist: ui/node_modules
	cd ui; npm run build

tls/ca.pem tls/cert.pem tls/key.pem:
	mkcert -install -cert-file tls/cert.pem -key-file tls/key.pem \
    		 localhost localhostusercontent 127.0.0.1 ::1
	cp "`mkcert -CAROOT`/rootCA.pem" tls/ca.pem

.PHONY: rethinkdb
rethinkdb:
	cd rethinkdb; docker-compose up -d

.PHONY: db-provision
db-provision: rethinkdb
	go run cmd/dbprovision/main.go -pass config/rethinkdb_pass \
		suiteserve suiteserve attachments suites cases logs

.PHONY: clean
clean:
	cd data; find . ! -name . ! -name .gitignore -exec rm -r {} +
	cd tls; find . ! -name . ! -name .gitignore -exec rm -r {} +
	rm -rf ui/dist ui/node_modules
	rm -f suiteserve
