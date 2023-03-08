package analyzer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	os.Setenv("PLATFORM_ENV", "test")
}

// func TestParseToJSON(t *testing.T) {
// 	json := ParseToJSON("select 42;")
// 	assert.NotEmpty(t, json)

// 	fmt.Println(json)
// }

// func TestParse(t *testing.T) {
// 	p := Parse("select 42;")
// 	assert.NotNil(t, p)

// 	fmt.Printf("Parse: %+v\n", p)
// }

// func TestParseNodes(t *testing.T) {
// 	stmts := Parse("select * from users where id = ?;")

// 	// stmt:
// 	//   select_stmt:
// 	//     {
// 	// 			target_list:
// 	// 				{res_target:
// 	// 					{val:
// 	// 						{column_ref:
// 	// 							{fields: {a_star:{}} location:7}
// 	// 						} location:7
// 	// 					}
// 	// 				}
// 	// 			from_clause:
// 	// 				{range_var:
// 	// 					{relname:"users" inh:true relpersistence:"p" location:14}
// 	// 				}
// 	// 			where_clause:
// 	// 				{a_expr:
// 	// 					{
// 	// 						kind:AEXPR_OP name:{string:{sval:"="}}
// 	// 						lexpr:{column_ref:{fields:{string:{sval:"id"}} location:26}}
// 	// 						rexpr:{a_const:{ival:{ival:42} location:31}}
// 	// 						location:29
// 	// 					}
// 	// 				}
// 	// 			limit_option:LIMIT_OPTION_DEFAULT op:SETOP_NONE
// 	// 	  }

// 	for _, s := range stmts.Stmts {
// 		fmt.Printf("stmt: %+v\n", s.Stmt)
// 		fmt.Printf("location: %+v\n", s.StmtLocation)
// 		fmt.Printf("len: %+v\n", s.StmtLen)

// 	}
// }

func TestWalk(t *testing.T) {
	stmts := Walk("select id, email, 1 + 2 as math from users where id = 42 and email = 'joe@example.com' and foo like concat('1','2');")
	assert.NotNil(t, stmts)
	// fmt.Printf("stmt: %+v\n", stmts.Stmts)

}

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
