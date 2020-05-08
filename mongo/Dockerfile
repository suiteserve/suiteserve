FROM mongo:4.2.5
USER mongodb
COPY mongod.conf /etc/
COPY provision_*.js /data/
CMD ["--config", "/etc/mongod.conf"]