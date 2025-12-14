package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
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
func RequestResponseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read and restore request body
		var reqBodyBytes []byte
		if r.Body != nil {
			reqBodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		log.Printf("[Request] %s %s | Body: %s", r.Method, r.URL.Path, string(reqBodyBytes))

		// Initialize custom response writer
		observer := &ResponseObserver{
			ResponseWriter: w,
			StatusCode:     http.StatusOK, // Default to 200 OK
			Body:           &bytes.Buffer{},
		}

		next.ServeHTTP(observer, r)

		log.Printf("[Response] %s %s | Status: %d | Duration: %v | Body: %s",
			r.Method, r.URL.Path, observer.StatusCode, time.Since(start), observer.Body.String())
	})
}
