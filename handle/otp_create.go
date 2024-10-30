package handle

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/otp"
)

func (s *Server) HandleOTPCreate(ctx context.Context, req model.ValidOTPCreateRequest) (response model.OTPCreateResponse, err error) {
	otpParams := createOTPParams(req.OTPCreateParams)

	otpRes, err := s.Store.OTPCreate(ctx, otpParams)
	if err != nil {
		return model.OTPCreateResponse{}, fmt.Errorf("failed to create OTP: %w", err)
	}

	resp := model.OTPCreateResponse{
		OTPCreateResults: model.OTPCreateResults{
			ID:  otpRes.ID,
			OTP: otpRes.OTP,
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return resp, nil
}

func createOTPParams(otpParams model.ValidOTPCreateParams) otp.CreateParams {
	params := otp.CreateParams{
		CharSetAlphaLower: otpParams.CharSetAlphaLower,
		CharSetAlphaUpper: otpParams.CharSetAlphaUpper,
		CharSetNumeric:    otpParams.CharSetNumeric,
		Expires:           time.Now().Add(otpParams.Lifespan),
		Length:            otpParams.Length,
	}
	return params
}
