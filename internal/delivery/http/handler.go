package http

import (
	"reflect"
	"strings"
	"time"

	v1 "github.com/Oidiral/auth-provider/internal/delivery/http/v1"
	apiv1 "github.com/Oidiral/auth-provider/internal/generated/api/v1"
	"github.com/Oidiral/auth-provider/internal/service"
	"github.com/Oidiral/auth-provider/pkg/logger"
	validators "github.com/Oidiral/auth-provider/pkg/validator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/rs/cors"
)

type Handler struct {
	services    *service.Services
	serviceName string
}

func NewHandler(services *service.Services, serviceName string) *Handler {
	return &Handler{
		services:    services,
		serviceName: serviceName,
	}
}

func (h *Handler) Init(log logger.Logger) chi.Router {
	r := chi.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	})

	r.Use(c.Handler,
		middleware.Recoverer,
		middleware.RequestID,
		middleware.RealIP,
		TracingMiddleware(h.serviceName),
		ZerologMiddleware(log),
		middleware.Timeout(60*time.Second),
		middleware.Throttle(100),
		middleware.RequestSize(1024*1024*10),
		middleware.Heartbeat("/health"),
	)

	h.initAPI(r, log)

	return r
}

func initValidation() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	if err := validators.RegisterCustomValidators(v); err != nil {
		panic("failed to register custom validators: " + err.Error())
	}

	return v

}

func (h *Handler) initAPI(router chi.Router, log logger.Logger) {
	v := initValidation()
	userAPI := v1.NewHandler(h.services.Users, v, log)
	apiv1.HandlerFromMux(userAPI, router)
}
