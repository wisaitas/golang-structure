package httpx

type ErrorContext struct {
	FilePath     *string `json:"filePath"`
	ErrorMessage string  `json:"errorMessage"`
}

type StandardResponse[T any] struct {
	Timestamp     string      `json:"timestamp"`
	StatusCode    int         `json:"statusCode"`
	Code          string      `json:"code"`
	Data          *T          `json:"data"`
	Pagination    *Pagination `json:"pagination,omitempty"`
	PublicMessage *string     `json:"publicMessage,omitempty"`
}

type Pagination struct {
	Page          int   `json:"page"`
	PageSize      int   `json:"pageSize"`
	HasNext       bool  `json:"hasNext"`
	HasPrev       bool  `json:"hasPrev"`
	TotalElements int   `json:"totalElements"`
	Windows       []int `json:"windows"`
}

type PaginationQuery struct {
	Page     *int `query:"page"`
	PageSize *int `query:"pageSize"`
}

type Log struct {
	TraceID    string `json:"traceId"`
	Timestamp  string `json:"timestamp"`
	DurationMs string `json:"durationMs"`

	Current *Block `json:"current"`
	Source  *Block `json:"source,omitempty"`
}

type Block struct {
	Service      string  `json:"service"`
	Method       string  `json:"method"`
	ErrorMessage *string `json:"errorMessage,omitempty"`
	Path         string  `json:"path"`
	StatusCode   string  `json:"statusCode"`
	Code         string  `json:"code"`
	File         *string `json:"file,omitempty"`
	Request      *Body   `json:"request"`
	Response     *Body   `json:"response"`
}

type Body struct {
	Headers map[string]string `json:"headers"`
	Body    map[string]any    `json:"body,omitempty"`
}

type LoggerOption func(*loggerConfig)

type loggerConfig struct {
	maskMap map[string]string
}
