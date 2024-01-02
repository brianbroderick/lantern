package counter

import (
	"fmt"
	"sort"

	"github.com/brianbroderick/lantern/pkg/sql/token"
)

func NewStats() *QueryStats {
	return &QueryStats{
		ByCount:              make([]*Query, 0),
		ByDuration:           make([]*Query, 0),
		SumCommandByCount:    map[token.TokenType]int64{},
		SumCommandByDuration: map[token.TokenType]int64{},
	}
}

func PopulateStats(stats *QueryStats, queries *Queries) {
	for _, query := range queries.Queries {
		stats.AddQuery(query)
	}
}

func (q *QueryStats) AddQuery(query *Query) {
	q.ByCount = append(q.ByCount, query)
	q.ByDuration = append(q.ByDuration, query)

	// Get summation stats by command
	if _, ok := q.SumCommandByCount[query.Command]; !ok {
		q.SumCommandByCount[query.Command] = 0
	}
	q.SumCommandByCount[query.Command] += query.TotalCount

	if _, ok := q.SumCommandByDuration[query.Command]; !ok {
		q.SumCommandByDuration[query.Command] = 0
	}
	q.SumCommandByDuration[query.Command] += query.TotalDuration
}

func (q *QueryStats) SortByCount() {
	sort.Slice(q.ByCount, func(i, j int) bool {
		return q.ByCount[i].TotalCount > q.ByCount[j].TotalCount
	})
}

func (q *QueryStats) SortByDuration() {
	sort.Slice(q.ByDuration, func(i, j int) bool {
		return q.ByDuration[i].TotalDuration > q.ByDuration[j].TotalDuration
	})
}

func (q *QueryStats) Stats() []string {
	stats := make([]string, 0)

	stats = append(stats, fmt.Sprintf("Number of unique queries: %d", len(q.ByCount)))

	return stats
}
