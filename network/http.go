package network

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/handle"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network/middleware"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/storage"
)

const (
	responseDontRegisteredClaims = "Do not provide JWT registered claims."
)

// Validatable is an interface for validating a model.
type Validatable[T any] interface {
	Validate(config model.Validation) (T, error)
}

// HTTPReady creates an HTTP handler that always returns http.StatusOK.
func HTTPReady(_ *handle.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// HTTPServiceAccountCreate creates an HTTP handler for the HandleServiceAccountCreate method.
func HTTPServiceAccountCreate(s *handle.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)

		validated, done := unmarshalRequest[model.ServiceAccountCreateRequest, model.ValidServiceAccountCreateRequest](r, s.Config.Validation, w)
		if done {
			return
		}

		response, err := s.HandleServiceAccountCreate(ctx, validated)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create service account.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to commit transaction for create service account.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		logger.InfoContext(ctx, "Created new service account.",
			mld.LogRequestBody, validated,
			mld.LogResponseBody, response,
		)

		writeResponse(ctx, http.StatusCreated, response, w)
	})
}

// HTTPJWTCreate creates an HTTP handler for the HandleJWTCreate method.
func HTTPJWTCreate(s *handle.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)

		validated, done := unmarshalRequest[model.JWTCreateRequest, model.ValidJWTCreateRequest](r, s.Config.Validation, w)
		if done {
			return
		}

		response, err := s.HandleJWTCreate(ctx, validated)
		if err != nil {
			if errors.Is(err, handle.ErrRegisteredClaimProvided) {
				middleware.WriteErrorBody(ctx, http.StatusBadRequest, responseDontRegisteredClaims, w)
				return
			}
			if errors.Is(err, handle.ErrJWTAlgNotFound) {
				middleware.WriteErrorBody(ctx, http.StatusBadRequest, "Specified JWT algorithm not found.", w)
				return
			}
			logger.ErrorContext(ctx, "Failed to create JWT.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to commit transaction for create JWT.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		writeResponse(ctx, http.StatusCreated, response, w)
	})
}

// HTTPJWTValidate creates an HTTP handler for the HandleJWTValidate method.
func HTTPJWTValidate(s *handle.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)

		validated, done := unmarshalRequest[model.JWTValidateRequest, model.ValidJWTValidateRequest](r, s.Config.Validation, w)
		if done {
			return
		}

		response, err := s.HandleJWTValidate(ctx, validated)
		if err != nil {
			if errors.Is(err, handle.ErrToken) {
				middleware.WriteErrorBody(ctx, http.StatusUnprocessableEntity, fmt.Sprintf("Invalid JWT: %s", err), w)
				return
			}
			logger.ErrorContext(ctx, "Failed to validate JWT.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to handle request: %s", err), w)
			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to commit transaction for validate JWT.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, fmt.Sprintf("Failed to commit storage transaction: %s.", err), w)
			return
		}

		writeResponse(ctx, http.StatusOK, response, w)
	})
}

// HTTPMagicLinkCreate creates an HTTP handler for the HandleMagicLinkCreate method.
func HTTPMagicLinkCreate(s *handle.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)

		validated, done := unmarshalRequest[model.MagicLinkCreateRequest, model.ValidMagicLinkCreateRequest](r, s.Config.Validation, w)
		if done {
			return
		}

		response, err := s.HandleMagicLinkCreate(ctx, validated)
		if err != nil {
			if errors.Is(err, handle.ErrRegisteredClaimProvided) {
				middleware.WriteErrorBody(ctx, http.StatusBadRequest, responseDontRegisteredClaims, w)
				return
			}
			logger.ErrorContext(ctx, "Failed to commit transaction for create link.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to commit transaction for create link.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		writeResponse(ctx, http.StatusCreated, response, w)
	})
}

// HTTPMagicLinkEmailCreate creates an HTTP handler for the HandleMagicLinkEmailCreate method.
func HTTPMagicLinkEmailCreate(s *handle.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)

		validated, done := unmarshalRequest[model.MagicLinkEmailCreateRequest, model.ValidMagicLinkEmailCreateRequest](r, s.Config.Validation, w)
		if done {
			return
		}

		response, err := s.HandleMagicLinkEmailCreate(ctx, validated)
		if err != nil {
			if errors.Is(err, handle.ErrRegisteredClaimProvided) {
				middleware.WriteErrorBody(ctx, http.StatusBadRequest, responseDontRegisteredClaims, w)
				return
			}
			logger.ErrorContext(ctx, "Failed to create email link.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to commit transaction for create email link.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		writeResponse(ctx, http.StatusCreated, response, w)
	})
}

// HTTPOTPCreate creates an HTTP handler for the HandleOTPCreate method.
func HTTPOTPCreate(s *handle.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)

		validated, done := unmarshalRequest[model.OTPCreateRequest, model.ValidOTPCreateRequest](r, s.Config.Validation, w)
		if done {
			return
		}

		response, err := s.HandleOTPCreate(ctx, validated)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create OTP.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to commit transaction for create OTP.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		writeResponse(ctx, http.StatusCreated, response, w)
	})
}

// HTTPOTPValidate creates an HTTP handler for the HandleOTPValidate method.
func HTTPOTPValidate(s *handle.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)

		validated, done := unmarshalRequest[model.OTPValidateRequest, model.ValidOTPValidateRequest](r, s.Config.Validation, w)
		if done {
			return
		}

		response, err := s.HandleOTPValidate(ctx, validated)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to validate OTP.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to commit transaction for validate OTP.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		writeResponse(ctx, http.StatusOK, response, w)
	})
}

// HTTPOTPEmailCreate creates an HTTP handler for the HandleOTPEmailCreate method.
func HTTPOTPEmailCreate(s *handle.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)

		validated, done := unmarshalRequest[model.OTPEmailCreateRequest, model.ValidOTPEmailCreateRequest](r, s.Config.Validation, w)
		if done {
			return
		}

		response, err := s.HandleOTPEmailCreate(ctx, validated)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create OTP email.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		err = tx.Commit(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to commit transaction for create OTP email.",
				mld.LogErr, err,
			)
			middleware.WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, w)
			return
		}

		writeResponse(ctx, http.StatusCreated, response, w)
	})
}

func writeResponse(ctx context.Context, code int, response any, w http.ResponseWriter) {
	body, err := json.Marshal(response)
	if err != nil {
		middleware.WriteErrorBody(ctx, http.StatusInternalServerError, "Failed to JSON marshal response body.", w)
		return
	}
	w.Header().Set(mld.HeaderContentType, mld.ContentTypeJSON)

	w.WriteHeader(code)
	_, _ = w.Write(body)
}

func unmarshalRequest[T Validatable[K], K any](r *http.Request, validationConfig model.Validation, w http.ResponseWriter) (K, bool) {
	ctx := r.Context()
	var validated K

	if r.Header.Get(mld.HeaderContentType) != mld.ContentTypeJSON {
		middleware.WriteErrorBody(ctx, http.StatusUnsupportedMediaType, fmt.Sprintf("Invalid content type: %q.", r.Header.Get(mld.HeaderContentType)), w)
		return validated, true
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		middleware.WriteErrorBody(ctx, http.StatusBadRequest, "Failed to read response body", w)
		return validated, true
	}
	err = r.Body.Close()
	if err != nil {
		middleware.WriteErrorBody(ctx, http.StatusBadRequest, "Failed to close response body.", w)
		return validated, true
	}

	var unvalidated T
	err = json.Unmarshal(body, &unvalidated)
	if err != nil {
		middleware.WriteErrorBody(ctx, http.StatusBadRequest, "Failed to JSON unmarshal request body.", w)
		return validated, true
	}

	validated, err = unvalidated.Validate(validationConfig)
	if err != nil {
		middleware.WriteErrorBody(ctx, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %s.", err), w)
		return validated, true
	}

	return validated, false
}
