package email

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	mld "github.com/MicahParks/magiclinksdev"
)

// MultiProviderOptions are the options for the multi-provider.
type MultiProviderOptions struct {
	Logger *slog.Logger
}

type multiProvider struct {
	logger    *slog.Logger
	providers []Provider
}

// NewMultiProvider creates a new multiple email provider.
func NewMultiProvider(providers []Provider, options MultiProviderOptions) (Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("%w: no providers given in multi-provider creation", ErrProvider)
	}
	if options.Logger == nil {
		options.Logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	m := multiProvider{
		providers: providers,
		logger:    options.Logger,
	}
	return m, nil
}

func (m multiProvider) SendMagicLink(ctx context.Context, e Email) error {
	combinedErr := fmt.Errorf("%w: no providers were able to send the email", ErrProvider)
	for _, p := range m.providers {
		err := p.SendMagicLink(ctx, e)
		if err == nil {
			return nil
		}
		m.logger.ErrorContext(ctx, "Failed to send email with using multi-provider. Attempting with next provider.",
			mld.LogErr, err,
		)
		combinedErr = fmt.Errorf("%w: %w", combinedErr, err)
	}
	m.logger.ErrorContext(ctx, "Failed to send email with using multi-provider. No providers were able to send the email.",
		mld.LogErr, combinedErr,
	)
	return combinedErr
}
func (m multiProvider) SendOTP(ctx context.Context, e Email) error {
	combinedErr := fmt.Errorf("%w: no providers were able to send the email", ErrProvider)
	for _, p := range m.providers {
		err := p.SendOTP(ctx, e)
		if err == nil {
			return nil
		}
		m.logger.ErrorContext(ctx, "Failed to send email with using multi-provider. Attempting with next provider.",
			mld.LogErr, err,
		)
		combinedErr = fmt.Errorf("%w: %w", combinedErr, err)
	}
	m.logger.ErrorContext(ctx, "Failed to send email with using multi-provider. No providers were able to send the email.",
		mld.LogErr, combinedErr,
	)
	return combinedErr
}
