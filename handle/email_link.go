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

// HandleEmailLinkCreate handles the email link creation endpoint.
func (s *Server) HandleEmailLinkCreate(ctx context.Context, req model.ValidEmailLinkCreateRequest) (model.EmailLinkCreateResponse, error) {
	emailArgs := req.EmailArgs
	linkArgs := req.LinkArgs

	magicLinkResp, err := s.createLink(ctx, linkArgs)
	if err != nil {
		return model.EmailLinkCreateResponse{}, fmt.Errorf("failed to create magic link: %w", err)
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
		return model.EmailLinkCreateResponse{}, fmt.Errorf("failed to send email: %w", err)
	}

	linkCreateResponse := model.LinkCreateResults{
		MagicLink: magicLinkResp.MagicLink.String(),
		Secret:    magicLinkResp.Secret,
	}
	resp := model.EmailLinkCreateResponse{
		EmailLinkCreateResults: model.EmailLinkCreateResults{
			LinkCreateResults: linkCreateResponse,
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return resp, nil
}
