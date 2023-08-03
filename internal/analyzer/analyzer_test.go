package analyzer

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

// func TestParseToJSON(t *testing.T) {
// 	json := ParseToJSON("select u.id, u.name from users u inner join addresses a on a.user_id = u.id where id = 42;")
// 	assert.NotEmpty(t, json)

// 	fmt.Println(json)
// }

// func TestParseToJSON(t *testing.T) {
// 	json := ParseToJSON("select u.id, u.name from users where id = 42;")
// 	assert.NotEmpty(t, json)

// 	fmt.Println(json)
// }

func TestParseToJSON(t *testing.T) {
	json := ParseToJSON("select u.id, u.name from users u inner join addresses a on a.user_id = u.id inner join phone_numbers p on p.user_id = u.id where id = 42;")
	assert.NotEmpty(t, json)

	fmt.Println(json)
}

// func TestParse(t *testing.T) {
// 	p := Parse("select 42;")
// 	assert.NotNil(t, p)

// 	fmt.Printf("Parse: %+v\n", p)
// }

// func TestWalk(t *testing.T) {
// 	stmts := Walk("select id, email, 1 + 2 as math from users where id = 42 and email = 'joe@example.com' and foo like concat('1','2');")
// 	assert.NotNil(t, stmts)
// 	for _, s := range stmts {
// 		for _, c := range s.Columns {
// 			fmt.Printf("col: %+v\n", c)
// 		}
// 		for _, t := range s.Tables {
// 			fmt.Printf("table: %+v\n", t)
// 		}
// 	}
// }

// func TestWalk(t *testing.T) {
// 	stmts := Walk("select u.id from users u inner join addresses a on a.user_id = u.id where id = 42;")
// 	assert.NotNil(t, stmts)
// 	for _, s := range stmts {
// 		for _, c := range s.Columns {
// 			fmt.Printf("col: %+v\n", c)
// 		}
// 		// for _, t := range s.Tables {
// 		// 	fmt.Printf("table: %+v\n", t)
// 		// }
// 	}
// }

// func TestJSON(t *testing.T) {
// 	stmts := Parse("select id, email as email_address, 1 + 2 as math from users where id = 42 and email = 'joe@example.com' and foo like concat('1','2');")
// 	fmt.Printf("%+v\n", stmts)
// }

// stmts := Parse("select id, email, 1 + 2 as math from users where id = 42 and email = 'joe@example.com' and foo like concat('1','2');")
// fmt.Printf("stmt: %+v\n", stmts.Stmts)

// stmt:
//   [
// 		stmt:
// 		  {select_stmt:
// 				{
// 				target_list:
// 					{res_target:
// 						{val:
// 							{column_ref:{fields:{string:{sval:"id"}}  location:7}}
// 							location:7
// 						}
// 					}
// 				target_list:
// 				  {res_target:
// 						{val:
// 							{column_ref:{fields:{string:{sval:"email"}}  location:11}}
// 							location:11
// 						}
// 					}
// 				target_list:
// 				  {res_target:
// 						{name:"math" val: {
// 						  a_expr:{kind:AEXPR_OP  name:{string:{sval:"+"}}
// 							lexpr:{a_const:{ival:{ival:1}  location:18}}
// 							rexpr:{a_const:{ival:{ival:2}  location:22}}
// 							location:20}}  location:18}}
// 				from_clause:
// 				  {range_var:{relname:"users"  inh:true  relpersistence:"p"  location:37}}
// 				where_clause:
// 				  {bool_expr:{
// 						boolop:AND_EXPR
// 						args:{
// 							a_expr:{kind:AEXPR_OP  name:{string:{sval:"="}}
// 						  lexpr:{column_ref:{fields:{string:{sval:"id"}}  location:49}}
// 						  rexpr:{a_const:{ival:{ival:42}  location:54}}  location:52}
// 						}
// 						args:{
// 							a_expr:{kind:AEXPR_OP  name:{string:{sval:"="}}
// 							lexpr:{column_ref:{fields:{string:{sval:"email"}}  location:61}}
// 							rexpr:{a_const:{sval:{sval:"joe@example.com"}  location:69}}  location:67}
// 						}
// 						args:{
// 							a_expr:{kind:AEXPR_LIKE  name:{string:{sval:"~~"}}
// 							lexpr:{column_ref:{fields:{string:{sval:"foo"}}  location:91}}
// 							rexpr:{
// 								func_call:{
// 									funcname:{string:{sval:"concat"}}
// 									args:{a_const:{sval:{sval:"1"}  location:107}}
// 									args:{a_const:{sval:{sval:"2"}  location:111}}
// 									funcformat:COERCE_EXPLICIT_CALL  location:100}} location:95}}
// 									location:57}}
// 					limit_option:LIMIT_OPTION_DEFAULT  op:SETOP_NONE}
// 			}  stmt_len:115
// 	]
