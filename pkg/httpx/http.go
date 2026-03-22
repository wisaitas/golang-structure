package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/gofiber/fiber/v3"
)

func Client[T any](c fiber.Ctx, method string, url string, req any, resp *StandardResponse[T]) error {
	client := HttpClient
	reqJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("[httpx] : %w", err)
	}
	body := bytes.NewReader(reqJson)

	reqHttp, err := http.NewRequestWithContext(c.Context(), method, url, body)
	if err != nil {
		return fmt.Errorf("[httpx] : %w", err)
	}

	for key, values := range c.GetReqHeaders() {
		reqHttp.Header.Add(HeaderInternal, "true")
		for _, value := range values {
			reqHttp.Header.Add(key, value)
		}
	}

	respHttp, err := client.Do(reqHttp)
	if err != nil {
		return fmt.Errorf("[httpx] : %w", err)
	}
	defer respHttp.Body.Close()

	for key, values := range respHttp.Header {
		for _, value := range values {
			if key != HeaderTraceID {
				c.Response().Header.Add(key, value)
			}
		}
	}

	if err = json.NewDecoder(respHttp.Body).Decode(resp); err != nil {
		return fmt.Errorf("[httpx] : %w", err)
	}

	if !CheckStatusCode2xx(respHttp.StatusCode) {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			log.Println("[httpx] : runtime.Caller failed")
		}
		filePath := fmt.Sprintf("%s:%d", file, line)
		resp.Data = nil
		resp.Pagination = nil
		c.Locals("errorContext", ErrorContext{
			FilePath:     &filePath,
			ErrorMessage: fmt.Sprintf("[httpx] : error when call %s", url),
		})
	}

	return nil
}

// fiber:context-methods migrated
