package handle

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/email"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
)

func (s *Server) HandleOTPEmailCreate(ctx context.Context, req model.ValidOTPEmailCreateRequest) (response model.OTPEmailCreateResponse, err error) {
	emailParams := req.OTPEmailCreateParams
	otpParams := createOTPParams(req.OTPCreateParams)

	otpRes, err := s.Store.OTPCreate(ctx, otpParams)
	if err != nil {
		return model.OTPEmailCreateResponse{}, fmt.Errorf("failed to create OTP: %w", err)
	}

	meta := email.TemplateMetadata{
		HTMLInstruction: fmt.Sprintf("One-Time Password from %s.", emailParams.ServiceName),
		HTMLTitle:       fmt.Sprintf("One-Time Password from %s", emailParams.ServiceName),
		MSOButtonStop:   email.MSOButtonStop,
		MSOButtonStart:  email.MSOButtonStart,
		MSOHead:         email.MSOHead,
	}
	tData := email.OTPTemplateData{
		Expiration:   req.OTPCreateParams.Lifespan.String(),
		Greeting:     emailParams.Greeting,
		Meta:         meta,
		OTP:          otpRes.OTP,
		Subtitle:     emailParams.SubTitle,
		Title:        emailParams.Title,
		LogoImageURL: emailParams.LogoImageURL,
		LogoClickURL: emailParams.LogoClickURL,
		LogoAltText:  "logo",
	}
	e := email.Email{
		Subject:      emailParams.Subject,
		TemplateData: tData,
		To:           emailParams.ToEmail,
	}
	err = s.EmailProvider.SendOTP(ctx, e)
	if err != nil {
		return model.OTPEmailCreateResponse{}, fmt.Errorf("failed to send email: %w", err)
	}

	resp := model.OTPEmailCreateResponse{
		OTPEmailCreateResults: model.OTPEmailCreateResults{
			OTPCreateResults: model.OTPCreateResults{
				ID:  otpRes.ID,
				OTP: otpRes.OTP,
			},
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return resp, nil
}
