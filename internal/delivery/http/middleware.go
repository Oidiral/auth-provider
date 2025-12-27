package http

import (
	"context"
	"net/http"
	"time"

	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"
)

func ZerologMiddleware(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := middleware.GetReqID(r.Context())
			ctx := context.WithValue(r.Context(), "request_id", requestID)

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r.WithContext(ctx))

			log.Info("HTTP request",
				logger.Field{Key: "method", Value: r.Method},
				logger.Field{Key: "path", Value: r.URL.Path},
				logger.Field{Key: "status", Value: ww.Status()},
				logger.Field{Key: "bytes", Value: ww.BytesWritten()},
				logger.Field{Key: "duration_ms", Value: time.Since(start).Milliseconds()},
				logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
				logger.Field{Key: "request_id", Value: requestID},
			)
		})
	}
}
