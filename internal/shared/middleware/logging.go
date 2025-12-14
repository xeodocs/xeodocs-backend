package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
)

// ANSI Color Codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
)

// ResponseObserver wraps http.ResponseWriter to capture status and body
type ResponseObserver struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func (w *ResponseObserver) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *ResponseObserver) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestResponseLogger logs the details of each request and response including bodies
func RequestResponseLogger(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Read and restore request body
			var reqBodyBytes []byte
			if r.Body != nil {
				reqBodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
			}

			// Initialize custom response writer
			observer := &ResponseObserver{
				ResponseWriter: w,
				StatusCode:     http.StatusOK, // Default to 200 OK
				Body:           &bytes.Buffer{},
			}

			next.ServeHTTP(observer, r)

			duration := time.Since(start)

			// Formatting
			method := colorizeMethod(r.Method)
			path := fmt.Sprintf("%s%s%s", ColorCyan, r.URL.Path, ColorReset)
			status := colorizeStatus(observer.StatusCode)
			durationStr := fmt.Sprintf("%s%v%s", ColorGray, duration, ColorReset)

			// Visual Divider (not too long)
			log.Println(ColorGray + "···············································" + ColorReset)

			// Request Log
			log.Printf("%s %s", method, path)

			// Only log body in development
			if cfg.Environment == "development" {
				reqBody := prettyPrintJSON(reqBodyBytes)
				if len(reqBody) > 0 {
					log.Printf("%sRequest Body:%s\n%s", ColorPurple, ColorReset, reqBody)
				}
			}

			// Response Log
			log.Printf("➜ Status: %s | Duration: %s", status, durationStr)

			// Only log body in development
			if cfg.Environment == "development" {
				resBody := prettyPrintJSON(observer.Body.Bytes())
				if len(resBody) > 0 {
					log.Printf("%sResponse Body:%s\n%s", ColorBlue, ColorReset, resBody)
				}
			}
		})
	}
}

func colorizeMethod(method string) string {
	color := ColorReset
	switch method {
	case "GET":
		color = ColorBlue
	case "POST":
		color = ColorGreen
	case "PUT":
		color = ColorYellow
	case "PATCH":
		color = ColorYellow
	case "DELETE":
		color = ColorRed
	}
	return fmt.Sprintf("%s%s%s", color, method, ColorReset)
}

func colorizeStatus(code int) string {
	color := ColorReset
	switch {
	case code >= 200 && code < 300:
		color = ColorGreen
	case code >= 300 && code < 400:
		color = ColorBlue
	case code >= 400 && code < 500:
		color = ColorYellow
	case code >= 500:
		color = ColorRed
	}
	return fmt.Sprintf("%s%d%s", color, code, ColorReset)
}

func prettyPrintJSON(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	var out bytes.Buffer
	if err := json.Indent(&out, b, "", "  "); err != nil {
		return string(b)
	}
	// Colorize keys and strings roughly if needed, but indentation is a good start.
	// For now just indentation.
	return out.String()
}
