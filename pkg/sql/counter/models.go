package counter

type Queries struct {
	Queries map[string]*Query // the key is the sha of the query
}

type Query struct {
	Sha           string // unique sha of the query
	OriginalQuery string // the original query
	MaskedQuery   string // the query with parameters masked
	TotalCount    int64  // the number of times the query was executed
	TotalDuration int64  // the total duration of all executions of the query in microseconds
}

type QueryStats struct {
	ByCount    []*Query
	ByDuration []*Query
}
