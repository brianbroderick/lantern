package main

type queryDetails struct {
	fragment string
	// "columns" is a map of column names that includes a map of all the values found
	// in queries matching the fragment.
	// for each distinct value found, totalCount and totalDuration are summed up.
	columns map[string]map[string]column
}

type column struct {
	totalCount    int32
	totalDuration float64
}
