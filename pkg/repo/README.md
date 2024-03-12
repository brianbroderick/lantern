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