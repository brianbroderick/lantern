These files represent a SQL parser. Not all keywords are support yet as this is work in progress. 

PG Tokens: https://www.postgresql.org/docs/13/sql-keywords-appendix.html

TODO: 

VALUES: https://www.postgresql.org/docs/current/queries-values.html
"values" provides a way to generate a constant table, with an example being used in a CTE

Check this one: 
select count(*) as total_rows, count(*) filter( where id is not null) as id_set_count from table;