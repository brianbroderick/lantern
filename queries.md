# These are some interesting queries to run once there's some data. The returned values are representing a specific test run to check for anomolies.

```
select count(1) from tables_in_queries;
-- 19462

select count(1) from tables t join tables_in_queries q on t.uid = q.table_uid;
-- 19462

select count(1) from queries;
-- 6955

select count(distinct query_uid) from tables_in_queries;
-- 5480

select count(1) from tables_in_queries t left join queries q on q.uid = t.query_uid;
-- 19462

select count(1) from queries q left join tables_in_queries t on q.uid = t.query_uid;
-- 20937

select count(1) from queries q left join tables_in_queries t on q.uid = t.query_uid where t.query_uid is null;
-- 1475

select * from queries q left join tables_in_queries t on q.uid = t.query_uid where t.query_uid is null order by total_duration desc;

select count(1) from columns_in_queries;
-- 14584

select count(1) from table_joins_in_queries;
-- 27192

select count(1) from databases;
-- 13

select count(1) from sources;
-- 1074

select count(1) from tables;
-- 2107

select count(distinct table_uid) from tables_in_queries;
-- 2107

select tiq.table_name, ciq.table_name
from tables_in_queries tiq
         join columns_in_queries ciq on tiq.table_uid = ciq.table_uid and tiq.table_name != ciq.table_name;

select tjiq.table_a, tiq.table_name from table_joins_in_queries tjiq
    join tables_in_queries tiq on tiq.table_uid = tjiq.table_uid_a;

select tjiq.schema_a, tjiq.table_b, tiq.table_name from table_joins_in_queries tjiq
    join tables_in_queries tiq on tiq.table_uid = tjiq.table_uid_b;

-- shows tables appearing in the most unique queries
select table_uid, max(schema_name) as sn, max(table_name) as tn, count(1) as counter  from tables_in_queries group by table_uid order by counter desc;

-- shows tables appearing in the most queries.
-- note the total_duration will become accurate when PG query logs are parsed 
-- because we'll know specific times for each query rather than the full transaction. 

select t.uid, max(t.schema_name) as schema_name, max(t.table_name) as table_name, sum(q.total_count) as total_count, sum(q.total_duration) as total_duration
from tables_in_queries tq
         join tables t on t.uid = tq.table_uid
         join queries q on tq.query_uid = q.uid
group by t.uid
order by total_count desc;
-- 2107 rows

-- get a specific value from the above query to check it's returning the right values
select q.masked_query, q.total_count, t.table_name from tables_in_queries tq
         join tables t on t.uid = tq.table_uid
         join queries q on tq.query_uid = q.uid
where t.uid = '?'
```


```
Sequence showing temp table usage:

select count(1) from tables where table_name like 'temp%'

select * from tables where table_name like 'temp_property_prefs_%'

select regexp_replace(table_name, '[0-9]', '', 'g') as name, count(1) as counter
from tables
where table_name like 'temp%'
group by regexp_replace(table_name, '[0-9]', '', 'g')
order by counter desc

select q.uid,count(1) as counter from tables_in_queries tq  join queries q on q.uid = tq.query_uid
                                     where table_name like 'temp_property_prefs_%' group by q.uid order by counter desc

select max(q.masked_query), q.uid,count(1) as counter from tables_in_queries tq  join queries q on q.uid = tq.query_uid
                                     where table_name like 'temp_property_prefs_%' group by q.uid order by counter desc

select * from queries where uid = '47b64ca7-dc46-5829-b6ed-0bbc4fc465bc'

select max(name), max(url), s.uid, count(1) as counter
from tables_in_queries tq
         join queries q on q.uid = tq.query_uid
         join sources s on s.uid = q.source_uid
where table_name like 'temp_property_prefs_%'
group by s.uid
order by counter desc;

select * from sources




select * from sources where uid = '46178a2f-823f-5200-aa03-8913e725cc1b'
```

Sequence Showing JSONB usage:
```
select cq.clause, count(1) as counter from columns_in_queries cq where cq.column_name = 'details' group by cq.clause order by cq.clause;


select * from queries

select q.masked_query, cq.clause from columns_in_queries cq
            join queries q on q.uid = cq.query_uid
            where cq.column_name = 'details'
group by cq.clause order by cq.clause;

select count(1) from table_joins_in_queries where on_condition ilike '%details%'

select * from table_joins_in_queries where on_condition ilike '%details%'
````