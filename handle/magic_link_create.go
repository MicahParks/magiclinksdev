package handle

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"

	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
)

func (s *Server) HandleMagicLinkCreate(ctx context.Context, req model.ValidMagicLinkCreateRequest) (response model.MagicLinkCreateResponse, err error) {
	linkParams := req.MagicLinkParams

	magicLinkRes, err := s.createLink(ctx, linkParams)
	if err != nil {
		return model.MagicLinkCreateResponse{}, fmt.Errorf("failed to create magic link: %w", err)
	}

	resp := model.MagicLinkCreateResponse{
		MagicLinkCreateResults: model.MagicLinkCreateResults{
			MagicLink: magicLinkRes.MagicLink.String(),
			Secret:    magicLinkRes.Secret,
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return resp, nil
}

func (s *Server) createLink(ctx context.Context, linkParams model.ValidMagicLinkCreateParams) (magiclink.CreateResponse, error) {
	magicLinkCreateParams, err := s.createLinkParams(ctx, linkParams)
	if err != nil {
		return magiclink.CreateResponse{}, fmt.Errorf("failed to create magic link create args: %w", err)
	}

	magicLinkRes, err := s.MagicLink.NewLink(ctx, magicLinkCreateParams)
	if err != nil {
		return magiclink.CreateResponse{}, fmt.Errorf("failed to create magic link: %w", err)
	}

	return magicLinkRes, nil
}
