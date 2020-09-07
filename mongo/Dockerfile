FROM mongo:4.4
COPY mongod.conf /etc/
COPY *.js *.sh /docker-entrypoint-initdb.d/
CMD ["-f", "/etc/mongod.conf"]
