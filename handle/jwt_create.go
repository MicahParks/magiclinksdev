package handle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/MicahParks/magiclinksdev/magiclink"
	"github.com/MicahParks/magiclinksdev/model"
	"github.com/MicahParks/magiclinksdev/network/middleware/ctxkey"
	"github.com/MicahParks/magiclinksdev/storage"
)

const (
	issAttr = "iss"
	subAttr = "sub"
	audAttr = "aud"
	expAttr = "exp"
	nbfAttr = "nbf"
	iatAttr = "iat"
	jtiAttr = "jti"
)

var (
	// ErrRegisteredClaimProvided is returned when a registered claim is provided.
	ErrRegisteredClaimProvided = errors.New("registered claims should not be provided")
	// ErrJWTAlgNotFound is returned when a JWT alg is not found.
	ErrJWTAlgNotFound = errors.New("JWT alg not found")
)

// HandleJWTCreate handles the creation of a JWT.
func (s *Server) HandleJWTCreate(ctx context.Context, req model.ValidJWTCreateRequest) (model.JWTCreateResponse, error) {
	jwtCreateArgs := req.JWTCreateArgs

	edited, err := s.addRegisteredClaims(ctx, jwtCreateArgs)
	if err != nil {
		return model.JWTCreateResponse{}, fmt.Errorf("failed to add registered claims to JWT claims: %w", err)
	}

	options := storage.ReadSigningKeyOptions{
		JWTAlg: jwtCreateArgs.JWTAlg,
	}
	jwk, err := s.Store.ReadSigningKey(ctx, options)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return model.JWTCreateResponse{}, fmt.Errorf("could not fing signing key with specified JWT alg: %w", ErrJWTAlgNotFound)
		}
		return model.JWTCreateResponse{}, fmt.Errorf("failed to get JWT signing key: %w", err)
	}
	method := magiclink.BestSigningMethod(jwk.Key)

	bytesClaims := SigningBytesClaims{
		Claims: edited,
	}
	token := jwt.NewWithClaims(method, bytesClaims)
	token.Header[jwkset.HeaderKID] = jwk.Marshal().KID
	signed, err := token.SignedString(jwk.Key)
	if err != nil {
		return model.JWTCreateResponse{}, fmt.Errorf("%w: %w", magiclink.ErrJWTSign, err)
	}

	response := model.JWTCreateResponse{
		JWTCreateResults: model.JWTCreateResults{
			JWT: signed,
		},
		RequestMetadata: model.RequestMetadata{
			UUID: ctx.Value(ctxkey.RequestUUID).(uuid.UUID),
		},
	}

	return response, nil
}

// SigningBytesClaims is a JWT claims type that allows for signing claims represented in bytes.
type SigningBytesClaims struct {
	Claims json.RawMessage
}

func (s SigningBytesClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(gjson.GetBytes(s.Claims, expAttr).Int(), 0)), nil
}
func (s SigningBytesClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(gjson.GetBytes(s.Claims, iatAttr).Int(), 0)), nil
}
func (s SigningBytesClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(gjson.GetBytes(s.Claims, nbfAttr).Int(), 0)), nil
}
func (s SigningBytesClaims) GetIssuer() (string, error) {
	return gjson.GetBytes(s.Claims, issAttr).String(), nil
}
func (s SigningBytesClaims) GetSubject() (string, error) {
	return gjson.GetBytes(s.Claims, subAttr).String(), nil
}
func (s SigningBytesClaims) GetAudience() (jwt.ClaimStrings, error) {
	var aud jwt.ClaimStrings
	err := json.Unmarshal([]byte(gjson.GetBytes(s.Claims, audAttr).Raw), &aud)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal audience: %w", err)
	}
	return aud, nil
}

// Valid helps implement the jwt.Claims interface.
func (s SigningBytesClaims) Valid() error {
	valid := json.Valid(s.Claims)
	if !valid {
		return fmt.Errorf("%w: invalid JSON for JWT claims", model.ErrInvalidModel)
	}
	return nil
}

// MarshalJSON helps implement the json.Marshaler interface.
func (s SigningBytesClaims) MarshalJSON() ([]byte, error) {
	return s.Claims, nil
}

func (s *Server) addRegisteredClaims(ctx context.Context, args model.ValidJWTCreateArgs) (json.RawMessage, error) {
	sa, ok := ctx.Value(ctxkey.ServiceAccount).(model.ServiceAccount)
	if !ok {
		return nil, fmt.Errorf("%w: service account context not found", ctxkey.ErrCtxKey)
	}

	valid := json.Valid(args.JWTClaims)
	if !valid {
		return nil, fmt.Errorf("%w: invalid JSON for JWT claims", model.ErrInvalidModel)
	}

	n := time.Now()
	now := jwt.NewNumericDate(n)
	u, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}
	registeredClaims := jwt.RegisteredClaims{
		Issuer:    s.Config.Iss,
		Subject:   "", // Don't set.
		Audience:  jwt.ClaimStrings{sa.Aud.String()},
		ExpiresAt: jwt.NewNumericDate(n.Add(args.JWTLifespan)),
		NotBefore: now,
		IssuedAt:  now,
		ID:        u.String(),
	}

	registeredMarshaled, err := json.Marshal(registeredClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registered claims: %w", err)
	}

	// https://tools.ietf.org/html/rfc7519#section-4.1
	rfc5119 := []string{
		issAttr,
		subAttr,
		audAttr,
		expAttr,
		nbfAttr,
		iatAttr,
		jtiAttr,
	}

	edited := make(json.RawMessage, len(args.JWTClaims))
	copy(edited, args.JWTClaims)
	for _, attr := range rfc5119 {
		if gjson.GetBytes(edited, attr).Exists() {
			return nil, fmt.Errorf("%w: %s", ErrRegisteredClaimProvided, attr)
		}
		registered := gjson.GetBytes(registeredMarshaled, attr)
		if registered.Exists() {
			edited, err = sjson.SetBytes(edited, attr, registered.Value())
			if err != nil {
				return nil, fmt.Errorf("failed to set registered claims in JWT claims: %w", err)
			}
		} else {
			edited, err = sjson.DeleteBytes(edited, attr)
			if err != nil {
				return nil, fmt.Errorf("failed to delete %s from registered claims: %w", attr, err)
			}
		}
	}

	return edited, nil
}

func (s *Server) createLinkArgs(ctx context.Context, args model.ValidLinkCreateArgs) (magiclink.CreateArgs[storage.MagicLinkCustomCreateArgs], error) {
	var createArgs magiclink.CreateArgs[storage.MagicLinkCustomCreateArgs]

	edited, err := s.addRegisteredClaims(ctx, args.JWTCreateArgs)
	if err != nil {
		return createArgs, fmt.Errorf("failed to add registered claims to JWT claims: %w", err)
	}

	claims := SigningBytesClaims{
		Claims: edited,
	}

	options := storage.ReadSigningKeyOptions{
		JWTAlg: args.JWTCreateArgs.JWTAlg,
	}
	jwk, err := s.Store.ReadSigningKey(ctx, options)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return createArgs, fmt.Errorf("could not fing signing key with specified JWT alg: %w", ErrJWTAlgNotFound)
		}
		return createArgs, fmt.Errorf("failed to get JWT signing key: %w", err)
	}

	kID := jwk.Marshal().KID
	createArgs = magiclink.CreateArgs[storage.MagicLinkCustomCreateArgs]{
		Custom: storage.MagicLinkCustomCreateArgs{
			Expires: time.Now().Add(args.LinkLifespan),
		},
		JWTClaims:        claims,
		JWTKeyID:         &kID,
		RedirectQueryKey: args.RedirectQueryKey,
		RedirectURL:      args.RedirectURL,
	}

	return createArgs, nil
}
