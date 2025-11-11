package httpx

type ErrorContext struct {
	FilePath     *string `json:"filePath"`
	ErrorMessage string  `json:"errorMessage"`
}

type StandardResponse[T any] struct {
	Timestamp     string      `json:"timestamp"`
	StatusCode    int         `json:"status_code"`
	Code          string      `json:"code"`
	Data          *T          `json:"data"`
	Pagination    *Pagination `json:"pagination,omitempty"`
	PublicMessage *string     `json:"public_message,omitempty"`
}

type Pagination struct {
	Page          int   `json:"page"`
	PageSize      int   `json:"page_size"`
	HasNext       bool  `json:"has_next"`
	HasPrev       bool  `json:"has_prev"`
	TotalElements int   `json:"total_elements"`
	Windows       []int `json:"windows"`
}

type PaginationQuery struct {
	Page     *int `query:"page"`
	PageSize *int `query:"page_size"`
}

type Log struct {
	TraceID    string `json:"trace_id"`
	Timestamp  string `json:"timestamp"`
	DurationMs string `json:"duration_ms"`

	Current *Block `json:"current"`
	Source  *Block `json:"source,omitempty"`
}

type Block struct {
	Service      string  `json:"service"`
	Method       string  `json:"method"`
	ErrorMessage *string `json:"error_message,omitempty"`
	Path         string  `json:"path"`
	StatusCode   string  `json:"status_code"`
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
