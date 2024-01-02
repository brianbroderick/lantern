package counter

import "sort"

func NewStats() *QueryStats {
	return &QueryStats{
		ByCount:    make([]*Query, 0),
		ByDuration: make([]*Query, 0),
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
