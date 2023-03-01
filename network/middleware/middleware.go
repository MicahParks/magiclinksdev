package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/handle"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/storage"
)

const (
	// APIKeyHeader is the header that contains the API key.
	APIKeyHeader = "X-API-KEY"
)

// Middleware is a function that wraps a handler.
type Middleware func(next http.Handler) http.Handler

// Apply applies the middleware to the handler.
func Apply(server *handle.Server, options handle.MiddlewareOptions) http.Handler {
	options = server.MiddlewareHook.Hook(options)
	h := options.Handler
	if options.Toggle.CommitTx {
		h = wrap(h, commitTx)
	}
	if options.Toggle.RateLimit {
		h = wrap(h, createRateLimit(server))
	}
	if options.Toggle.Admin {
		h = wrap(h, admin)
	}
	if options.Toggle.Authn {
		h = wrap(h, createAuthn(server))
	}
	h = wrap(h, createTx(server), createLogger(server), createTimeout(server), requestUUID, createLimitRequestBody(server))
	return h
}

func admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		sa := ctx.Value(ctxkey.ServiceAccount).(model.ServiceAccount)
		if !sa.Admin {
			WriteErrorBody(ctx, http.StatusForbidden, "Forbidden", writer)
			return
		}
		next.ServeHTTP(writer, req)
	})
}

func createAuthn(server *handle.Server) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			reqUUID := ctx.Value(ctxkey.RequestUUID).(uuid.UUID)

			headerValue := req.Header.Get(APIKeyHeader)
			if headerValue == "" {
				WriteErrorBody(ctx, http.StatusUnauthorized, mld.ResponseUnauthorized, w)
				return
			}

			apiKey, err := uuid.Parse(headerValue)
			if err != nil {
				WriteErrorBody(ctx, http.StatusUnauthorized, mld.ResponseUnauthorized, w)
				return
			}

			sa, err := server.Store.ReadSAFromAPIKey(ctx, apiKey)
			if err != nil {
				WriteErrorBody(ctx, http.StatusUnauthorized, mld.ResponseUnauthorized, w)
				return
			}
			ctx = context.WithValue(ctx, ctxkey.ServiceAccount, sa)

			fields := createZapFields(req, reqUUID)
			fields = append(fields, zap.String("saUUID", sa.UUID.String()))
			logger := server.Sugared.With(fields...)
			logger.Debug("Request authenticated.")
			ctx = context.WithValue(ctx, ctxkey.Sugared, logger)

			req = req.WithContext(ctx)
			next.ServeHTTP(w, req)
		})
	}
}

func createZapFields(req *http.Request, reqUUID uuid.UUID) []interface{} {
	return []interface{}{
		zap.String("method", req.Method),
		zap.String("requestUUID", reqUUID.String()),
		zap.String("url", req.URL.String()),
	}
}

func createLogger(server *handle.Server) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			reqUUID := ctx.Value(ctxkey.RequestUUID).(uuid.UUID)
			logger := server.Sugared.With(createZapFields(req, reqUUID)...)
			logger.Debug("Request started.")
			ctx = context.WithValue(ctx, ctxkey.Sugared, logger)
			req = req.WithContext(ctx)
			next.ServeHTTP(writer, req)
		})
	}
}

// commitTx commits the transaction after the given Handler has completed. The given Handler is expected to rollback
// if the request fails.
func commitTx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(writer, req)
		ctx := req.Context()
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)
		err := tx.Commit(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrTxClosed) {
				return
			}
			sugared := ctx.Value(ctxkey.Sugared).(*zap.SugaredLogger)
			sugared.Errorw("Failed to commit transaction.",
				mld.LogErr, err,
			)
			WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, writer)
			return
		}
	})
}

func createLimitRequestBody(server *handle.Server) Middleware {
	maxBodyBytes := server.Config.RequestMaxBodyBytes
	if maxBodyBytes == 0 {
		maxBodyBytes = 1 << 20 // 1 MB
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			req.Body = http.MaxBytesReader(writer, req.Body, maxBodyBytes)
			next.ServeHTTP(writer, req)
		})
	}
}

func createRateLimit(server *handle.Server) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			sa := ctx.Value(ctxkey.ServiceAccount).(model.ServiceAccount)
			err := server.Limiter.Wait(ctx, sa.UUID.String())
			if err != nil {
				sugared := ctx.Value(ctxkey.Sugared).(*zap.SugaredLogger)
				sugared.Warnw("Service account request exceeds rate limit.",
					mld.LogErr, err,
				)
				WriteErrorBody(ctx, http.StatusTooManyRequests, mld.ResponseTooManyRequests, writer)
				return
			}
			next.ServeHTTP(writer, req)
		})
	}
}

func createTx(server *handle.Server) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			sugared := ctx.Value(ctxkey.Sugared).(*zap.SugaredLogger)

			tx, err := server.Store.Begin(ctx)
			if err != nil {
				sugared.Errorw("Failed to start transaction.",
					mld.LogErr, err,
				)
				WriteErrorBody(ctx, http.StatusInternalServerError, mld.ResponseInternalServerError, writer)
				return
			}

			ctx = context.WithValue(ctx, ctxkey.Tx, tx)
			req = req.WithContext(ctx)
			next.ServeHTTP(writer, req)

			err = tx.Rollback(ctx)
			if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
				sugared.Errorw("Failed to rollback transaction.",
					mld.LogErr, err,
				)
				return
			}
		})
	}
}

func createTimeout(server *handle.Server) Middleware {
	timeout := server.Config.RequestTimeout.Get()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			req = req.WithContext(ctx)
			next.ServeHTTP(writer, req)
		})
	}
}

func requestUUID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		reqUUID, _ := uuid.NewRandom()
		ctx = context.WithValue(ctx, ctxkey.RequestUUID, reqUUID)
		req = req.WithContext(ctx)
		next.ServeHTTP(writer, req)
	})
}

func wrap(handler http.Handler, middleware ...Middleware) http.Handler {
	for _, m := range middleware {
		handler = m(handler)
	}
	return handler
}

func WriteErrorBody(ctx context.Context, code int, message string, writer http.ResponseWriter) {
	data, err := json.Marshal(model.NewError(ctx, code, message))
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Header().Set(mld.HeaderContentType, mld.ContentTypeJSON)
	writer.WriteHeader(code)
	_, _ = writer.Write(data)
}
