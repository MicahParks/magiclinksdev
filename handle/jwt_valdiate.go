package handle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
)

var (
	// ErrToken is returned when the JWT is invalid.
	ErrToken = errors.New("JWT invalid")
)

// HandleJWTValidate handles the JWT validation endpoint.
func (s *Server) HandleJWTValidate(ctx context.Context, req model.ValidJWTValidateRequest) (model.JWTValidateResponse, error) {
	jwtValidateArgs := req.JWTValidateArgs
	sa := ctx.Value(ctxkey.ServiceAccount).(model.ServiceAccount)

	token, err := jwt.Parse(jwtValidateArgs.JWT, func(token *jwt.Token) (interface{}, error) {
		jwksBytes, err := s.JWKS.JSONPublic(ctx) // Change to JSONPrivate if HMAC support is added.
		if err != nil {
			return nil, fmt.Errorf("failed to get JWKS JSON: %w", err)
		}
		jwks, err := keyfunc.NewJSON(jwksBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to create keyfunc from JSON: %w", err)
		}
		return jwks.Keyfunc(token)
	})
	if err != nil {
		var errJWTValid *jwt.ValidationError
		if errors.Is(err, keyfunc.ErrKIDNotFound) || errors.As(err, &errJWTValid) {
			return model.JWTValidateResponse{}, fmt.Errorf("%w: %s", ErrToken, err)
		}
		return model.JWTValidateResponse{}, fmt.Errorf("failed to parse JWT: %w", err)
	}
	if !token.Valid {
		return model.JWTValidateResponse{}, fmt.Errorf("%w: token invalid", ErrToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return model.JWTValidateResponse{}, fmt.Errorf("%w: claims invalid", ErrToken)
	}
	ok = claims.VerifyIssuer(s.Config.Iss, true)
	if !ok {
		return model.JWTValidateResponse{}, fmt.Errorf("%w: issuer invalid", ErrToken)
	}
	ok = claims.VerifyAudience(sa.Aud.String(), true)
	if !ok {
		return model.JWTValidateResponse{}, fmt.Errorf("%w: incorrect audience for this service account", ErrToken)
	}

	raw, err := json.Marshal(claims)
	if err != nil {
		return model.JWTValidateResponse{}, fmt.Errorf("failed to marshal claims: %w", err)
	}

	response := model.JWTValidateResponse{
		JWTValidateResults: model.JWTValidateResults{
			JWTClaims: raw,
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return response, nil
}
