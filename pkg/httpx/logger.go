package httpx

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/wisaitas/github.com/wisaitas/golang-structure/pkg/mask"
)

func NewLogger(config LoggerConfig) fiber.Handler {
	maskMap := NormalizeMaskMap(mask.ParsePatternMap(config.MaskMapPattern))
	return func(c fiber.Ctx) error {
		traceID := c.Get(HeaderTraceID)
		if traceID == "" {
			tid, _ := uuid.NewUUID()
			traceID = tid.String()
		}
		c.Request().Header.Set(HeaderTraceID, traceID)
		c.Set(HeaderTraceID, traceID)

		return HandleJSON(c, config.ServiceName, maskMap)
	}
}

func NewErrorResponse[T any](c fiber.Ctx, statusCode int, code ResponseCode, err error, publicMessage *string, wrapHandlerOp string) error {
	if err == nil {
		return nil
	}

	if wrapHandlerOp != "" {
		statusCode = StatusCodeFromError(err, fiber.StatusInternalServerError)
		err = WrapError(wrapHandlerOp, err, statusCode)
		code = ResponseCodeFromError(err)
		if code == "" {
			code = CodeForHTTPStatus(statusCode)
		}
	}

	errorMessage := RootErrorMessage(err)
	if errorMessage == "" {
		errorMessage = err.Error()
	}
	stackTraces := BuildErrorStackTraces(err)

	c.Locals("errorContext", ErrorContext{
		ErrorMessage: fmt.Sprintf("[httpx] : %s", errorMessage),
		StackTraces:  stackTraces,
	})

	return c.Status(statusCode).JSON(&StandardResponse[T]{
		Timestamp:     time.Now().Format(time.RFC3339),
		StatusCode:    statusCode,
		Data:          new(T),
		Code:          code,
		Pagination:    nil,
		PublicMessage: publicMessage,
	})
}

func NewSuccessResponse[T any](c fiber.Ctx, data *T, statusCode int, code ResponseCode, pagination *Pagination, publicMessage *string) error {
	return c.Status(statusCode).JSON(&StandardResponse[T]{
		Timestamp:     time.Now().Format(time.RFC3339),
		StatusCode:    statusCode,
		Data:          data,
		Code:          code,
		Pagination:    pagination,
		PublicMessage: publicMessage,
	})

}

func orgCodeFromResponseBody(body map[string]any) string {
	if body == nil {
		return ""
	}
	raw, ok := body["code"]
	if !ok || raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case string:
		return v
	default:
		return fmt.Sprint(v)
	}
}

func HandleJSON(c fiber.Ctx, serviceName string, maskMap map[string]string) error {
	start := time.Now()
	requestContext := WithDBLogCollector(c.Context())
	c.Locals("requestContext", requestContext)

	hasMask := len(maskMap) > 0

	var requestBody any
	contentType := string(c.Request().Header.ContentType())

	if len(contentType) >= 19 && contentType[:19] == "multipart/form-data" {
		payload := ReadMultipartForm(c, 64<<10)
		if hasMask {
			payload = MaskData(payload, maskMap)
		}
		requestBody = payload
	} else if hasMask {
		payload := ReadJSONMapLimited(c.Body(), 64<<10)
		requestBody = MaskData(payload, maskMap)
	} else {
		requestBody = clampedRawJSON(c.Body(), 64<<10)
	}

	requestHeaders := make(map[string]string)
	for key, value := range c.Request().Header.All() {
		if string(key) != HeaderTraceID {
			requestHeaders[string(key)] = string(value)
		}
	}

	if hasMask {
		requestHeaders = MaskHeaders(requestHeaders, maskMap)
	}

	if err := c.Next(); err != nil {
		return err
	}

	responseBody := c.Response().Body()
	responsePayload := ReadJSONMapLimited(responseBody, 64<<10)
	orgCode := orgCodeFromResponseBody(responsePayload)

	var responseLogBody any
	if hasMask {
		responseLogBody = MaskData(responsePayload, maskMap)
	} else if responsePayload != nil {
		limited := responseBody
		if len(limited) > 64<<10 {
			limited = limited[:64<<10]
		}
		copied := make([]byte, len(limited))
		copy(copied, limited)
		responseLogBody = json.RawMessage(copied)
	}

	responseHeaders := make(map[string]string)
	for key, value := range c.Response().Header.All() {
		if string(key) != HeaderTraceID && string(key) != HeaderSource {
			responseHeaders[string(key)] = string(value)
		}
	}

	if hasMask {
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

	var errMsgPtr *string
	if errorContext.ErrorMessage != "" {
		errMsgPtr = &errorContext.ErrorMessage
	}

	current := &Block{
		Service:      serviceName,
		Method:       c.Method(),
		Path:         c.Hostname() + string(c.Request().URI().RequestURI()),
		StatusCode:   strconv.Itoa(c.Response().StatusCode()),
		Code:         orgCode,
		Request:      &Body{Headers: requestHeaders, Body: requestBody},
		Response:     &Body{Headers: responseHeaders, Body: responseLogBody},
		ErrorMessage: errMsgPtr,
		StackTraces:  errorContext.StackTraces,
		DBLogs:       GetDBLogs(requestContext),
	}

	logInfo := Log{
		TraceID:    c.Get(HeaderTraceID),
		Timestamp:  start.Format(time.RFC3339),
		DurationMs: strconv.Itoa(int(time.Since(start).Milliseconds())),
		Current:    current,
	}

	sourceHeader := c.Response().Header.Peek(HeaderSource)

	if len(sourceHeader) > 0 {
		source := new(Block)
		if err := json.Unmarshal(sourceHeader, source); err != nil {
			log.Printf("[httpx] : %s", err.Error())
		}
		logInfo.Source = source
	}

	if c.Get(HeaderInternal) == "true" {
		if len(sourceHeader) == 0 {
			jsonResp, err := json.Marshal(current)
			if err != nil {
				log.Printf("[httpx] : %s", err.Error())
			}
			c.Response().Header.Set(HeaderSource, string(jsonResp))
		}
	} else {
		c.Response().Header.Del(HeaderSource)
	}

	jsonResp, err := json.Marshal(logInfo)
	if err != nil {
		log.Printf("[httpx] : %s", err.Error())
	}

	fmt.Println(string(jsonResp))
	return err
}
