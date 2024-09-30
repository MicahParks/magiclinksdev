package magiclinksdev

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tidwall/gjson"
)

const (
	AttrIss = "iss"
	AttrSub = "sub"
	AttrAud = "aud"
	AttrExp = "exp"
	AttrNbf = "nbf"
	AttrIat = "iat"
	AttrJti = "jti"
)

// SigningBytesClaims is a JWT claims type that allows for signing claims represented in bytes.
type SigningBytesClaims struct {
	Claims json.RawMessage
}

func (s SigningBytesClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(gjson.GetBytes(s.Claims, AttrExp).Int(), 0)), nil
}
func (s SigningBytesClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(gjson.GetBytes(s.Claims, AttrIat).Int(), 0)), nil
}
func (s SigningBytesClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(gjson.GetBytes(s.Claims, AttrNbf).Int(), 0)), nil
}
func (s SigningBytesClaims) GetIssuer() (string, error) {
	return gjson.GetBytes(s.Claims, AttrIss).String(), nil
}
func (s SigningBytesClaims) GetSubject() (string, error) {
	return gjson.GetBytes(s.Claims, AttrSub).String(), nil
}
func (s SigningBytesClaims) GetAudience() (jwt.ClaimStrings, error) {
	var aud jwt.ClaimStrings
	err := json.Unmarshal([]byte(gjson.GetBytes(s.Claims, AttrAud).Raw), &aud)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal audience: %w", err)
	}
	return aud, nil
}

// MarshalJSON helps implement the json.Marshaler interface.
func (s SigningBytesClaims) MarshalJSON() ([]byte, error) {
	return s.Claims, nil
}
