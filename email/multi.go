package email

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	mld "github.com/MicahParks/magiclinksdev"
)

// MultiProviderOptions are the options for the multi-provider.
type MultiProviderOptions struct {
	Sugared *zap.SugaredLogger
}

type multiProvider struct {
	providers []Provider
	sugared   *zap.SugaredLogger
}

// NewMultiProvider creates a new multiple email provider.
func NewMultiProvider(providers []Provider, options MultiProviderOptions) (Provider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("%w: no providers given in multi-provider creation", ErrProvider)
	}
	if options.Sugared == nil {
		options.Sugared = zap.NewNop().Sugar()
	}
	m := multiProvider{
		providers: providers,
		sugared:   options.Sugared,
	}
	return m, nil
}

// Send sends an email using the multiple email provider.
func (m multiProvider) Send(ctx context.Context, e Email) error {
	combinedErr := fmt.Errorf("%w: no providers were able to send the email", ErrProvider)
	for _, p := range m.providers {
		err := p.Send(ctx, e)
		if err == nil {
			return nil
		}
		m.sugared.Errorw("Failed to send email with using multi-provider. Attempting with next provider.",
			mld.LogErr, err,
		)
		combinedErr = fmt.Errorf("%w: %w", combinedErr, err)
	}
	m.sugared.Errorw("Failed to send email with using multi-provider. No providers were able to send the email.",
		mld.LogErr, combinedErr,
	)
	return combinedErr
}
