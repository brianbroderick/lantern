package workflow

type Workflow struct {
	Env         string       `json:"env"`
	StartedAt   string       `json:"started_at"`
	PageObjects []PageObject `json:"page_objects"`
}

type PageObject struct {
	PageObjectName string  `json:"page_object_name"`
	URL            string  `json:"url"`
	DateTime       string  `json:"date_time"`
	Modals         []Modal `json:"modals,omitempty"`
}

type Modal struct {
	URL                   string  `json:"url"`
	ExecutionTime         string  `json:"execution_time"`
	QueryTime             string  `json:"query_time"`
	TotalQueries          string  `json:"total_queries"`
	DbQueries             string  `json:"db_queries"`
	PercentQuery          string  `json:"percent_query"`
	RowCount              string  `json:"row_count"`
	MemoryUsage           string  `json:"memory_usage"`
	PeakMemoryUsage       string  `json:"peak_memory_usage"`
	PageSize              string  `json:"page_size"`
	ResourceVersionNumber string  `json:"resource_version_number"`
	TrackName             string  `json:"track_name"`
	Server                string  `json:"server"`
	HTTPXForwardedFor     string  `json:"http_x_forwarded_for"`
	RemoteAddress         string  `json:"remote_address"`
	Queries               []Query `json:"queries"`
}

type Query struct {
	Db            string `json:"db"`
	QueryTime     string `json:"query_time"`
	ExecutionTime string `json:"execution_time"`
	RowCount      string `json:"row_count"`
	MemoryUsage   string `json:"memory_usage"`
	Query         string `json:"query"`
}
