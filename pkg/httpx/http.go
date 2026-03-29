package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

func Client[T any](c fiber.Ctx, method string, url string, req any, resp *StandardResponse[T]) error {
	client := HttpClient
	reqJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("[httpx] marshal request body : %w", err)
	}
	body := bytes.NewReader(reqJson)

	reqHttp, err := http.NewRequestWithContext(c.Context(), method, url, body)
	if err != nil {
		return fmt.Errorf("[httpx] new request with context : %w", err)
	}

	for key, values := range c.GetReqHeaders() {
		reqHttp.Header.Add(HeaderInternal, "true")
		for _, value := range values {
			reqHttp.Header.Add(key, value)
		}
	}

	respHttp, err := client.Do(reqHttp)
	if err != nil {
		fmt.Println("eiei0", req)
		resp.StatusCode = http.StatusBadGateway
		return fmt.Errorf("[httpx] do request : %w", err)
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
		return fmt.Errorf("[httpx] decode response body : %w", err)
	}

	if resp.StatusCode == 0 {
		resp.StatusCode = http.StatusBadGateway
	}

	return nil
}
