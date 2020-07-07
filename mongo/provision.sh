#!/usr/bin/env bash

set -e
echo "Bringing up MongoDB..."
docker-compose up -d mongo
echo "Waiting 10 seconds for MongoDB..."
sleep 10
echo "Provisioning replica set..."
docker-compose exec mongo bash -c "mongo /data/provision_rs.js"
echo "Creating admin user..."
docker-compose exec mongo bash -c "mongo admin /data/provision_admin.js"
echo "Creating suiteserve user..."
docker-compose exec mongo bash -c "mongo -u admin -p admin --authenticationDatabase admin admin /data/provision_app.js"
echo "Done provisioning MongoDb!"
echo "Bringing up everything..."
docker-compose up
