package handle

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"

	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
)

// HandleLinkCreate handles the link creation endpoint.
func (s *Server) HandleLinkCreate(ctx context.Context, req model.ValidLinkCreateRequest) (response model.LinkCreateResponse, err error) {
	linkArgs := req.LinkArgs

	magicLinkResp, err := s.createLink(ctx, linkArgs)
	if err != nil {
		return model.LinkCreateResponse{}, fmt.Errorf("failed to create magic link: %w", err)
	}

	resp := model.LinkCreateResponse{
		LinkCreateResults: model.LinkCreateResults{
			MagicLink: magicLinkResp.MagicLink.String(),
			Secret:    magicLinkResp.Secret,
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return resp, nil
}

func (s *Server) createLink(ctx context.Context, linkArgs model.ValidLinkCreateArgs) (magiclink.CreateResponse, error) {
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
