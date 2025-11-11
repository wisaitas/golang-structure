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
