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

Show creates by object type:

```
select cs.object_type,
       count(1) as unique_objects,
       sum(total_count) as total_count,
       sum(total_duration_us) as total_duration_us
from queries q
join create_statements_in_queries csq on q.uid = csq.query_uid
join create_statements cs on csq.create_statement_uid = cs.uid
group by cs.object_type
order by cs.object_type;
```

Show temp tables and indexes by name, removing numbers because a lot of people add a random number to the name:

```
select cs.object_type, regexp_replace(cs.name, '[0-9]', '', 'g'),
       count(1) as unique_objects,
       sum(total_count) as total_count,
       sum(total_duration_us) as total_duration_us
from queries q
join create_statements_in_queries csq on q.uid = csq.query_uid
join create_statements cs on csq.create_statement_uid = cs.uid
group by cs.object_type, regexp_replace(cs.name, '[0-9]', '', 'g')
order by total_count desc;
```

Show temp tables and indexes - per unique query:

```
select cs.object_type, cs.name, q.masked_query, total_count, total_duration_us
from queries q
join create_statements_in_queries csq on q.uid = csq.query_uid
join create_statements cs on csq.create_statement_uid = cs.uid
order by
    -- total_duration_us desc,
    total_count desc;
```

Show queries by username:

```
select user_name, count(1) as uniq_queries, sum(total_count) as total_count, trunc(sum(total_duration_us)/1000000, 3) as total_duration_sec
from query_users group by user_name order by total_count desc
```

Show slowest queries for a specific username. Note that we're pulling the totals from the query_users table since multiple users can run the same query:

```
select q.masked_query, qu.total_count, trunc(qu.total_duration_us::decimal/1000000,3) as total_duration_sec, trunc(q.average_duration_us::decimal/1000000,3) as average_duration_sec
from queries q
join query_users qu on q.uid = qu.query_uid where qu.user_name = 'queue_executor' order by q.total_duration_us desc;
```