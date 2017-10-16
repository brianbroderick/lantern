## pg_log_shipper

This reads Postgres logs from Redis, normalizes the query, and then pushes it into Elasticsearch.

The normalization function is from the excellent [pg_query_go](https://github.com/brianbroderick/pg_query_cli/tree/master/vendor/github.com/lfittl/pg_query_go) Go package

Assuming that your Go workspace is set up, run `go build`. If not, first visit [Golang](https://golang.org/) to set it up.
Building the file may take a couple of minutes because it's compiling parts of the PG lib.

## Authors

- [Brian Broderick](https://github.com/brianbroderick)


## License

Copyright (c) 2017, [Brian Broderick](https://github.com/brianbroderick)<br>
pg_log_shipper is licensed under the 3-clause BSD license, see LICENSE file for details.

This project includes code derived from the [PostgreSQL project](http://www.postgresql.org/) and [pg_query_go](https://github.com/brianbroderick/pg_query_cli/tree/master/vendor/github.com/lfittl/pg_query_go)
see LICENSE.POSTGRESQL and LICENSE.PG_QUERY_GO respectively for details. 
