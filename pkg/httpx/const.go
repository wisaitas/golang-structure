package httpx

import (
	"net/http"
	"time"
)

const (
	HeaderTraceID      = "X-Trace-Id"
	HeaderErrSignature = "X-Error-Signature"
	HeaderInternal     = "X-Internal-Call"
	HeaderSource       = "X-Source"
)

var HttpClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:          100,              // จำนวน connection สูงสุดที่เก็บไว้ใน connection pool เพื่อนำกลับมาใช้ใหม่
		IdleConnTimeout:       90 * time.Second, // เวลาที่จะปิด connection หากไม่มีการใช้งาน
		TLSHandshakeTimeout:   5 * time.Second,  // เวลาสูงสุดที่รอให้ TLS handshake เสร็จสมบูรณ์
		ResponseHeaderTimeout: 3 * time.Second,  // เวลาสูงสุดที่รอการตอบกลับ header จากเซิร์ฟเวอร์
		ExpectContinueTimeout: 1 * time.Second,  // เวลาสูงสุดที่รอการตอบกลับ "100 Continue" จากเซิร์ฟเวอร์ก่อนส่งข้อมูล request body
	},
}

const (
	ErrorCodeConnectionRefused = "E50000"
)

type ErrorCode string

type ResponseCode string

func (c ResponseCode) String() string {
	return string(c)
}

const (
	// 2xx
	CodeOK        ResponseCode = "E20000"
	CodeCreated   ResponseCode = "E20001"
	CodeNoContent ResponseCode = "E20004"

	// 3xx
	CodeNotModified ResponseCode = "E30400"

	// 4xx
	CodeBadRequest   ResponseCode = "E40000"
	CodeUnauthorized ResponseCode = "E40002"
	CodeForbidden    ResponseCode = "E40003"
	CodeNotFound     ResponseCode = "E40004"
	CodeConflict     ResponseCode = "E40900"

	// 5xx
	CodeInternal           ResponseCode = "E50000"
	CodeBadGateway         ResponseCode = "E50200"
	CodeServiceUnavailable ResponseCode = "E50300"
)

func CodeForHTTPStatus(statusCode int) ResponseCode {
	switch statusCode {
	case 304:
		return CodeNotModified
	case 400:
		return CodeBadRequest
	case 401:
		return CodeUnauthorized
	case 403:
		return CodeForbidden
	case 404:
		return CodeNotFound
	case 409:
		return CodeConflict
	case 502:
		return CodeBadGateway
	case 500:
		return CodeInternal
	default:
		if statusCode >= 200 && statusCode < 300 {
			return CodeOK
		}
		if statusCode >= 400 && statusCode < 500 {
			return CodeBadRequest
		}
		return CodeInternal
	}
}
