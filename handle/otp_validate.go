package handle

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
)

func (s *Server) HandleOTPValidate(ctx context.Context, req model.ValidOTPValidateRequest) (response model.OTPValidateResponse, err error) {
	err = s.Store.OTPValidate(ctx, req.OTPValidateParams.ID, req.OTPValidateParams.OTP)
	if err != nil {
		return model.OTPValidateResponse{}, fmt.Errorf("failed to validate OTP: %w", err)
	}
	resp := model.OTPValidateResponse{
		OTPValidateResults: model.OTPValidateResults{},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}
	return resp, nil
}
