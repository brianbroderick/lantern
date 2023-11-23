These files represent a SQL parser. Not all keywords are support yet as this is work in progress. 

PG Tokens: https://www.postgresql.org/docs/13/sql-keywords-appendix.html

TODO: 

Lexer Position is showing for peekTokenFour, not curToken (or maybe it needed to be peekToken)

VALUES: https://www.postgresql.org/docs/current/queries-values.html
"values" provides a way to generate a constant table, with an example being used in a CTE

Check this one: 
select count(*) as total_rows, count(*) filter( where id is not null) as id_set_count from table;


-------

Parse Errors:

"no prefix parse function for DOT found at line 0 char 25"
{"select u . id from users u", 2, "(SELECT u.id FROM users);"},