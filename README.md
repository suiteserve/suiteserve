# SuiteServe
Test reporting API and real-time UI.

## Developing
For the best development experience, install the following:
- [Docker](https://www.docker.com)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [GNU Make](https://www.gnu.org/software/make/)
- [Go](https://golang.org)
- [migrate](https://github.com/golang-migrate/migrate)
  - Install with `go get -tags 'mongodb' -u github.com/golang-migrate/migrate/v4/cmd/migrate`.
  - Ensure that `$GOPATH/bin` is in your path.
- [mkcert](https://github.com/FiloSottile/mkcert)
- [Node.js](https://nodejs.org)
- [NPM](https://www.npmjs.com)

### Run for Development
First, run `make tls/cert.pem` to generate the TLS certificate and key for development-only use with SuiteServe, `webpack-dev-server`, and Mongo Express. This command also installs the root CA into your web browser in order to avoid security warnings, but you may have to restart your browser for that to take effect. You only have to do this once.

Next, bring up MongoDB:
```bash
$ make dev-db
```

This will start the MongoDB Docker container and then provision it with the necessary users and permissions, followed by the migrations. To bring down the container, run:
```bash
$ cd db
$ docker-compose down
```

Optionally, purge the database with `docker-compose rm -fv`. Keep MongoDB running during development.

Now, start SuiteServe:
```bash
$ go run cmd/suiteserve/main.go -debug -seed
```

The `-debug` option adds precise timestamps and code locations to log messages. The `-seed` option inserts sample data into the database if the database tables are empty.

Finally, in another terminal, start `webpack-dev-server`:
```bash
$ cd ui
$ npm start
```

The following services will now be available:
- **SuiteServe** &mdash; [https://localhost:8080](https://localhost:8080)
  - Serves the UI and test reporting API.
  - Does not hot-reload on code changes.
  - Code changes in `ui/` will not be seen until `ui/dist/` is built.
- **`webpack-dev-server`** &mdash; [https://localhost:8081](https://localhost:8081)
  - Serves the UI and forwards non-UI requests to [localhost:8080](https://localhost:8080).
  - Hot-reloads on code changes in `ui/`.
  - Does not require `ui/dist/` to exist.
  - Useful for UI development, but not for production.
- **Mongo Express** &mdash; [https://localhost:8082](https://localhost:8082)

As needed, build or rebuild `ui/dist/` with:
```bash
$ rm -rf ui/dist
$ make ui/dist
```

### Build for Production
Run `make` to build the SuiteServe binary file named "suiteserve".

### Build for Docker
Build the SuiteServe Docker image:
```bash
$ docker build -t suiteserve .
```

To start a container, make sure a MongoDB instance is available and then update `config/config.json` to point to it. Run:
```bash
$ docker run -v $(pwd)/config:/app/config -v $(pwd)/tls:/app/tls suiteserve
```
