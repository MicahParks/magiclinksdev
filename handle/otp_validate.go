package handle

import (
	"context"
	"fmt"

	"github.com/MicahParks/magiclinksdev/model"
)

func (s *Server) HandleOTPValidate(ctx context.Context, req model.ValidOTPValidateRequest) (response model.OTPValidateResponse, err error) {
	err = s.Store.OTPValidate(ctx, req.OTPValidateParams.ID, req.OTPValidateParams.OTP)
	if err != nil {
		return model.OTPValidateResponse{}, fmt.Errorf("failed to validate OTP: %w", err)
	}
	resp := model.OTPValidateResponse{}
	return resp, nil
}
