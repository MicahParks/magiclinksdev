package handle

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
)

// HandleServiceAccountCreate handles the service account creation endpoint.
func (s *Server) HandleServiceAccountCreate(ctx context.Context, args model.ValidServiceAccountCreateRequest) (model.ServiceAccountCreateResponse, error) {
	saArgs := args.ServiceAccountCreateArgs

	createdSA, err := s.Store.CreateSA(ctx, saArgs)
	if err != nil {
		return model.ServiceAccountCreateResponse{}, fmt.Errorf("failed to create service account: %w", err)
	}

	serviceAccount, err := s.Store.ReadSA(ctx, createdSA.UUID)
	if err != nil {
		return model.ServiceAccountCreateResponse{}, fmt.Errorf("failed to get service account as marshallable data structure: %w", err)
	}

	resp := model.ServiceAccountCreateResponse{
		ServiceAccountCreateResults: model.ServiceAccountCreateResults{
			ServiceAccount: serviceAccount,
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return resp, nil
}
