# Repo

This package aggregates queries and stores them into a PG database. 


## Helpful DB Queries

Show all queries by database ordered by the most queried.

```
select d.name as db_name, sum(total_count) as tc, sum(total_duration) as td
from databases d join queries q ON d.uid = q.database_uid
group by d.uid
order by tc desc;
```

Show all queries by query ordered by the most queries in a specific database.

```
select masked_query, sum(total_count) as tc, sum(total_duration) as td
from databases d join queries q ON d.uid = q.database_uid
where d.name = 'datname'
group by q.uid
order by tc desc;
```

Show totals per command tag:

```
select
    command,
    count(1) as uniq_queries,
    sum(total_count) as total_count,
    trunc(sum(total_duration_us)/1000000, 3) as total_duration_sec,
    trunc((sum(total_duration_us)/sum(total_count))/1000000,3) as average_duration_sec
from queries
where command in ('ANALYZE', 'COMMIT', 'CREATE', 'DELETE', 'DROP', 'INSERT', 'ROLLBACK', 'SELECT', 'SET', 'UPDATE', 'WITH')
group by command
order by total_duration_sec desc;
```