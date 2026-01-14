package http

import (
	"context"
	"net/http"
	"time"

	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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

func TracingMiddleware(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

			ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.target", r.URL.Path),
					attribute.String("http.host", r.Host),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("net.peer.ip", r.RemoteAddr),
				),
			)
			defer span.End()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r.WithContext(ctx))

			status := ww.Status()
			span.SetAttributes(attribute.Int("http.status_code", status))

			if status >= 400 {
				span.SetStatus(codes.Error, http.StatusText(status))
			}
		})
	}
}
