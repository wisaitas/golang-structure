package httpx

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func WithMaskMap(maskMap map[string]string) LoggerOption {
	return func(c *loggerConfig) {
		c.maskMap = maskMap
	}
}

func NewLogger(serviceName string, opts ...LoggerOption) fiber.Handler {
	config := &loggerConfig{
		maskMap: make(map[string]string),
	}

	for _, opt := range opts {
		opt(config)
	}

	return func(c *fiber.Ctx) error {
		traceID := c.Get(HeaderTraceID)
		if traceID == "" {
			tid, _ := uuid.NewV7()
			traceID = tid.String()
		}
		c.Request().Header.Set(HeaderTraceID, traceID)
		c.Set(HeaderTraceID, traceID)

		return HandleJSON(c, serviceName, config.maskMap)
	}
}

func NewErrorResponse[T any](c *fiber.Ctx, statusCode int, err error, publicMessage ...string) error {
	if err == nil {
		return nil
	}

	var code string

	switch statusCode {
	case 304:
		code = "E30400"
	case 400:
		code = "E40000"
	case 401:
		code = "E40002"
	case 403:
		code = "E40003"
	case 404:
		code = "E40004"
	case 500:
		code = "E50000"
	default:
		code = "E50000"
	}

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Println("[httpx] : runtime.Caller failed")
	}

	filePath := fmt.Sprintf("%s:%d", file, line)

	c.Locals("errorContext", ErrorContext{
		FilePath:     &filePath,
		ErrorMessage: err.Error(),
	})

	var msg *string
	if len(publicMessage) > 0 {
		msg = &publicMessage[0]
	}

	return c.Status(statusCode).JSON(&StandardResponse[T]{
		Timestamp:     time.Now().Format(time.RFC3339),
		StatusCode:    statusCode,
		Data:          new(T),
		Code:          code,
		Pagination:    nil,
		PublicMessage: msg,
	})
}

func NewSuccessResponse[T any](c *fiber.Ctx, data *T, statusCode int, pagination *Pagination, publicMessage ...string) error {
	var msg *string
	if len(publicMessage) > 0 {
		msg = &publicMessage[0]
	}

	var code string

	switch statusCode {
	case 200:
		code = "E20000"
	case 201:
		code = "E20001"
	case 204:
		code = "E20004"
	default:
		code = "E20000"
	}

	return c.Status(statusCode).JSON(&StandardResponse[T]{
		Timestamp:     time.Now().Format(time.RFC3339),
		StatusCode:    statusCode,
		Data:          data,
		Code:          code,
		Pagination:    pagination,
		PublicMessage: msg,
	})

}

func HandleJSON(c *fiber.Ctx, serviceName string, maskMap map[string]string) error {
	start := time.Now()

	var payload map[string]any
	contentType := string(c.Request().Header.ContentType())

	if len(contentType) >= 19 && contentType[:19] == "multipart/form-data" {
		payload = ReadMultipartForm(c, 64<<10)
	} else {
		payload = ReadJSONMapLimited(c.Body(), 64<<10)
	}

	if len(maskMap) > 0 {
		payload = MaskData(payload, maskMap)
	}

	requestHeaders := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		if string(key) != HeaderTraceID {
			requestHeaders[string(key)] = string(value)
		}
	})

	if len(maskMap) > 0 {
		requestHeaders = MaskHeaders(requestHeaders, maskMap)
	}

	if err := c.Next(); err != nil {
		return err
	}

	responseBody := c.Response().Body()
	responsePayload := ReadJSONMapLimited(responseBody, 64<<10)

	if len(maskMap) > 0 {
		responsePayload = MaskData(responsePayload, maskMap)
	}

	responseHeaders := make(map[string]string)
	c.Response().Header.VisitAll(func(key, value []byte) {
		if string(key) != HeaderTraceID && string(key) != HeaderSource {
			responseHeaders[string(key)] = string(value)
		}
	})

	if len(maskMap) > 0 {
		responseHeaders = MaskHeaders(responseHeaders, maskMap)
	}

	errorContext := &ErrorContext{}
	if !CheckStatusCode2xx(c.Response().StatusCode()) {
		errorContextLocal, ok := c.Locals("errorContext").(ErrorContext)
		if !ok {
			log.Println("[httpx] : errorContext not found")
		}
		errorContext = &errorContextLocal
	}

	current := &Block{
		Service:      serviceName,
		Method:       c.Method(),
		Path:         c.Hostname() + string(c.Request().URI().RequestURI()),
		StatusCode:   strconv.Itoa(c.Response().StatusCode()),
		Request:      &Body{Headers: requestHeaders, Body: payload},
		Response:     &Body{Headers: responseHeaders, Body: responsePayload},
		ErrorMessage: &errorContext.ErrorMessage,
		File:         errorContext.FilePath,
	}

	logInfo := Log{
		TraceID:    c.Get(HeaderTraceID),
		Timestamp:  start.Format(time.RFC3339),
		DurationMs: strconv.Itoa(int(time.Since(start).Milliseconds())),
		Current:    current,
	}

	if string(c.Response().Header.Peek(HeaderSource)) != "" {
		source := new(Block)
		if err := json.Unmarshal(c.Response().Header.Peek(HeaderSource), source); err != nil {
			log.Printf("[httpx] : %s", err.Error())
		}

		logInfo.Source = source
	} else if string(c.Response().Header.Peek(HeaderSource)) == "" {
		source := &Block{
			Service:      serviceName,
			Method:       c.Method(),
			Path:         c.Hostname() + string(c.Request().URI().RequestURI()),
			StatusCode:   strconv.Itoa(c.Response().StatusCode()),
			File:         errorContext.FilePath,
			ErrorMessage: &errorContext.ErrorMessage,
			Request:      &Body{Headers: requestHeaders, Body: payload},
			Response:     &Body{Headers: responseHeaders, Body: responsePayload},
		}

		jsonResp, err := json.Marshal(source)
		if err != nil {
			log.Printf("[httpx] : %s", err.Error())
		}
		c.Response().Header.Set(HeaderSource, string(jsonResp))
	}

	if c.Get(HeaderInternal) != "true" {
		c.Response().Header.Del(HeaderSource)
	}

	jsonResp, err := json.Marshal(logInfo)
	if err != nil {
		log.Printf("[httpx] : %s", err.Error())
	}

	fmt.Println(string(jsonResp))
	return err
}
