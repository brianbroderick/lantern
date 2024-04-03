package repo

// This function is used to extract data from queries
// and store them in the database
// func (q *Queries) Extract() {
// 	if len(q.Queries) == 0 {
// 		return
// 	}

// 	for _, query := range q.Queries {
// 		for _, stmt := range query.Statements {
// 			env := object.NewEnvironment()
// 			r := extractor.NewExtractor(&stmt, true)
// 			r.Extract(*r.Ast, env)

// 			output := stmt.String(true) // maskParams = true, i.e. replace all values with ?

// 			q.addQuery(databases, database, source, input, output, duration, stmt.Command())
// 		}
// 	}
// }
