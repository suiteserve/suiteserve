# SuiteServe
Test reporting API and real-time UI.

## Developing
For the best development experience, ensure that you have installed:

- [Docker](https://www.docker.com)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [GNU Make](https://www.gnu.org/software/make/)
- [Go](https://golang.org)
- [mkcert](https://github.com/FiloSottile/mkcert)
- [Node.js](https://nodejs.org)
- [NPM](https://www.npmjs.com)

### Run for Development
Ensure that RethinkDB is running and provisioned:
```bash
$ make db-provision
```
This will bring up the RethinkDB Docker container and then provision it with the necessary user, database, and tables. To bring down the container, run:
```bash
$ cd rethinkdb
$ docker-compose down
```
Append `-v` to the `docker-compose down` command to also purge the database. Keep RethinkDB running during development.

Run `make tls/cert.pem` to generate the TLS certificate and key for development-only use with SuiteServe and the Webpack DevServer. This command also installs the root CA into your web browser in order to avoid security warnings, but you may have to restart your browser for it to take effect.

Now start SuiteServe:
```bash
$ go run cmd/suiteserve/main.go -debug -seed
```
The `-debug` option adds precise timestamps and code locations to log messages. The `-seed` option inserts sample data into the database if the database tables are empty.

In another terminal, start the Webpack DevServer:
```bash
$ cd ui
$ npm start
```
The following services are now available:
- **SuiteServe** &mdash; [https://localhost:8080](https://localhost:8080)
  - Serves the UI and test reporting API.
  - Does not hot-reload on code changes.
  - Code changes in `ui/` will not be seen until `ui/dist/` is built.
- **Webpack DevServer** &mdash; [https://localhost:8081](https://localhost:8081)
  - Serves the UI and forwards non-UI requests to [localhost:8080](https://localhost:8080).
  - Hot-reloads on code changes in `ui/`.
  - Does not require `ui/dist/` to exist.
  - Useful for UI development, but not for production.
- **RethinkDB Administration Console** &mdash; [http://localhost:8082](http://localhost:8082)

As needed, build or rebuild `ui/dist/` with:
```bash
$ rm -rf ui/dist
$ make ui/dist
```

### Build for Production
Run `make` to build the SuiteServe binary file named "suiteserve".

### Build for Docker
Build with `docker build -t suiteserve .` and run with `docker run -v $(pwd)/tls:/app/tls suiteserve`.
