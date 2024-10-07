package handle

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"

	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
)

// HandleMagicLinkCreate handles the link creation endpoint.
func (s *Server) HandleMagicLinkCreate(ctx context.Context, req model.ValidMagicLinkCreateRequest) (response model.MagicLinkCreateResponse, err error) {
	linkArgs := req.MagicLinkArgs

	magicLinkResp, err := s.createLink(ctx, linkArgs)
	if err != nil {
		return model.MagicLinkCreateResponse{}, fmt.Errorf("failed to create magic link: %w", err)
	}

	resp := model.MagicLinkCreateResponse{
		MagicLinkCreateResults: model.MagicLinkCreateResults{
			MagicLink: magicLinkResp.MagicLink.String(),
			Secret:    magicLinkResp.Secret,
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return resp, nil
}

func (s *Server) createLink(ctx context.Context, linkArgs model.ValidMagicLinkCreateArgs) (magiclink.CreateResponse, error) {
	magicLinkCreateArgs, err := s.createLinkArgs(ctx, linkArgs)
	if err != nil {
		return magiclink.CreateResponse{}, fmt.Errorf("failed to create magic link create args: %w", err)
	}

	magicLinkResp, err := s.MagicLink.NewLink(ctx, magicLinkCreateArgs)
	if err != nil {
		return magiclink.CreateResponse{}, fmt.Errorf("failed to create magic link: %w", err)
	}

	return magicLinkResp, nil
}
