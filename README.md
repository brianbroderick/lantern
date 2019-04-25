## pg_log_shipper

This reads Postgres logs from Redis, normalizes the query, and then pushes it into Elasticsearch.

The normalization function is from the excellent [pg_query_go](https://github.com/brianbroderick/pg_query_cli/tree/master/vendor/github.com/lfittl/pg_query_go) Go package

Assuming that your Go workspace is set up, run `go build`. If not, first visit [Golang](https://golang.org/) to set it up.
Building the file may take a couple of minutes because it's compiling parts of the PG lib.

## Config

It's easy to config. You will need the redislog extension compiled into your PG install. Assuming that's the case, you'll add these lines to the postgresql.conf:

```
shared_preload_libraries = 'redislog'
redislog.hosts = 'your.redis.host'
redislog.key = 'your-app-master'
```

For the pg_log_shipper app, you'll need these 3 env vars set:

```
PLS_REDIS_URL='your.redis.host'
PLS_REDIS_QUEUES="your-app-master"
PLS_ELASTIC_URL="your.es.server"
```

What I've done is set the Redis queue to whatever the server is. For example, we have a lot of servers based on planets, so I have them set like this saturn-master, saturn-follower, pluto-master, pluto-follower, etc.  The queue name is sent with the payload, so you can use it to filter or group your data.


## Authors

- [Brian Broderick](https://github.com/brianbroderick)


## License

Copyright (c) 2017, [Brian Broderick](https://github.com/brianbroderick)<br>
pg_log_shipper is licensed under the 3-clause BSD license, see LICENSE file for details.

This project includes code derived from the [PostgreSQL project](http://www.postgresql.org/) and [pg_query_go](https://github.com/brianbroderick/pg_query_cli/tree/master/vendor/github.com/lfittl/pg_query_go)
see LICENSE.POSTGRESQL and LICENSE.PG_QUERY_GO respectively for details. 
