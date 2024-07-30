package parser

import (
	"testing"

	"github.com/brianbroderick/lantern/internal/postgresql/lexer"
	"github.com/stretchr/testify/assert"
)

// func TestParser(t *testing.T) {
// 	s := "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@testdb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo where bar = $1"
// 	l := lexer.New(s)
// 	p := New(l)
// 	program := p.ParseProgram()
// 	checkParserErrors(t, p)

// 	assert.Equal(t, 1, len(program.Statements))
// 	assert.Equal(t, "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@testdb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo where bar = $1", program.Statements[0].String())
// }

func TestParserStatements(t *testing.T) {
	var tests = []struct {
		str           string
		result        string
		lenStatements int
	}{
		// Single line log entry
		{
			str:           "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from users where id = $1",
			result:        "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from users where id = $1",
			lenStatements: 1,
		},
		// Multiple line log entry
		{
			str:           "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo\n where bar = $1",
			result:        "2023-07-10 09:52:46 MDT:127.0.0.1(50032):postgres@sampledb:[24649]:LOG:  duration: 0.059 ms  execute <unnamed>: select * from foo\n where bar = $1",
			lenStatements: 1,
		},
		// Multiple log entries
		{
			str:           "2023-01-02 01:02:03 MDT:127.0.0.1(12345):postgres@sampledb:[23456]:LOG:  duration: 0.123 ms  execute <unnamed>: select * from foo where bar = $1\n2023-07-10 10:11:12 MDT:127.0.0.1(56789):postgres@sampledb:[98765]:LOG:  duration: 0.159 ms  execute <unnamed>: select * from bar where baz = $1",
			result:        "2023-01-02 01:02:03 MDT:127.0.0.1(12345):postgres@sampledb:[23456]:LOG:  duration: 0.123 ms  execute <unnamed>: select * from foo where bar = $1\n",
			lenStatements: 2,
		},
		// Multiple log entries on multiple lines
		{
			str:           "2023-01-02 01:02:03 MDT:127.0.0.1(12345):postgres@sampledb:[23456]:LOG:  duration: 0.123 ms  execute <unnamed>: select * \n  from multi\n  where \nline = $1\n2023-07-10 10:11:12 MDT:127.0.0.1(56789):postgres@sampledb:[98765]:LOG:  duration: 0.159 ms  execute <unnamed>: select * from\n second where id = $1",
			result:        "2023-01-02 01:02:03 MDT:127.0.0.1(12345):postgres@sampledb:[23456]:LOG:  duration: 0.123 ms  execute <unnamed>: select * \n from multi\n where \n line = $1\n",
			lenStatements: 2,
		},
		// From RDS Log
		{
			str:           "2024-07-10 17:48:11 UTC:127.0.0.1(42200):my_app@my_db:[46542]:LOG:  duration: 0.212 ms  statement: SET STATEMENT_TIMEOUT = '360s';",
			result:        "2024-07-10 17:48:11 UTC:127.0.0.1(42200):my_app@my_db:[46542]:LOG:  duration: 0.212 ms  statement: SET STATEMENT_TIMEOUT = '360s';",
			lenStatements: 1,
		},
		{
			str: `2024-07-10 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.031 ms  parse <unnamed>: SHOW TRANSACTION ISOLATION LEVEL
				2024-07-10 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.006 ms  bind <unnamed>: SHOW TRANSACTION ISOLATION LEVEL
				2024-07-10 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.015 ms  execute <unnamed>: SHOW TRANSACTION ISOLATION LEVEL`,
			result:        "2024-07-10 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.031 ms  parse <unnamed>: SHOW TRANSACTION ISOLATION LEVEL\n",
			lenStatements: 3,
		},
		// From RDS Log on multiple lines.
		{
			str: `2024-07-10 17:48:11 UTC:127.0.0.1(42200):my_app@my_db:[46542]:LOG:  duration: 0.212 ms  statement: SELECT
				                                                        c.id AS id,
				                                                        c.name AS name
				                                                FROM
				                                                        user_groups ug
				                                                        JOIN companies c ON ( c.id = ug.group_id )
				                                                WHERE
				                                                        ug.user_id = 204782`,
			result:        "2024-07-10 17:48:11 UTC:127.0.0.1(42200):my_app@my_db:[46542]:LOG:  duration: 0.212 ms  statement: SELECT c.id AS id,\n c.name AS name\n FROM \n user_groups ug\n JOIN companies c ON ( c.id = ug.group_id )\n WHERE \n ug.user_id = 204782",
			lenStatements: 1,
		},
		// From RDS Log on multiple lines.
		{
			str: `2024-07-10 17:48:14 UTC:127.0.0.1(42200):my_app@my_db:[46542]:LOG:  duration: 0.212 ms  statement: SELECT
				                                                        c.id AS id,
				                                                        c.name AS name
				                                                FROM
				                                                        user_groups ug
				                                                        JOIN companies c ON ( c.id = ug.group_id )
				                                                WHERE
				                                                        ug.user_id = 204782
				2024-07-10 17:48:14 UTC:127.0.0.1(42200):my_app@my_db:[46542]:LOG:  duration: 0.212 ms  statement: SELECT
				                                                        c.id AS id,
				                                                        c.name AS name
				                                                FROM
				                                                        user_groups ug
				                                                        JOIN companies c ON ( c.id = ug.group_id )
				                                                WHERE
				                                                        ug.user_id = 204782`,
			result:        "2024-07-10 17:48:14 UTC:127.0.0.1(42200):my_app@my_db:[46542]:LOG:  duration: 0.212 ms  statement: SELECT c.id AS id,\n c.name AS name\n FROM \n user_groups ug\n JOIN companies c ON ( c.id = ug.group_id )\n WHERE \n ug.user_id = 204782\n",
			lenStatements: 2,
		},
		{
			str: `2024-07-10 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  bind R42: select * from users
		2024-07-10 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  execute R42: select * from users`,
			result:        "2024-07-10 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  bind R42: select * from users\n",
			lenStatements: 2,
		},
		{
			str:           `2024-07-10 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  bind R42: BEGIN`,
			result:        "2024-07-10 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  bind R42: BEGIN",
			lenStatements: 1,
		},
		{
			str: `2024-07-04 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  execute <unnamed>: BEGIN
					2024-07-05 17:48:14 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  execute <unnamed>: COMMIT
					2024-07-05 17:48:14 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  execute <unnamed>: ROLLBACK`,
			result:        "2024-07-04 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  execute <unnamed>: BEGIN",
			lenStatements: 3,
		},
		{
			str: `2024-07-04 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  execute <unnamed>: select
					'2024-07-05 17:48:14 UTC' from users;`,
			result:        "2024-07-04 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  execute <unnamed>: select '2024-07-05 17:48:14 UTC' from users;",
			lenStatements: 1,
		},
		{
			str:           "2024-07-10 17:48:11 UTC:10.1.1.1(48684):pp@mydb:[40113]:LOG:  duration: 1.410 ms  statement: SET STATEMENT_TIMEOUT = '360s'; /*some comments*/SELECT * from users",
			result:        "2024-07-10 17:48:11 UTC:10.1.1.1(48684):pp@mydb:[40113]:LOG:  duration: 1.410 ms  statement: SET STATEMENT_TIMEOUT = '360s'; /*some comments*/SELECT * from users",
			lenStatements: 1,
		},
		{
			str:           "2024-07-10 17:48:11 UTC:10.1.1.1(48684):pp@mydb:[40113]:LOG:  duration: 1.410 ms",
			result:        "2024-07-10 17:48:11 UTC:10.1.1.1(48684):pp@mydb:[40113]:LOG:  duration: 1.410 ms",
			lenStatements: 1,
		},
		{
			str: `2024-07-10 17:48:11 UTC:10.1.1.1(48684):pp@mydb:[40113]:LOG:  duration: 1.410 ms
					2024-07-04 17:48:11 UTC:10.1.1.1(51010):sys_user@my_db_01_11866:[46031]:LOG:  duration: 0.004 ms  execute <unnamed>: BEGIN`,
			result:        "2024-07-10 17:48:11 UTC:10.1.1.1(48684):pp@mydb:[40113]:LOG:  duration: 1.410 ms",
			lenStatements: 2,
		},
		{
			str:           `2024-07-09 15:22:18 MDT:::1(63248):postgres@lantern:[29550]:ERROR:  syntax error at or near "limit" at character 429`,
			result:        `2024-07-09 15:22:18 MDT:::1(63248):postgres@lantern:[29550]:ERROR:  syntax error at or near "limit" at character 429`,
			lenStatements: 1,
		},
		{
			str: `2024-07-10 17:48:27 UTC:10.1.1.1(38502):postgres@lantern:[34633]:LOG:  duration: 0.398 ms  statement: DISCARD ALL;
		2024-07-10 17:48:11 UTC:10.1.1.1(45924):postgres@lantern:[43863]:LOG:  duration: 0.278 ms  statement: SET STATEMENT_TIMEOUT = '360s';`,
			result:        "2024-07-10 17:48:27 UTC:10.1.1.1(38502):postgres@lantern:[34633]:LOG:  duration: 0.398 ms  statement: DISCARD ALL; \n",
			lenStatements: 2,
		},
		{
			str:           `2024-07-10 17:48:42 UTC:10.1.1.1(38502):postgres@lantern:[34633]:LOG:  duration: 0.398 ms  statement: SET STATEMENT_TIMEOUT = '360s'; /*{"id":15514,"team_name":"my_team"}*/DROP TABLE IF EXISTS my_tmp_tbl;`,
			result:        `2024-07-10 17:48:42 UTC:10.1.1.1(38502):postgres@lantern:[34633]:LOG:  duration: 0.398 ms  statement: SET STATEMENT_TIMEOUT = '360s'; /*{"id":15514,"team_name":"my_team"}*/DROP TABLE IF EXISTS my_tmp_tbl;`,
			lenStatements: 1,
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.str)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		assert.Equal(t, tt.lenStatements, len(program.Statements))
		assert.Equal(t, tt.result, program.Statements[0].String())
	}

}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
