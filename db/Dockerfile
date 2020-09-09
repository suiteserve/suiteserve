FROM mongo:4.4
COPY mongod.conf /etc/
COPY init/*.js /docker-entrypoint-initdb.d/
CMD ["-f", "/etc/mongod.conf"]
