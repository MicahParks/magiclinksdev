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
	emailArgs := req.MagicLinkEmailCreateArgs
	linkArgs := req.MagicLinkCreateArgs

	magicLinkResp, err := s.createLink(ctx, linkArgs)
	if err != nil {
		return model.MagicLinkEmailCreateResponse{}, fmt.Errorf("failed to create magic link: %w", err)
	}

	meta := email.TemplateMetadata{
		HTMLInstruction: fmt.Sprintf("You've been sent a magic link from %s.", emailArgs.ServiceName),
		HTMLTitle:       fmt.Sprintf("Magic link from %s", emailArgs.ServiceName),
		MSOButtonStop:   email.MSOButtonStop,
		MSOButtonStart:  email.MSOButtonStart,
		MSOHead:         email.MSOHead,
	}
	tData := email.TemplateData{
		ButtonText:   emailArgs.ButtonText,
		Expiration:   linkArgs.LinkLifespan.String(),
		Greeting:     emailArgs.Greeting,
		LogoImageURL: emailArgs.LogoImageURL,
		LogoClickURL: emailArgs.LogoClickURL,
		LogoAltText:  "logo",
		MagicLink:    magicLinkResp.MagicLink.String(),
		Meta:         meta,
		Subtitle:     emailArgs.SubTitle,
		Title:        emailArgs.Title,
		ReCATPTCHA:   s.Config.PreventRobots.Method == config.PreventRobotsReCAPTCHAV3,
	}
	e := email.Email{
		Subject:      emailArgs.Subject,
		TemplateData: tData,
		To:           emailArgs.ToEmail,
	}
	err = s.EmailProvider.Send(ctx, e)
	if err != nil {
		return model.MagicLinkEmailCreateResponse{}, fmt.Errorf("failed to send email: %w", err)
	}

	linkCreateResponse := model.MagicLinkCreateResults{
		MagicLink: magicLinkResp.MagicLink.String(),
		Secret:    magicLinkResp.Secret,
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
