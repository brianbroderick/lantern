# Lantern

The original version shipped Postgres query logs to Elasticsearch. All of this code has moved to the v0 folder, which is a complete runable program.

The problem with the original concept for Lantern is that it depends on the Redislog extension being installed on Postgres. Since most people use managed Postgres, such as AWS RDS, this isn't an option.

Lantern V1 will be reimagined to do more query analysis, but it will be designed differently to accomodate managed Postgres hosting. 