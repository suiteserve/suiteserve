Run `./provision.sh` from within this directory to bring up and provision a single-instanced MongoDB replica set with Mongo Express. Access Mongo Express at [localhost:8081](http://localhost:8081). To bring up the services without provisioning them, use the standard `docker-compose up` command and friends.

Created replica sets:
- *ID:* `rs0`  
  *Members:*
  - `localhost:27017`

Created users:
- *User:* `admin`  
  *Password:* `admin`  
  *Roles:* `root`  
  *Auth DB:* `admin`
- *User:* `suiteserve`  
  *Password:* `suiteserve`  
  *Roles:* `readWrite` on DB `suiteserve`  
  *Auth DB:* `admin`

⚠️ *Warning:* This configuration of MongoDB must only be used for development purposes. Nothing here is secure or production-ready.
