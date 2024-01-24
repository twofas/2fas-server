package sign

import (
	"errors"
	"fmt"
	"slices"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidClaims = errors.New("invalid claims")

type customClaims struct {
	ConnectionID string `json:"c_id"`
	jwt.RegisteredClaims
}

// CanI establish connection with type tp given claims in token.
// Returns extension_id from claims if token is valid for given type.
func (s Service) CanI(tokenString string, ct ConnectionType) (string, error) {
	cl := customClaims{}

	// In Sign we removed `jwtHeader` from JWT before returning it.
	// We need to add it again before doing the verification.
	tokenString = jwtHeader + tokenString

	token, err := jwt.ParseWithClaims(
		tokenString,
		&cl,
		func(token *jwt.Token) (interface{}, error) {
			return s.publicKey, nil
		},
		jwt.WithValidMethods([]string{"ES256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	audClaims, err := token.Claims.GetAudience()
	if err != nil {
		return "", fmt.Errorf("failed to get claims: %w", err)
	}
	if !slices.Contains(audClaims, string(ct)) {
		return "", fmt.Errorf("%w: claim %q not found in claims", ErrInvalidClaims, ct)
	}
	if cl.ConnectionID == "" {
		return "", fmt.Errorf("%w: claim %q not found in claims", ErrInvalidClaims, "c_id")
	}
	// TODO: rename connectionID to extensionID.
	return cl.ConnectionID, nil
}
