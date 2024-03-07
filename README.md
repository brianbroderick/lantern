# Lantern

The original version of Lantern aggregated and exported Postgres query logs to Elasticsearch. This code has moved to the `deprecated` folder and is only for reference. If you’re using this code, create an issue, and I’ll ensure all the import paths are correct and able to build. 

The original concept depends on the Redislog extension installed on Postgres. However, managed Postgres, such as AWS RDS, does not allow the installation of custom extensions. 

I used Redislog because it outputs logs as JSON, which should be a native option in Postgres 15. In the meantime, Lantern will include a custom log parser.  

Lantern has been an invaluable tool for me over the past six years to identify queries that are run the most and are the slowest. PGBadger also addresses this concern; however, it’s a manual and time-consuming process. Nevertheless, using Lantern for the past several years has shed light on some of its limitations, which I hope to address.

## Work In Progress

I’m reimagining Lantern as a tool to be more broadly used by DBAs and engineers with a mind for performance and optimization. Since most people use RDS or other managed versions of Postgres, I’m removing the Redislog dependency and will parse native log files instead. 

While the original version focused on counting queries and their durations, I will extend it to look for other poor data patterns, such as the overuse of JSONB columns.

If you have ideas for things you’d like to see, post a GitHub issue, and we can discuss it. 

## DB Migrations:

We're using https://github.com/golang-migrate/migrate to manage DB migrations, which are located in `/pkg/repo/migrations`

To create a new migration run: `migrate create -ext sql -dir pkg/repo/migrations -seq <migration_name>`
To run migrations: `migrate -database ${POSTGRESQL_URL} -path pkg/repo/migrations up`
To rollback: `migrate -database ${POSTGRESQL_URL} -path pkg/repo/migrations down`

PG tutorial: https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md
Installing: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate
