# WB-TASK1
The application is a service for processing orders received from NATS

## Principle of operation
In my case NATS works with docker image.
Data received via NATS-streaming in json format is divided into several tables in the DB and simultaneously stored in CACHE.
To speed up the application's work with queries, when the application starts, all DB data is loaded into the cache.
Now, to obtain order data, it will come not from the DB, but from CACHE

## Getting Started
NATS and DB connection parameters are located in the [.env] file
The server is running at http://localhost:3002. If desired, you can change the connection settings in the [config.yml] file