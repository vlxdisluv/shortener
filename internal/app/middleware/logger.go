package middleware

import (
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/vlxdisluv/shortener/internal/app/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.bytes += size
	return size, err
}

func (lrw *loggingResponseWriter) WriteHeader(status int) {
	lrw.ResponseWriter.WriteHeader(status)
	lrw.status = status
}

func (lrw *loggingResponseWriter) Status() int {
	return lrw.status
}

func (lrw *loggingResponseWriter) BytesWritten() int {
	return lrw.bytes
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{ResponseWriter: w}
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := newLoggingResponseWriter(w)
		start := time.Now()

		next.ServeHTTP(lrw, r)
		duration := time.Since(start)

		requestID := chiMiddleware.GetReqID(r.Context())

		logger.Log.Info("incoming HTTP request",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.RequestURI),
			zap.Int("status", lrw.Status()),
			zap.Duration("duration", duration),
			zap.Int("size", lrw.BytesWritten()),
		)
	})
}
