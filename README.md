# Lantern

## Shipping Postgres query logs to Elasticsearch for near realtime query analysis.

![Lantern](https://user-images.githubusercontent.com/7585181/80007270-43b8fb00-8483-11ea-996f-275529aa3863.png)

This reads Postgres logs from Redis, normalizes the query, and then pushes it into Elasticsearch.

The normalization function is from the excellent [pg_query_go](https://github.com/brianbroderick/pg_query_cli/tree/master/vendor/github.com/lfittl/pg_query_go) Go package

Assuming that your Go workspace is set up, run `go build`. If not, first visit [Golang](https://golang.org/) to set it up.
Building the file may take a couple of minutes because it's compiling parts of the PG lib.

## You'll be able to see

* Frequently run queries
* Slowest queries
* Line of code that ran the query
* Which client ran the query
* Filter/Group by day of week

## How it works

It ships PG query logs to Redis using the Redislog extension. PG log shipper takes these messages from Redis, aggregates them to a 1 minute grain, and then bulk inserts them into Elasticsearch. You can then build whatever visualisations and dashboards that you want using Kibana or anther tool that can query ES.

## Not compatible with

* AWS' RDS
* AWS' Aurora
* Most other managed PG services

You must be able to compile and install the Redislog extension. As of this writing, it's not available on AWS' managed services.

## Config

It's easy to config. You will need the [Redislog extension](https://github.com/2ndquadrant-it/redislog) compiled into your PG install. Assuming that's the case, you'll add these lines to the postgresql.conf:

```
shared_preload_libraries = 'redislog'
redislog.hosts = 'your.redis.host'
redislog.key = 'your-app-master'
```

For Lantern, you'll need these to either set environment variables or flags to connect to Elasticsearch and Redis.

The env vars are:

```
PLS_REDIS_URL="your.redis.host"    # Default: "127.0.0.1:6379"
PLS_REDIS_QUEUES="your-app-master" # comma separated whitelist of redis keys being published to
PLS_REDIS_STATS_QUEUES="stats"     # 
PLS_ELASTIC_URL="your.es.server"   # Default: "http://127.0.0.1:9200"
PLS_REDIS_PASSWORD="your.password" # (Optional)
PLS_ELASTIC_SNIFF=boolean          # (Optional) it sets elastic.SetSniff. This is may need to be set to false if ES is on a                                         different network, such as within Docker.
```

The flags are:

```
  -elasticUrl string
    	Elasticsearch URL. Can also set via PLS_ELASTIC_URL env var
  -queues string
    	comma separated list of queues that overrides env vars. Can also be set via PLS_REDIS_QUEUES env var
	-statsQueues string
	    comma separated list of queues handling the stats payload. The stats payload doesn't come from RedisLog. It can be used
			by other systems that mimic the RedisLog payload. 
  -redisPassword string
    	Redis password (optional). Can also set via PLS_REDIS_PASSWORD env var
  -redisUrl string
    	Redis URL. Can also set via PLS_REDIS_URL env var
```

What I've done is set the Redis queue to whatever the server is. For example, we have a lot of servers based on planets, so I have them set like this saturn-master, saturn-follower, pluto-master, pluto-follower, etc.  The queue name is sent with the payload, so you can use it to filter or group your data.

## 1 minute grain

Lantern aggregates normalized queries to a 1 minute interval. This

* Limits ES' storage requirements
* Speeds up ES' queries

* Only the first params are stored per query, per minute
* Up to 2 minute latency

In this context, a normalized query means that it converts `SELECT * FROM table WHERE id = 42;` to `SELECT * FROM table WHERE id = ?;` All string and number variables are converted to a question mark. It then hashes the text and uses that to determine uniqueness. 

## Local Dependencies
* `docker-compose up -d`

## Testing

```
# search elasticsearch
curl http://localhost:9200/pg-$(date +"%Y-%m-%d")/_search 

# delete data in elasticsearch
curl -XDELETE localhost:9200/pg-$(date +"%Y-%m-%d") 
```

## Kibana

Once the data is in Elasticsearch, you can create all the charts you'd like such as queries by total duration or total count.

![Total Duration](https://user-images.githubusercontent.com/7585181/80007253-3d2a8380-8483-11ea-9f77-93e2813c3b70.png)

![Total Count](https://user-images.githubusercontent.com/7585181/80007228-36037580-8483-11ea-8225-29507c9b32db.png)

## Kubernetes

If you deploy this in K8s and are using 1 core or less, consider setting GOMAXPROCS to a low number like 1 or 2. 
If it's set too high, K8s throttles the process when it exceeds the CPU you've allotted it. This can cause 100ms pauses, which will significantly slow down the worker.

To set GOMAXPROCS, you simply need to set it as an environment variable.

## Benchmark

On an AWS C5.large, which has 2 virtual cores and 4GB of memory, it's able to process a little more than 10k queries per second. The process is CPU intensive and uses minimal memory, so the compute optimized servers are the best.

## Scaling Caveats

Each Redis key you're consuming from the environment variable PLS_REDIS_QUEUES spawns a separate Goroutine, which are light weight threads. For this reason, you'll get the most performance if you're consuming from more than 1 key per worker. 

It's good practice to use a different key per PG server. For example `my-application-prod` and `my-application-follower-1`.  

Workers don't know of the existence of other workers. Therefore, if multiple workers consume from the same Redis key, it will result in additional storage requirements on your Elasticsearch cluster. This is because each worker will insert a document for each normalized query it's exposed to, even if another worker is inserting data for the same normalized query.

All of the stats will still be correct because `total_count`, and `total_duration` are calculated based on the queries that particular worker consumed assuming your Kibana graphs are calculating the sum. 

The `avg_duration` key will be problematic if you're consuming the same key from multiple workers. This is because one worker will have consumed a larger percentage of a particular query. You _could_ do an average of this key, but that would give both messages equal weight, which wouldn't be true. In this case, it would be more accurate to do a calculated field in ES with the formula: `total_duration / total_count`. 

To optimize for performance, have enough workers that each one can consume 2 or more different keys, and don't share keys across workers. 

To optimize for uptime, you can duplicate keys across workers; however, as noted, this will increase the storage requirements on your ES cluster. 

## Authors

- [Brian Broderick](https://github.com/brianbroderick)

## Why name it Lantern?

Lanterns light up dark places, and PG query logs are dark places indeed. Hopefully with the help of this tool, you'll be able to shine a light on your database performance problems.

## License

Copyright (c) 2017, [Brian Broderick](https://github.com/brianbroderick)<br>
Lantern is licensed under the 3-clause BSD license, see LICENSE file for details.

This project includes code derived from the [PostgreSQL project](http://www.postgresql.org/) and [pg_query_go](https://github.com/brianbroderick/pg_query_cli/tree/master/vendor/github.com/lfittl/pg_query_go)
see LICENSE.POSTGRESQL and LICENSE.PG_QUERY_GO respectively for details. 
