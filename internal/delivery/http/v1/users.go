package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apiv1 "github.com/Oidiral/auth-provider/internal/generated/api/v1"
	"github.com/Oidiral/auth-provider/internal/service"
	"github.com/Oidiral/auth-provider/pkg/logger"
	"github.com/go-playground/validator/v10"
)

var _ apiv1.ServerInterface = (*Handler)(nil)

type Handler struct {
	userService service.Users
	validate    *validator.Validate
	log         logger.Logger
}

func NewHandler(userService service.Users, validate *validator.Validate, log logger.Logger) *Handler {
	return &Handler{
		userService: userService,
		validate:    validate,
		log:         log,
	}
}

func (h *Handler) UserSignUp(w http.ResponseWriter, r *http.Request) {
	var req apiv1.UserSignUpJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "malformed JSON or invalid data types")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	err := h.userService.SignUp(r.Context(), service.UserSignUpInput{
		Username:  req.Username,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Email:     string(req.Email),
		Password:  req.Password,
	})
	if err != nil {
		respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, apiv1.SuccessResponse{
		Status:  http.StatusCreated,
		Message: "success",
	})
}

func (h *Handler) UserSignIn(w http.ResponseWriter, r *http.Request) {
	var req apiv1.UserSignInJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "malformed JSON or invalid data types")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	res, err := h.userService.SignIn(r.Context(), service.UserSignInInput{
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, apiv1.SuccessResponse{
		Status:  http.StatusOK,
		Message: "success",
		Data: &map[string]interface{}{
			"AccessToken":  res.AccessToken,
			"RefreshToken": res.RefreshToken,
		},
	})
}

func (h *Handler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	token, err := extractBearerToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userId, roles, err := h.userService.ValidateToken(r.Context(), token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("X-User-Id", userId)
	w.Header().Set("X-User-Roles", strings.Join(roles, ","))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UserVerify(w http.ResponseWriter, r *http.Request) {
	var req apiv1.UserVerifyJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "malformed JSON or invalid data types")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	err := h.userService.Verify(r.Context(), req.UserId.String(), req.Otp)
	if err != nil {
		respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, apiv1.SuccessResponse{
		Status:  http.StatusOK,
		Message: "success",
	})
}

func (h *Handler) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	var req apiv1.RefreshTokensJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", "malformed JSON or invalid data types")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	res, err := h.userService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		respondServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, apiv1.SuccessResponse{
		Status:  http.StatusOK,
		Message: "success",
		Data: &map[string]interface{}{
			"AccessToken":  res.AccessToken,
			"RefreshToken": res.RefreshToken,
		},
	})
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return strings.TrimPrefix(authHeader, bearerPrefix), nil
}
