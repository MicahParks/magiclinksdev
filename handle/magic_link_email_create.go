package handle

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/config"
	"github.com/MicahParks/magiclinksdev/email"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
)

// HandleMagicLinkEmailCreate handles the email link creation endpoint.
func (s *Server) HandleMagicLinkEmailCreate(ctx context.Context, req model.ValidMagicLinkEmailCreateRequest) (model.MagicLinkEmailCreateResponse, error) {
	emailParams := req.MagicLinkEmailCreateParams
	linkParams := req.MagicLinkCreateParams

	magicLinkRes, err := s.createLink(ctx, linkParams)
	if err != nil {
		return model.MagicLinkEmailCreateResponse{}, fmt.Errorf("failed to create magic link: %w", err)
	}

	meta := email.TemplateMetadata{
		HTMLInstruction: fmt.Sprintf("You've been sent a magic link from %s.", emailParams.ServiceName),
		HTMLTitle:       fmt.Sprintf("Magic link from %s", emailParams.ServiceName),
		MSOButtonStop:   email.MSOButtonStop,
		MSOButtonStart:  email.MSOButtonStart,
		MSOHead:         email.MSOHead,
	}
	tData := email.MagicLinkTemplateData{
		ButtonText:   emailParams.ButtonText,
		Expiration:   linkParams.LinkLifespan.String(),
		Greeting:     emailParams.Greeting,
		LogoImageURL: emailParams.LogoImageURL,
		LogoClickURL: emailParams.LogoClickURL,
		LogoAltText:  "logo",
		MagicLink:    magicLinkRes.MagicLink.String(),
		Meta:         meta,
		Subtitle:     emailParams.SubTitle,
		Title:        emailParams.Title,
		ReCATPTCHA:   s.Config.PreventRobots.Method == config.PreventRobotsReCAPTCHAV3,
	}
	e := email.Email{
		Subject:      emailParams.Subject,
		TemplateData: tData,
		To:           emailParams.ToEmail,
	}
	err = s.EmailProvider.Send(ctx, e)
	if err != nil {
		return model.MagicLinkEmailCreateResponse{}, fmt.Errorf("failed to send email: %w", err)
	}

	linkCreateResponse := model.MagicLinkCreateResults{
		MagicLink: magicLinkRes.MagicLink.String(),
		Secret:    magicLinkRes.Secret,
	}
	resp := model.MagicLinkEmailCreateResponse{
		MagicLinkEmailCreateResults: model.MagicLinkEmailCreateResults{
			MagicLinkCreateResults: linkCreateResponse,
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return resp, nil
}
