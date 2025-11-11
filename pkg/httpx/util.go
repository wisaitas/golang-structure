package httpx

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func CheckStatusCode2xx(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

func ReadJSONMapLimited(b []byte, limit int) map[string]any {
	if len(b) > limit {
		b = b[:limit]
	}
	return TryParseJSON(b)
}

func TryParseJSON(b []byte) map[string]any {
	if len(b) == 0 {
		return nil
	}
	var m map[string]any
	if json.Valid(b) && json.Unmarshal(b, &m) == nil {
		return m
	}
	return nil
}

func ReadMultipartForm(c *fiber.Ctx, limit int) map[string]any {
	form, err := c.MultipartForm()
	if err != nil {
		return nil
	}

	result := make(map[string]any)

	for key, values := range form.Value {
		if len(values) == 1 {
			result[key] = values[0]
		} else {
			result[key] = values
		}
	}

	for key, files := range form.File {
		fileInfos := make([]map[string]any, 0, len(files))
		for _, file := range files {
			fileInfo := map[string]any{
				"filename": file.Filename,
				"size":     file.Size,
				"header":   convertMultipartHeader(file.Header),
			}
			fileInfos = append(fileInfos, fileInfo)
		}

		if len(fileInfos) == 1 {
			result[key] = fileInfos[0]
		} else {
			result[key] = fileInfos
		}
	}

	return result
}

func convertMultipartHeader(header map[string][]string) map[string]any {
	result := make(map[string]any)
	for key, values := range header {
		if len(values) == 1 {
			result[key] = values[0]
		} else {
			result[key] = values
		}
	}
	return result
}

func MaskData(data map[string]any, maskMap map[string]string) map[string]any {
	if data == nil || maskMap == nil || len(maskMap) == 0 {
		return data
	}

	result := make(map[string]any)
	for k, v := range data {
		masked := false
		for maskKey, maskValue := range maskMap {
			if strings.EqualFold(k, maskKey) {
				result[k] = applyMask(v, maskValue)
				masked = true
				break
			}
		}

		if !masked {
			switch val := v.(type) {
			case map[string]any:
				result[k] = MaskData(val, maskMap)
			case []any:
				result[k] = maskSlice(val, maskMap)
			default:
				result[k] = v
			}
		}
	}

	return result
}

func applyMask(value any, maskPattern string) any {
	strValue, ok := value.(string)
	if !ok {
		strValue = toString(value)
	}

	length := len(strValue)

	if maskPattern == "*" {
		if length <= 2 {
			return "**"
		}
		if length == 3 {
			return string(strValue[0]) + "*" + string(strValue[2])
		}
		maskLen := length - 2
		return string(strValue[0]) + strings.Repeat("*", maskLen) + string(strValue[length-1])
	}

	if maskPattern == "**" {
		if length <= 4 {
			return "**"
		}
		maskLen := length - 4
		return strValue[:2] + strings.Repeat("*", maskLen) + strValue[length-2:]
	}

	if maskPattern == "***" {
		if length <= 6 {
			return "***"
		}
		maskLen := length - 6
		return strValue[:3] + strings.Repeat("*", maskLen) + strValue[length-3:]
	}

	return maskPattern
}

func toString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", v)
	}
}

func maskSlice(slice []any, maskMap map[string]string) []any {
	if slice == nil {
		return nil
	}

	result := make([]any, len(slice))
	for i, item := range slice {
		switch val := item.(type) {
		case map[string]any:
			result[i] = MaskData(val, maskMap)
		case []any:
			result[i] = maskSlice(val, maskMap)
		default:
			result[i] = val
		}
	}

	return result
}

func MaskHeaders(headers map[string]string, maskMap map[string]string) map[string]string {
	if headers == nil || maskMap == nil || len(maskMap) == 0 {
		return headers
	}

	result := make(map[string]string)
	for k, v := range headers {
		masked := false
		for maskKey, maskValue := range maskMap {
			if strings.EqualFold(k, maskKey) {
				if maskedValue, ok := applyMask(v, maskValue).(string); ok {
					result[k] = maskedValue
				} else {
					result[k] = maskValue
				}
				masked = true
				break
			}
		}
		if !masked {
			result[k] = v
		}
	}

	return result
}

func MaskQueryParams(c *fiber.Ctx, maskMap map[string]string) map[string]string {
	if len(maskMap) == 0 {
		return nil
	}

	result := make(map[string]string)
	c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
		keyStr := string(key)
		valueStr := string(value)

		masked := false
		for maskKey, maskValue := range maskMap {
			if strings.EqualFold(keyStr, maskKey) {
				result[keyStr] = maskValue
				masked = true
				break
			}
		}
		if !masked {
			result[keyStr] = valueStr
		}
	})

	return result
}

func MaskParams(c *fiber.Ctx, maskMap map[string]string) map[string]string {
	if len(maskMap) == 0 {
		return nil
	}

	result := make(map[string]string)
	for _, paramName := range c.Route().Params {
		paramValue := c.Params(paramName)

		masked := false
		for maskKey, maskValue := range maskMap {
			if strings.EqualFold(paramName, maskKey) {
				result[paramName] = maskValue
				masked = true
				break
			}
		}
		if !masked {
			result[paramName] = paramValue
		}
	}

	return result
}
