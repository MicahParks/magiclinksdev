package setup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MicahParks/magiclinksdev/config"
	"github.com/MicahParks/magiclinksdev/email/sendgrid"
	"github.com/MicahParks/magiclinksdev/email/ses"
	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/mldtest"
	"github.com/MicahParks/magiclinksdev/network"
	"github.com/MicahParks/magiclinksdev/rlimit"

	mld "github.com/MicahParks/magiclinksdev"
	"github.com/MicahParks/magiclinksdev/email"
	"github.com/MicahParks/magiclinksdev/handle"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/storage"
)

// NopConfig is the configuration for a no-operation email provider magiclinksdev server.
type NopConfig struct {
	Server      config.Config  `json:"server"`
	Storage     storage.Config `json:"storage"`
	RateLimiter rlimit.Config  `json:"rateLimiter"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (n NopConfig) DefaultsAndValidate() (NopConfig, error) {
	const errMsg = "failed to validate and apply defaults to nop %s configuration: %w"
	var err error
	n.Server, err = n.Server.DefaultsAndValidate()
	if err != nil {
		return NopConfig{}, fmt.Errorf(errMsg, "server", err)
	}
	n.Storage, err = n.Storage.DefaultsAndValidate()
	if err != nil {
		return NopConfig{}, fmt.Errorf(errMsg, "storage", err)
	}
	n.RateLimiter, err = n.RateLimiter.DefaultsAndValidate()
	if err != nil {
		return NopConfig{}, fmt.Errorf(errMsg, "rate limiter", err)
	}
	return n, nil
}

// MultiConfig is the configuration for a multiple email provider magiclinksdev server.
type MultiConfig struct {
	SES         ses.Config      `json:"ses"`
	SendGrid    sendgrid.Config `json:"sendgrid"`
	Server      config.Config   `json:"server"`
	Storage     storage.Config  `json:"storage"`
	RateLimiter rlimit.Config   `json:"rateLimiter"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (m MultiConfig) DefaultsAndValidate() (MultiConfig, error) {
	const errMsg = "failed to validate and apply defaults to multi-provider %s configuration: %w"
	var err error
	m.SES, err = m.SES.DefaultsAndValidate()
	if err != nil {
		return m, fmt.Errorf(errMsg, "ses", err)
	}
	m.SendGrid, err = m.SendGrid.DefaultsAndValidate()
	if err != nil {
		return m, fmt.Errorf(errMsg, "sendgrid", err)
	}
	m.Server, err = m.Server.DefaultsAndValidate()
	if err != nil {
		return MultiConfig{}, fmt.Errorf(errMsg, "server", err)
	}
	m.Storage, err = m.Storage.DefaultsAndValidate()
	if err != nil {
		return MultiConfig{}, fmt.Errorf(errMsg, "storage", err)
	}
	m.RateLimiter, err = m.RateLimiter.DefaultsAndValidate()
	if err != nil {
		return MultiConfig{}, fmt.Errorf(errMsg, "rate limiter", err)
	}
	return m, nil
}

// SESConfig is the configuration for a single email provider magiclinksdev server.
type SESConfig struct {
	SES         ses.Config     `json:"ses"`
	Server      config.Config  `json:"server"`
	Storage     storage.Config `json:"storage"`
	RateLimiter rlimit.Config  `json:"rateLimiter"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (s SESConfig) DefaultsAndValidate() (SESConfig, error) {
	const errMsg = "failed to validate and apply defaults to multi-provider %s configuration: %w"
	var err error
	s.SES, err = s.SES.DefaultsAndValidate()
	if err != nil {
		return s, fmt.Errorf(errMsg, "ses", err)
	}
	s.Server, err = s.Server.DefaultsAndValidate()
	if err != nil {
		return SESConfig{}, fmt.Errorf(errMsg, "server", err)
	}
	s.Storage, err = s.Storage.DefaultsAndValidate()
	if err != nil {
		return SESConfig{}, fmt.Errorf(errMsg, "storage", err)
	}
	s.RateLimiter, err = s.RateLimiter.DefaultsAndValidate()
	if err != nil {
		return SESConfig{}, fmt.Errorf(errMsg, "rate limiter", err)
	}
	return s, nil
}

// SendGridConfig is the configuration for the SendGrid email provider.
type SendGridConfig struct {
	SendGrid    sendgrid.Config `json:"sendgrid"`
	Server      config.Config   `json:"server"`
	Storage     storage.Config  `json:"storage"`
	RateLimiter rlimit.Config   `json:"rateLimiter"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (m SendGridConfig) DefaultsAndValidate() (SendGridConfig, error) {
	const errMsg = "failed to validate and apply defaults to sendgrid provider %s configuration: %w"
	var err error
	m.SendGrid, err = m.SendGrid.DefaultsAndValidate()
	if err != nil {
		return m, fmt.Errorf(errMsg, "sendgrid", err)
	}
	m.Server, err = m.Server.DefaultsAndValidate()
	if err != nil {
		return SendGridConfig{}, fmt.Errorf(errMsg, "server", err)
	}
	m.Storage, err = m.Storage.DefaultsAndValidate()
	if err != nil {
		return SendGridConfig{}, fmt.Errorf(errMsg, "storage", err)
	}
	m.RateLimiter, err = m.RateLimiter.DefaultsAndValidate()
	if err != nil {
		return SendGridConfig{}, fmt.Errorf(errMsg, "rate limiter", err)
	}
	return m, nil
}

// TestConfig is the configuration for a test magiclinksdev server.
type TestConfig struct {
	Server      config.Config  `json:"server"`
	Storage     storage.Config `json:"storage"`
	RateLimiter rlimit.Config  `json:"rateLimiter"`
}

// DefaultsAndValidate implements the jsontype.Config interface.
func (t TestConfig) DefaultsAndValidate() (TestConfig, error) {
	const errMsg = "failed to validate and apply defaults to sendgrid provider %s configuration: %w"
	var err error
	t.Server, err = t.Server.DefaultsAndValidate()
	if err != nil {
		return TestConfig{}, fmt.Errorf(errMsg, "server", err)
	}
	t.Storage, err = t.Storage.DefaultsAndValidate()
	if err != nil {
		return TestConfig{}, fmt.Errorf(errMsg, "storage", err)
	}
	t.RateLimiter, err = t.RateLimiter.DefaultsAndValidate()
	if err != nil {
		return TestConfig{}, fmt.Errorf(errMsg, "rate limiter", err)
	}
	return t, nil
}

// ServerInterfaces holds all the interface implementations needed for a magiclinksdev server.
type ServerInterfaces struct {
	EmailProvider email.Provider
	RateLimiter   rlimit.RateLimiter
	Store         storage.Storage
}

// ServerOptions holds all the options for a magiclinksdev server.
type ServerOptions struct {
	HTTPMux               *http.ServeMux
	Logger                *slog.Logger
	MagicLinkErrorHandler magiclink.ErrorHandler
	MiddlewareHook        handle.MiddlewareHook
}

// ApplyDefaults applies the default values to the options.
func (o ServerOptions) ApplyDefaults() ServerOptions {
	if o.Logger == nil {
		o.Logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	if o.HTTPMux == nil {
		o.HTTPMux = http.NewServeMux()
	}
	if o.MiddlewareHook == nil {
		o.MiddlewareHook = nopMiddlewareHook{}
	}
	return o
}

// CreateNopProviderServer creates a new magiclinksdev server with a no-operation email provider.
func CreateNopProviderServer(ctx context.Context, conf NopConfig, options ServerOptions) (*handle.Server, error) {
	rateLimiter := rlimit.NewMemory(conf.RateLimiter)
	store, _, err := storage.NewWithSetup(ctx, conf.Storage, options.Logger.With("postgresSetup", true))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	nop := nopProvider{logger: options.Logger.With("provider", "nop")}
	interfaces := ServerInterfaces{
		EmailProvider: nop,
		RateLimiter:   rateLimiter,
		Store:         store,
	}
	return CreateServer(ctx, conf.Server, options, interfaces)
}

// CreateMultiProviderServer creates a new magiclinksdev server with multiple email providers.
func CreateMultiProviderServer(ctx context.Context, conf MultiConfig, options ServerOptions) (*handle.Server, error) {
	sesProvider, err := ses.NewProvider(ctx, conf.SES)
	if err != nil {
		return nil, fmt.Errorf("failed to create email provider: %w", err)
	}
	sendgridProvider, err := sendgrid.NewProvider(conf.SendGrid)
	if err != nil {
		return nil, fmt.Errorf("failed to create email provider: %w", err)
	}
	opts := email.MultiProviderOptions{
		Logger: options.Logger.With("provider", "multi"),
	}
	multiProvider, err := email.NewMultiProvider([]email.Provider{sesProvider, sendgridProvider}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create email provider: %w", err)
	}
	rateLimiter := rlimit.NewMemory(conf.RateLimiter)
	store, _, err := storage.NewWithSetup(ctx, conf.Storage, options.Logger.With("postgresSetup", true))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	interfaces := ServerInterfaces{
		EmailProvider: multiProvider,
		RateLimiter:   rateLimiter,
		Store:         store,
	}
	return CreateServer(ctx, conf.Server, options, interfaces)
}

// CreateSESProvider creates a new magiclinksdev server with a SES email provider.
func CreateSESProvider(ctx context.Context, conf SESConfig, options ServerOptions) (*handle.Server, error) {
	provider, err := ses.NewProvider(ctx, conf.SES)
	if err != nil {
		return nil, fmt.Errorf("failed to create email provider: %w", err)
	}
	rateLimiter := rlimit.NewMemory(conf.RateLimiter)
	store, _, err := storage.NewWithSetup(ctx, conf.Storage, options.Logger.With("postgresSetup", true))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	interfaces := ServerInterfaces{
		EmailProvider: provider,
		RateLimiter:   rateLimiter,
		Store:         store,
	}
	return CreateServer(ctx, conf.Server, options, interfaces)
}

// CreateSendGridProvider creates a new magiclinksdev server with a SendGrid email provider.
func CreateSendGridProvider(ctx context.Context, conf SendGridConfig, options ServerOptions) (*handle.Server, error) {
	provider, err := sendgrid.NewProvider(conf.SendGrid)
	if err != nil {
		return nil, fmt.Errorf("failed to create email provider: %w", err)
	}
	rateLimiter := rlimit.NewMemory(conf.RateLimiter)
	store, _, err := storage.NewWithSetup(ctx, conf.Storage, options.Logger.With("postgresSetup", true))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	interfaces := ServerInterfaces{
		EmailProvider: provider,
		RateLimiter:   rateLimiter,
		Store:         store,
	}
	return CreateServer(ctx, conf.Server, options, interfaces)
}

// CreateTestingProvider creates a new magiclinksdev server with a testing email provider.
func CreateTestingProvider(ctx context.Context, conf TestConfig, options ServerOptions) (*handle.Server, error) {
	provider := mldtest.NopProvider{}
	rateLimiter := rlimit.NewMemory(conf.RateLimiter)
	store, _, err := storage.NewWithSetup(ctx, conf.Storage, options.Logger.With("postgresSetup", true))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}
	interfaces := ServerInterfaces{
		EmailProvider: provider,
		RateLimiter:   rateLimiter,
		Store:         store,
	}
	return CreateServer(ctx, conf.Server, options, interfaces)
}

// CreateServer creates a new magiclinksdev server.
func CreateServer(ctx context.Context, conf config.Config, options ServerOptions, interfaces ServerInterfaces) (*handle.Server, error) {
	options = options.ApplyDefaults()
	logger := options.Logger

	magicLinkServiceURL, err := conf.BaseURL.Get().Parse(conf.RelativeRedirectURL.Get().EscapedPath())
	if err != nil {
		return nil, fmt.Errorf("failed to parse magic link service URL: %w", err)
	}

	var customRedirector magiclink.Redirector
	switch conf.PreventRobots.Method {
	case config.PreventRobotsReCAPTCHAV3:
		customRedirector = magiclink.NewReCAPTCHAV3Redirector(conf.PreventRobots.ReCAPTCHAV3)
	}

	magicLinkConfig := magiclink.Config{
		ErrorHandler: MagicLinkErrorHandler(options.MagicLinkErrorHandler),
		JWKS: magiclink.JWKSParams{
			CacheRefresh: time.Second,
			Store:        interfaces.Store,
		},
		CustomRedirector: customRedirector,
		ServiceURL:       magicLinkServiceURL,
		SecretQueryKey:   conf.SecretQueryKey,
		Store:            interfaces.Store,
	}

	tx, err := interfaces.Store.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	setupCtx := context.WithValue(ctx, ctxkey.Tx, tx)
	setupCtx, setupCancel := context.WithTimeout(setupCtx, time.Second)
	defer setupCancel()

	if !conf.JWKS.IgnoreDefault {
		_, existed, err := CreateKeysIfNotExists(setupCtx, interfaces.Store)
		if err != nil {
			return nil, fmt.Errorf("failed to create key if they didn't already exist: %w", err)
		}
		if existed {
			logger.InfoContext(ctx, "JWK Set keys already exist.")
		} else {
			logger.InfoContext(ctx, "JWK Set keys created.")
		}
	} else {
		logger.InfoContext(ctx, "Ignoring default JWK Set check.")
	}

	for _, adminCreateParams := range conf.AdminCreateParams {
		valid, err := adminCreateParams.Validate(conf.Validation)
		if err != nil {
			return nil, fmt.Errorf("failed to validate admin config: %w", err)
		}
		_, err = interfaces.Store.SARead(setupCtx, valid.UUID)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				err = interfaces.Store.SAAdminCreate(setupCtx, valid)
				if err != nil {
					return nil, fmt.Errorf("failed to setup admin: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to check if admin account exists: %w", err)
			}
		}
		logger.InfoContext(ctx, "Admin account already exists. Skipping creation.",
			"uuid", valid.UUID,
		)
	}

	magicLink, err := magiclink.NewMagicLink(setupCtx, magicLinkConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup magiclink: %w", err)
	}

	err = tx.Commit(setupCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	server := &handle.Server{
		Config:         conf,
		Ctx:            ctx,
		EmailProvider:  interfaces.EmailProvider,
		HTTPMux:        options.HTTPMux,
		JWKS:           magicLink.JWKSet(),
		Limiter:        interfaces.RateLimiter,
		Logger:         logger,
		MagicLink:      magicLink,
		MiddlewareHook: options.MiddlewareHook,
		Store:          interfaces.Store,
	}

	return server, err
}

// MagicLinkErrorHandler is a wrapper for magiclink.ErrorHandlerFunc.
func MagicLinkErrorHandler(h magiclink.ErrorHandler) magiclink.ErrorHandler {
	return magiclink.ErrorHandlerFunc(func(args magiclink.ErrorHandlerParams) {
		ctx := args.Request.Context()
		logger := ctx.Value(ctxkey.Logger).(*slog.Logger)
		logger.ErrorContext(ctx, "Failed to handle magic link.",
			mld.LogErr, args.Err,
		)
		tx := ctx.Value(ctxkey.Tx).(storage.Tx)
		err := tx.Rollback(ctx)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to rollback transaction.",
				mld.LogErr, err,
			)
		}
		if h == nil {
			args.Writer.WriteHeader(args.SuggestedResponseCode)
		} else {
			h.Handle(args)
		}
	})
}

type nopMiddlewareHook struct{}

// Hook implements handle.MiddlewareHook.
func (n nopMiddlewareHook) Hook(options handle.MiddlewareOptions) handle.MiddlewareOptions {
	return options
}

type nopProvider struct {
	logger *slog.Logger
}

func (n nopProvider) SendMagicLink(ctx context.Context, e email.Email) error {
	n.logger.DebugContext(ctx, "Sending magic link email.",
		"email", e,
	)
	return nil
}
func (n nopProvider) SendOTP(ctx context.Context, e email.Email) error {
	n.logger.DebugContext(ctx, "Sending OTP email.",
		"email", e,
	)
	return nil
}

func CreateLogger(srvConf config.Config) *slog.Logger {
	logOpts := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	badLevel := false
	switch srvConf.LogLevel {
	case config.LogLevelDebug:
		logOpts.Level = slog.LevelDebug
	case config.LogLevelInfo:
		logOpts.Level = slog.LevelInfo
	case config.LogLevelWarn:
		logOpts.Level = slog.LevelWarn
	case config.LogLevelError:
		logOpts.Level = slog.LevelError
	default:
		logOpts.Level = slog.LevelDebug
		badLevel = true
	}
	var logger *slog.Logger
	if srvConf.LogJSON {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &logOpts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &logOpts))
	}

	if badLevel {
		logger.Warn("Invalid log level. Using debug.",
			"level", srvConf.LogLevel,
		)
	}

	slog.SetDefault(logger)

	return logger
}

func RunServer(ctx context.Context, logger *slog.Logger, server *handle.Server) {
	mux, err := network.CreateHTTPHandlers(server)
	if err != nil {
		logger.ErrorContext(ctx, "Failed to create HTTP handlers.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", server.Config.Port),
		Handler: mux,
	}

	idleConnsClosed := make(chan struct{})
	go serverShutdown(ctx, server.Config, logger, idleConnsClosed, httpServer)

	logger.InfoContext(ctx, "Server is listening.",
		"port", server.Config.Port,
	)
	err = httpServer.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		logger.ErrorContext(ctx, "Failed to listen and serve.",
			mld.LogErr, err,
		)
		os.Exit(1)
	}

	<-idleConnsClosed
}

func serverShutdown(ctx context.Context, conf config.Config, logger *slog.Logger, idleConnsClosed chan struct{}, srv *http.Server) {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-shutdown:
		logger.InfoContext(ctx, "Got a SIGINT or SIGTERM.")
	case <-ctx.Done():
		logger.InfoContext(ctx, "Context over.",
			mld.LogErr, ctx.Err(),
		)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), conf.ShutdownTimeout.Get())
	defer cancel()
	err := srv.Shutdown(shutdownCtx)
	if err != nil {
		logger.InfoContext(ctx, "Couldn't shut down server before time ended.",
			mld.LogErr, err,
		)
	}

	close(idleConnsClosed)
}
